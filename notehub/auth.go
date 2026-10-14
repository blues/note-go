package notehub

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
)

// callbackAction classifies an inbound request to the local OAuth callback
// server.
type callbackAction int

const (
	// callbackIgnore: not the OAuth redirect (e.g. favicon, prefetch, a bare
	// visit to the root). Must be answered benignly without affecting the flow.
	callbackIgnore callbackAction = iota
	// callbackStateMismatch: a redirect whose state doesn't match the value the
	// flow generated (a CSRF attempt or a stale redirect). Must fail closed.
	callbackStateMismatch
	// callbackError: the provider reported an authorization failure (e.g. the
	// user denied consent), redirecting with an "error" parameter and no code.
	callbackError
	// callbackProceed: a well-formed redirect carrying a code whose state
	// matches.
	callbackProceed
)

// classifyCallback decides how to treat a request to the callback server. A
// request carrying neither an authorization code nor an OAuth error is not the
// redirect at all and is ignored. Anything that is the redirect is honored only
// if its state matches the value the flow generated; an "error" parameter (with
// no code) signals that authorization was refused or failed.
func classifyCallback(r *http.Request, expectedState string) callbackAction {
	q := r.URL.Query()
	code := q.Get("code")
	oauthErr := q.Get("error")

	if code == "" && oauthErr == "" {
		return callbackIgnore
	}
	if q.Get("state") != expectedState {
		return callbackStateMismatch
	}
	if oauthErr != "" {
		return callbackError
	}
	return callbackProceed
}

const browserLoginMessage = "Opening web browser to initiate authentication...\n"

// Both login modes preserve the provider's human-readable error description.
func authorizationError(code, description string) error {
	message := code
	if description != "" {
		message += ": " + description
	} else if code == "access_denied" {
		message = "sign-in was cancelled"
	}
	return fmt.Errorf("authorization was not granted: %s", message)
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

// The login flow's OAuth state and PKCE code verifier both have to be
// unguessable, so they come from crypto/rand. RFC 7636
func randString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

// loginHosts preserves the historical API/UI hostname mapping in both modes.
func loginHosts(apiHost string) (string, string) {
	if !strings.HasPrefix(apiHost, "api.") {
		apiHost = "api." + apiHost
	}
	uiHost := strings.TrimPrefix(apiHost, "api.")
	if apiHost == "api.notefile.net" {
		uiHost = "notehub.io"
	}
	return apiHost, uiHost
}

// signInCompleteURL is the Notehub page shown once the localhost callback is
// handled. result is one of "success", "denied" or "failed".
func signInCompleteURL(uiHost, result string) string {
	u := url.URL{Scheme: "https", Host: uiHost, Path: "/login-complete", RawQuery: url.Values{"result": {result}}.Encode()}
	return u.String()
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
	return InitiateBrowserBasedLoginWithOptions(notehubApiHost, BrowserLoginOptions{})
}

// BrowserLoginOptions supplies an optional editable sign-in name.
type BrowserLoginOptions struct {
	SignInName string // Empty uses the registered client name.
}

// InitiateBrowserBasedLoginWithOptions starts a localhost callback login.
func InitiateBrowserBasedLoginWithOptions(notehubApiHost string, options BrowserLoginOptions) (*AccessToken, error) {
	return initiateBrowserBasedLogin(notehubApiHost, open, options)
}

func initiateBrowserBasedLogin(notehubApiHost string, openBrowser func(string) error, options BrowserLoginOptions) (*AccessToken, error) {
	// this is the hard-coded OAuth client ID that's persisted in Hydra
	clientId := "notehub_cli"

	notehubApiHost, notehubUiHost := loginHosts(notehubApiHost)

	// Try these ports in order until one is available:
	//
	// these ports are randomly chosen and hard-coded into
	// the OAuth client in Hydra within Notehub (in the redirect_uris field)
	ports := []int{58766, 58767, 58768, 58769, 42100, 42101, 42102, 42103}

	loginCtx, cancelLogin := context.WithCancel(context.Background())
	defer cancelLogin()

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

		switch action {
		case callbackIgnore:
			// Not the OAuth redirect (favicon, prefetch, a bare visit to the
			// root); ignore it without affecting the flow.
			w.WriteHeader(http.StatusNoContent)
			return
		case callbackStateMismatch:
			// A request whose state doesn't match this attempt is unrelated to
			// it -- a stray local request, another browser tab, or (since the
			// callback ports are a predictable, hard-coded list) a webpage
			// probing localhost. Reject it benignly: it must neither abort the
			// in-progress login nor consume the single-callback slot below, so
			// it cannot be used to deny service to a legitimate sign-in.
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Only a redirect whose state matches reaches here. Handle exactly one;
		// answer any duplicate benignly so a second request cannot race the
		// shared result or signal shutdown twice.
		handled := false
		once.Do(func() { handled = true })
		if !handled {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "authentication already completed; you may close this window")
			return
		}

		// Every terminal callback, including denial and token-exchange failure,
		// must release the caller waiting for the callback server to stop.
		defer func() { quit <- os.Interrupt }()

		authorizationCode := r.URL.Query().Get("code")

		// The browser lands on a Notehub page styled like the consent screen;
		// failure details go to the caller and stderr, not on that page.
		complete := func(result string) {
			http.Redirect(w, r, signInCompleteURL(notehubUiHost, result), http.StatusSeeOther)
		}

		errHandler := func(result, msg string) {
			fmt.Fprintf(os.Stderr, "error: %s\n", msg)
			accessTokenErr = errors.New(msg)
			complete(result)
		}

		// The provider refused or failed the authorization (e.g. the user
		// denied consent). Report it instead of waiting for a code that will
		// never arrive.
		if action == callbackError {
			oauthErr := r.URL.Query().Get("error")
			result := "failed"
			if oauthErr == "access_denied" {
				result = "denied"
			}
			errHandler(result, authorizationError(oauthErr, r.URL.Query().Get("error_description")).Error())
			return
		}

		accessToken, accessTokenErr = exchangeAuthorizationCode(loginCtx, notehubApiHost, notehubUiHost, clientId, authorizationCode, codeVerifier, fmt.Sprintf("http://localhost:%d", chosenPort))
		if accessTokenErr != nil {
			errHandler("failed", accessTokenErr.Error())
			return
		}

		complete("success")
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
		cancelLogin()
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
	if options.SignInName != "" {
		params := authorizeUrl.Query()
		params.Set("signin_name", options.SignInName)
		authorizeUrl.RawQuery = params.Encode()
	}

	// Open web browser to authorize
	fmt.Fprint(os.Stderr, browserLoginMessage)
	if err := openBrowser(authorizeUrl.String()); err != nil {
		fmt.Fprintf(os.Stderr, "error opening web browser: %v", err)
	}

	// Wait for exchange to finish
	<-done
	if accessToken == nil && accessTokenErr == nil {
		accessTokenErr = context.Canceled
	}
	return accessToken, accessTokenErr
}
