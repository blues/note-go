package notehub

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// maxResponseBytes bounds how much of an OAuth/OIDC HTTP response body we
	// read, so a hostile or misconfigured endpoint cannot exhaust memory. It
	// is far larger than any legitimate token or userinfo response.
	maxResponseBytes = 1 << 20 // 1 MiB

	// maxDetailBytes bounds how much of an (untrusted) value is embedded into
	// a log line or returned error.
	maxDetailBytes = 1024
)

// readResponseBody reads up to maxResponseBytes from an HTTP response body.
func readResponseBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
}

// safeDetail renders an untrusted value for inclusion in logs and returned
// errors. It truncates to maxDetailBytes and quotes the result so control
// characters (such as terminal escape sequences) are shown as printable
// escapes rather than interpreted by a terminal.
func safeDetail(s string) string {
	suffix := ""
	if len(s) > maxDetailBytes {
		s = s[:maxDetailBytes]
		suffix = " (truncated)"
	}
	return strconv.Quote(s) + suffix
}

// sensitiveResponseFields are response fields that may carry credentials and
// must never be written to logs or returned errors.
var sensitiveResponseFields = []string{"access_token", "refresh_token", "id_token"}

// redactSensitive returns a representation of an HTTP response body that is
// safe to log: when the body is a JSON object, values of known credential-
// bearing fields are replaced with a placeholder; otherwise the body is
// returned unchanged (token endpoints return credentials as JSON, so a
// non-JSON body has no field to target).
func redactSensitive(body []byte) string {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return string(body)
	}
	redacted := false
	for _, field := range sensitiveResponseFields {
		if _, ok := obj[field]; ok {
			obj[field] = "[REDACTED]"
			redacted = true
		}
	}
	if !redacted {
		return string(body)
	}
	if out, err := json.Marshal(obj); err == nil {
		return string(out)
	}
	return string(body)
}

// callbackAction classifies an inbound request to the local OAuth callback
// server.
type callbackAction int

const (
	// callbackIgnore: not the OAuth redirect (e.g. favicon, prefetch, a bare
	// visit to the root). Must be answered benignly without affecting the flow.
	callbackIgnore callbackAction = iota
	// callbackStateMismatch: carries an authorization code but the wrong state
	// (a CSRF attempt or a stale redirect). Must fail closed.
	callbackStateMismatch
	// callbackProceed: a well-formed redirect whose state matches.
	callbackProceed
)

// classifyCallback decides how to treat a request to the callback server. A
// request without an authorization code is not the OAuth redirect at all and
// is ignored; one with a code is honored only if its state matches the value
// the flow generated.
func classifyCallback(r *http.Request, expectedState string) callbackAction {
	if r.URL.Query().Get("code") == "" {
		return callbackIgnore
	}
	if r.URL.Query().Get("state") != expectedState {
		return callbackStateMismatch
	}
	return callbackProceed
}

