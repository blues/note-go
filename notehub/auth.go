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
	"strings"
	"time"
)

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

	router := http.NewServeMux()

	// We'll fill this after we pick a port but declare it now so the handler can close over it.
	chosenPort := 0

	// The browser will be redirected to this endpoint with an authorization code
	// and then this endpoint will exchange that authorization code for an access token
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authorizationCode := r.URL.Query().Get("code")
		callbackState := r.URL.Query().Get("state")

		errHandler := func(msg string) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error: %s", msg)
			fmt.Printf("error: %s\n", msg)
			accessTokenErr = errors.New(msg)
		}

		if callbackState != state {
			errHandler("state mismatch")
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
			errHandler("error on /oauth2/token: " + err.Error())
			return
		}
		defer tokenResp.Body.Close()

		body, err := io.ReadAll(tokenResp.Body)
		if err != nil {
			errHandler("could not read body from /oauth2/token: " + err.Error())
			return
		}

		var tokenData map[string]interface{}
		if err := json.Unmarshal(body, &tokenData); err != nil {
			errHandler("could not unmarshal body from /oauth2/token: " + err.Error())
			return
		}

		if errCode, ok := tokenData["error"].(string); ok {
			if errDescription, ok2 := tokenData["error_description"].(string); ok2 {
				errHandler(fmt.Sprintf("%s: %s", errCode, errDescription))
			} else {
				errHandler(errCode)
			}
			return
		}

		accessTokenString, ok := tokenData["access_token"].(string)
		if !ok {
			errHandler("unexpected error: no access token returned")
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
			errHandler("could not create request for /userinfo: " + err.Error())
			return
		}
		req.Header.Set("Authorization", "Bearer "+accessTokenString)
		userinfoResp, err := http.DefaultClient.Do(req)
		if err != nil {
			errHandler("could not get userinfo: " + err.Error())
			return
		}
		defer userinfoResp.Body.Close()

		userinfoBody, err := io.ReadAll(userinfoResp.Body)
		if err != nil {
			errHandler("could not read body from /userinfo: " + err.Error())
			return
		}

		var userinfoData map[string]interface{}
		if err := json.Unmarshal(userinfoBody, &userinfoData); err != nil {
			errHandler("could not unmarshal body from /userinfo: " + err.Error())
			return
		}

		email, ok := userinfoData["email"].(string)
		if !ok {
			errHandler("could not retrieve email")
			return
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
	return accessToken, accessTokenErr
}
