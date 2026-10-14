package notehub

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type pollLoginResponse struct {
	BrowserURL       string `json:"browser_url"`
	PollToken        string `json:"poll_token"`
	ExpiresIn        int    `json:"expires_in"`
	Interval         int    `json:"interval"`
	Status           string `json:"status"`
	Code             string `json:"code"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
}

// PollingLoginOptions controls the sign-in label and browser/progress handling.
type PollingLoginOptions struct {
	SignInName       string                                              // Empty uses the registered client name.
	ExpiresInSeconds int                                                 // Asks the hub to keep the sign-in open this long, up to its own limit; 0 for its usual.
	OnURL            func(browserURL string, expiresInSeconds int) error // Overrides opening the browser, given the URL and the seconds the hub granted it.
	OnPending        func(remainingSeconds int) error                    // Called every ten seconds.
}

// InitiatePollingLogin opens the browser and waits using outbound HTTPS only.
func InitiatePollingLogin(ctx context.Context, apiHost string) (*AccessToken, error) {
	return InitiatePollingLoginWithOptions(ctx, apiHost, PollingLoginOptions{})
}

// InitiatePollingLoginWithOptions supports either interactive or agent login.
func InitiatePollingLoginWithOptions(ctx context.Context, apiHost string, options PollingLoginOptions) (*AccessToken, error) {
	return initiatePollingLogin(ctx, apiHost, &http.Client{Timeout: 30 * time.Second}, os.Stderr, open, options)
}

func initiatePollingLogin(ctx context.Context, apiHost string, client *http.Client, stderr io.Writer, openBrowser func(string) error, options PollingLoginOptions) (*AccessToken, error) {
	apiHost, uiHost := loginHosts(apiHost)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute) // a backstop beyond whatever the hub grants
	defer cancel()
	verifier := randString(50)
	hash := sha256.Sum256([]byte(verifier))
	endpoint := "https://" + uiHost + "/auth/poll"
	payload := map[string]any{
		"mode": "start", "client_id": "notehub_cli",
		"code_challenge": base64.RawURLEncoding.EncodeToString(hash[:]), "code_challenge_method": "S256",
	}
	if options.SignInName != "" {
		payload["signin_name"] = options.SignInName
	}
	if options.ExpiresInSeconds > 0 {
		payload["expires_in"] = options.ExpiresInSeconds
	}
	start, err := requestPollLogin(ctx, client, endpoint, "", payload)
	if err != nil {
		return nil, err
	}
	if start.BrowserURL == "" || start.PollToken == "" || start.ExpiresIn <= 0 || start.Interval <= 0 {
		return nil, fmt.Errorf("invalid polling login response")
	}
	ctx, cancelTransaction := context.WithTimeout(ctx, time.Duration(start.ExpiresIn)*time.Second)
	defer cancelTransaction()
	if options.OnURL != nil {
		if err := options.OnURL(start.BrowserURL, start.ExpiresIn); err != nil {
			return nil, err
		}
	} else {
		fmt.Fprint(stderr, browserLoginMessage)
		if err := openBrowser(start.BrowserURL); err != nil {
			fmt.Fprintf(stderr, "error opening web browser: %v", err)
		}
	}

	return awaitPollingLoginWithProgress(ctx, client, endpoint, start, func(code string) (*AccessToken, error) {
		return exchangeAuthorizationCode(ctx, apiHost, uiHost, "notehub_cli", code, verifier, endpoint)
	}, options.OnPending)
}

func awaitPollingLogin(ctx context.Context, client *http.Client, endpoint string, start pollLoginResponse, exchange func(string) (*AccessToken, error)) (*AccessToken, error) {
	return awaitPollingLoginWithProgress(ctx, client, endpoint, start, exchange, nil)
}

func awaitPollingLoginWithProgress(ctx context.Context, client *http.Client, endpoint string, start pollLoginResponse, exchange func(string) (*AccessToken, error), onPending func(remainingSeconds int) error) (*AccessToken, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		code string
		err  error
	}
	done := make(chan result, 1)
	go func() { code, err := pollAuthorizationCode(ctx, client, endpoint, start); done <- result{code, err} }()
	complete := func(r result) (*AccessToken, error) {
		if r.err != nil {
			return nil, r.err
		}
		return exchange(r.code)
	}
	var pending <-chan time.Time
	if onPending != nil {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		pending = ticker.C
	}
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authorization stopped: %w", ctx.Err())
		case r := <-done:
			return complete(r)
		case <-pending:
			// Do not emit another waiting message once a terminal result is ready.
			select {
			case r := <-done:
				return complete(r)
			default:
			}
			remainingSeconds := 0
			if deadline, ok := ctx.Deadline(); ok {
				remaining := time.Until(deadline)
				if remaining > 0 {
					remainingSeconds = int((remaining + time.Second - 1) / time.Second)
				}
			}
			if err := onPending(remainingSeconds); err != nil {
				return nil, err
			}
		}
	}
}

func pollAuthorizationCode(ctx context.Context, client *http.Client, endpoint string, start pollLoginResponse) (string, error) {
	interval := 2 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("authorization stopped: %w", ctx.Err())
		case <-timer.C:
			result, err := requestPollLogin(ctx, client, endpoint, start.PollToken, map[string]any{"mode": "poll"})
			if err != nil {
				var networkError *url.Error
				if errors.As(err, &networkError) && ctx.Err() == nil {
					timer.Reset(interval)
					continue
				}
				return "", err
			}
			switch result.Status {
			case "pending":
				delay := interval
				if time.Duration(result.Interval)*time.Second > interval {
					delay = time.Duration(result.Interval) * time.Second
				}
				timer.Reset(delay)
			case "code_available":
				if result.Code == "" {
					return "", fmt.Errorf("authorization returned no code")
				}
				return result.Code, nil
			case "denied", "failed":
				return "", authorizationError(result.Error, result.ErrorDescription)
			case "expired":
				return "", fmt.Errorf("sign-in expired; please try again")
			default:
				return "", fmt.Errorf("unexpected authorization status: %s", result.Status)
			}
		}
	}
}

func requestPollLogin(ctx context.Context, client *http.Client, endpoint, credential string, payload map[string]any) (pollLoginResponse, error) {
	var result pollLoginResponse
	body, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	if credential != "" {
		req.Header.Set("Authorization", "Bearer "+credential)
	}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if credential != "" && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
		delay := 5
		if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > delay {
			delay = seconds
		}
		if when, err := http.ParseTime(resp.Header.Get("Retry-After")); err == nil && int(time.Until(when).Seconds())+1 > delay {
			delay = int(time.Until(when).Seconds()) + 1
		}
		return pollLoginResponse{Status: "pending", Interval: delay}, nil
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16384)).Decode(&result); err != nil {
		return result, fmt.Errorf("polling login returned HTTP %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode == http.StatusGone && result.Status == "expired" {
		return result, nil
	}
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("polling login returned HTTP %d: %s", resp.StatusCode, result.Message)
	}
	return result, nil
}

// Both callback modes use the same token exchange and identity lookup.
func exchangeAuthorizationCode(ctx context.Context, apiHost, uiHost, clientID, code, verifier, redirectURI string) (*AccessToken, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	form := url.Values{
		"client_id": {clientID}, "code": {code}, "code_verifier": {verifier},
		"grant_type": {"authorization_code"}, "redirect_uri": {redirectURI},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+uiHost+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()
	var token struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int64  `json:"expires_in"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16384)).Decode(&token); err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	if token.Error != "" {
		return nil, fmt.Errorf("%s: %s", token.Error, token.ErrorDescription)
	}
	if resp.StatusCode != http.StatusOK || token.AccessToken == "" {
		return nil, fmt.Errorf("token exchange returned HTTP %d with no usable token", resp.StatusCode)
	}
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://"+apiHost+"/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	info, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo: %w", err)
	}
	defer info.Body.Close()
	var identity struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(io.LimitReader(info.Body, 16384)).Decode(&identity); err != nil {
		return nil, fmt.Errorf("userinfo: %w", err)
	}
	if info.StatusCode != http.StatusOK || identity.Email == "" {
		return nil, fmt.Errorf("userinfo returned HTTP %d with no email", info.StatusCode)
	}
	return &AccessToken{Host: apiHost, Email: identity.Email, AccessToken: token.AccessToken, ExpiresAt: expiresAt}, nil
}