type AccessToken struct {
	Host        string
	Email       string
	AccessToken string
	ExpiresAt   time.Time
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// listenOnAny tries each port in order and returns a bound net.Listener for the first available one.
func listenOnAny(ports []int) (net.Listener, int, error) {
	for _, p := range ports {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, errors.New("no ports available")
}

func randString(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RevokeAccessToken(hub, token string) error {
	form := url.Values{
		"token":           {token},
		"token_type_hint": {"access_token"},
		"client_id":       {"notehub_cli"},
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("https://%s/oauth2/revoke", hub),
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	// Per RFC 7009: 200 OK is returned even if the token is already revoked
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// InitiateBrowserBasedLogin starts the OAuth2 login flow by opening the user's browser.
// the `hub` parameter is the hostname of Notehub where it is assumed that an OAuth2 client
// with client ID `notehub_cli` is configured for authorization code flow.
func InitiateBrowserBasedLogin(notehubApiHost string) (*AccessToken, error) {
	// this is the hard-coded OAuth client ID that's persisted in Hydra
	clientId := "notehub_cli"

	if !strings.HasPrefix(notehubApiHost, "api.") {
		notehubApiHost = "api." + notehubApiHost
	}

	var notehubUiHost string
	if notehubApiHost == "api.notefile.net" {
		notehubUiHost = "notehub.io"
	} else {
		notehubUiHost = strings.TrimPrefix(notehubApiHost, "api.")
	}

	// Try these ports in order until one is available:
	//
	// these ports are randomly chosen and hard-coded into
	// the OAuth client in Hydra within Notehub (in the redirect_uris field)
	ports := []int{58766, 58767, 58768, 58769, 42100, 42101, 42102, 42103}

	// Return values
	var accessToken *AccessToken
	var accessTokenErr error

	state := randString(16)
	codeVerifier := randString(50) // must be at least 43 characters
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	// Ensures exactly one OAuth callback is processed; spurious or duplicate
	// requests to the callback server are answered benignly without touching
	// the shared result or triggering shutdown.
	var once sync.Once

	router := http.NewServeMux()

	// We'll fill this after we pick a port but declare it now so the handler can close over it.
	chosenPort := 0

	// The browser will be redirected to this endpoint with an authorization code
	// and then this endpoint will exchange that authorization code for an access token
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		action := classifyCallback(r, state)

		// A request that isn't the OAuth redirect (favicon, prefetch, a bare
		// visit to the root) must not be mistaken for a failed sign-in nor tear
		// the flow down before the real redirect arrives.
		if action == callbackIgnore {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Handle exactly one callback; answer any duplicate benignly so a
		// second request cannot race the shared result or signal shutdown twice.
		handled := false
		once.Do(func() { handled = true })
		if !handled {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "authentication already completed; you may close this window")
			return
		}

		authorizationCode := r.URL.Query().Get("code")

		// fail records an authentication failure exactly one way for every
		// error path in this handler. The browser callback only ever receives
		// `summary`, which is composed solely of developer-authored text and
		// non-injectable scalars such as the numeric HTTP status code -- never
		// free-form server-controlled bytes -- and is always written as
		// text/plain, so the callback cannot be used to reflect injected
		// markup or scripts. `detail` (already passed through safeDetail by
		// the caller) carries diagnostics to the local log and the returned
		// Go error only.
		fail := func(summary, detail string) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error: %s", summary)

			msg := summary
			if detail != "" {
				msg = summary + ": " + detail
			}
			fmt.Printf("error: %s\n", msg)
			accessTokenErr = errors.New(msg)

			// Signal the server to shut down so InitiateBrowserBasedLogin does
			// not block on <-done waiting for a success that will never come.
			// Non-blocking: the buffered channel may already hold a signal
			// (e.g. an OS interrupt or a prior callback).
			select {
			case quit <- os.Interrupt:
			default:
			}
		}

		if action == callbackStateMismatch {
			fail("state mismatch", "")
			return
		}

		///////////////////////////////////////////
		// Exchange code for access token
		///////////////////////////////////////////

		tokenResp, err := http.Post(
			(&url.URL{
				Scheme: "https",
				Host:   notehubUiHost,
				Path:   "/oauth2/token",
			}).String(),
			"application/x-www-form-urlencoded",
			strings.NewReader(url.Values{
				"client_id":     {clientId},
				"code":          {authorizationCode},
				"code_verifier": {codeVerifier},
				"grant_type":    {"authorization_code"},
				"redirect_uri":  {fmt.Sprintf("http://localhost:%d", chosenPort)},
			}.Encode()),
		)
		if err != nil {
			fail("could not reach /oauth2/token", safeDetail(err.Error()))
			return
		}
		defer tokenResp.Body.Close()

		body, err := readResponseBody(tokenResp)
		if err != nil {
			fail("could not read /oauth2/token response", safeDetail(err.Error()))
			return
		}

		// Treat any non-200 as a hard failure before consuming any fields, so
		// we never accept an access token from -- or hide the status of -- an
		// unsuccessful response, regardless of whether its body happens to
		// parse as JSON.
		if tokenResp.StatusCode != http.StatusOK {
			fail(fmt.Sprintf("/oauth2/token returned HTTP %d", tokenResp.StatusCode), safeDetail(redactSensitive(body)))
			return
		}

		var tokenData map[string]interface{}
		if err := json.Unmarshal(body, &tokenData); err != nil {
			fail("could not parse /oauth2/token response", safeDetail(err.Error()+": "+redactSensitive(body)))
			return
		}

		if errCode, ok := tokenData["error"].(string); ok {
			detail := errCode
			if errDescription, ok2 := tokenData["error_description"].(string); ok2 {
				detail = errCode + ": " + errDescription
			}
			fail("/oauth2/token returned an error", safeDetail(detail))
			return
		}

		accessTokenString, ok := tokenData["access_token"].(string)
		if !ok {
			fail("no access token in /oauth2/token response", "")
			return
		}

		// be defensive about type
		var expiresIn time.Duration
		switch v := tokenData["expires_in"].(type) {
		case float64:
			expiresIn = time.Duration(v) * time.Second
		case int:
			expiresIn = time.Duration(v) * time.Second
		default:
			expiresIn = 0
		}

		///////////////////////////////////////////
		// Get user's information (specifically email)
		///////////////////////////////////////////

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/userinfo", notehubApiHost), nil)
		if err != nil {
			fail("could not create /userinfo request", safeDetail(err.Error()))
			return
		}
		req.Header.Set("Authorization", "Bearer "+accessTokenString)
		userinfoResp, err := http.DefaultClient.Do(req)
		if err != nil {
			fail("could not reach /userinfo", safeDetail(err.Error()))
			return
		}
		defer userinfoResp.Body.Close()

		userinfoBody, err := readResponseBody(userinfoResp)
		if err != nil {
			fail("could not read /userinfo response", safeDetail(err.Error()))
			return
		}

		if userinfoResp.StatusCode != http.StatusOK {
			fail(fmt.Sprintf("/userinfo returned HTTP %d", userinfoResp.StatusCode), safeDetail(redactSensitive(userinfoBody)))
			return
		}

		var userinfoData map[string]interface{}
		if err := json.Unmarshal(userinfoBody, &userinfoData); err != nil {
			fail("could not parse /userinfo response", safeDetail(err.Error()+": "+redactSensitive(userinfoBody)))
			return
		}

		// /userinfo may omit "email" depending on IdP configuration; fall
		// back to the subject identifier, which OIDC requires the userinfo
		// response to include.
		email, _ := userinfoData["email"].(string)
		if email == "" {
			sub, _ := userinfoData["sub"].(string)
			if sub == "" {
				fail("/userinfo response missing both email and sub", "")
				return
			}
			email = sub
		}

		///////////////////////////////////////////
		// Build the access token response
		///////////////////////////////////////////

		accessToken = &AccessToken{
			Host:        notehubApiHost,
			Email:       email,
			AccessToken: accessTokenString,
			ExpiresAt:   time.Now().Add(expiresIn),
		}

		///////////////////////////////////////////
		// respond to the browser and quit
		///////////////////////////////////////////

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<p>Token exchange completed successfully</p><p>You may now close this window and return to the CLI application</p>")

		quit <- os.Interrupt
	})

	// Pick first available port and get a listener
	listener, port, err := listenOnAny(ports)
	if err != nil {
		return nil, fmt.Errorf("could not bind any callback port: %w", err)
	}
	chosenPort = port

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", chosenPort),
		Handler: router,
	}

	// Wait for OAuth callback to be hit, then shutdown HTTP server
	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("error: %v", err)
		}
		close(done)
	}()

	// Start HTTP server waiting for OAuth callback
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("error: %v", err)
		}
	}()

	// Build the authorize URL using the chosen port
	authorizeUrl := url.URL{
		Scheme: "https",
		Host:   notehubUiHost,
		Path:   "/oauth2/auth",
		RawQuery: url.Values{
			"client_id":             {clientId},
			"code_challenge":        {codeChallenge},
			"code_challenge_method": {"S256"},
			"redirect_uri":          {fmt.Sprintf("http://localhost:%d", chosenPort)},
			"response_type":         {"code"},
			"scope":                 {"openid email"},
			"state":                 {state},
		}.Encode(),
	}

	// Open web browser to authorize
	fmt.Printf("Opening web browser to initiate authentication (redirect port %d)...\n", chosenPort)
	if err := open(authorizeUrl.String()); err != nil {
		fmt.Printf("error opening web browser: %v", err)
	}

	// Wait for exchange to finish
	<-done

	// A shutdown with neither result set means the flow was interrupted (e.g.
	// an OS signal) before any callback completed. Return an error rather than
	// a nil token and nil error, which a caller would dereference.
	if accessToken == nil && accessTokenErr == nil {
		accessTokenErr = errors.New("authentication canceled before completion")
	}
	return accessToken, accessTokenErr
}
