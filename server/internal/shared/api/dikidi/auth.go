package dikidi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"labgrab/internal/shared/errors"
	"regexp"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTelegramCSRFToken(ctx context.Context, client *req.Client) (string, error) {
	resp := client.Get("https://dikidi.net/550001?p=0.pi-ssm").Do(ctx)
	if resp.Err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Fetch main page",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Read main page HTML",
		}
	}

	regex := regexp.MustCompile(`name="telegram_csrf" value="([^"]+)"`)
	matches := regex.FindStringSubmatch(string(body))
	if matches == nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Parse telegram CSRF token",
			Err:       fmt.Errorf("no telegram CSRF token found"),
		}
	}
	return matches[0], nil
}

func (c *Client) AcquireCSRFToken(ctx context.Context, client *req.Client, req CSRFTokenRequest) (string, error) {
	resp := client.Post("https://auth.dikidi.net/ajax/check/auth/").SetFormData(map[string]string{
		"telegram_csrf": req.TelegramCSRFToken,
		"number":        req.PhoneNumber,
	}).Do(ctx)
	if resp.Err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Fetch form CSRF token",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	var authData AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authData); err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Parse form HTML",
			Err:       err,
		}
	}

	regex := regexp.MustCompile(`name="csrf" value="([^"]+)"`)
	matches := regex.FindStringSubmatch(authData.HTML)
	if matches == nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Parse CSRF token",
			Err:       fmt.Errorf("no CSRF token found"),
		}
	}

	return matches[0], nil
}

func (c *Client) SendAuthRequest(ctx context.Context, client *req.Client, req AuthRequest) error {
	resp := client.Post("https://auth.dikidi.net/ajax/user/auth/").SetFormData(map[string]string{
		"number":        req.PhoneNumber,
		"password":      req.Password,
		"csrf":          req.CSRFToken,
		"telegram_csrf": req.TelegramCSRFToken,
	}).Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "SendAuthRequest",
			Step:      "Post auth request",
			Err:       resp.Err,
		}
	}

	if resp.StatusCode != 200 {
		return &errors.ExternalAPIError{
			Procedure: "SendAuthRequest",
			Step:      "Post auth request",
			Err:       fmt.Errorf("invalid status code %d", resp.StatusCode),
		}
	}
	return nil
}

func (c *Client) AcquireClientCookies(client *req.Client) ClientCookies {
	cookies := ClientCookies{
		Other: make(map[string]string),
	}
	for _, cookie := range client.Cookies {
		switch cookie.Name {
		case "cookie_name":
			cookies.CookieName = &cookie.Value
		case "session":
			cookies.SessionID = &cookie.Value
		default:
			cookies.Other[cookie.Name] = cookie.Value
		}
	}
	return cookies
}
