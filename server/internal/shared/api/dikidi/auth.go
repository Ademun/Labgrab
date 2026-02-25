package dikidi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"labgrab/internal/shared/errors"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTelegramCSRFToken(ctx context.Context, client *req.Client) (string, error) {
	resp := client.Get("https://dikidi.net/550001?p=0.pi-ssm").
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "document",
			"Sec-Fetch-Mode": "navigate",
			"Sec-Fetch-Site": "none",
			"Sec-Fetch-User": "?1",
		}).
		Do(ctx)
	if resp.Err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Fetch main page",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Read main page",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Read main page HTML",
			Err:       err,
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

	return matches[1], nil
}

func (c *Client) AcquireCSRFToken(ctx context.Context, client *req.Client, req CSRFTokenRequest) (string, error) {
	resp := client.Post("https://auth.dikidi.net/ajax/check/auth/").
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-site",
			"Origin":         "https://dikidi.net",
			"Referer":        "https://dikidi.net/550001?p=0.pi-ssm",
		}).
		SetFormData(map[string]string{
			"telegram_csrf": req.TelegramCSRFToken,
			"number":        req.PhoneNumber,
		}).
		Do(ctx)
	if resp.Err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Fetch form CSRF token",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Read form CSRF token",
			Err:       err,
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Read form CSRF token body",
			Err:       err,
		}
	}
	fmt.Printf("[AcquireCSRFToken] response body: %s\n", body)

	var authData AuthResponse
	if err := json.Unmarshal(body, &authData); err != nil {
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

	return matches[1], nil
}

func (c *Client) SendAuthRequest(ctx context.Context, client *req.Client, req AuthRequest) error {
	resp := client.Post("https://auth.dikidi.net/ajax/user/auth/").
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-site",
			"Origin":         "https://dikidi.net",
			"Referer":        "https://dikidi.net/550001?p=0.pi-ssm",
		}).
		SetFormData(map[string]string{
			"telegram_csrf": req.TelegramCSRFToken,
			"number":        req.PhoneNumber,
			"csrf":          req.CSRFToken,
			"password":      req.Password,
			"pdAgreement":   "1",
		}).
		Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "SendAuthRequest",
			Step:      "Post auth request",
			Err:       resp.Err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &errors.ExternalAPIError{
			Procedure: "SendAuthRequest",
			Step:      "Post auth request",
			Err:       fmt.Errorf("invalid status code %d", resp.StatusCode),
		}
	}

	return nil
}

func (c *Client) AcquireSessionID(cookieName string) (string, error) {
	parts := strings.Split(cookieName, "~")
	if len(parts) != 2 {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireSessionID",
			Step:      "Parse cookie_name",
			Err:       fmt.Errorf("invalid cookie_name"),
		}
	}
	return parts[1], nil
}
