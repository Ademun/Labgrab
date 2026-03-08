package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/errors"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTelegramCSRFToken(ctx context.Context, client *req.Client) (string, error) {
	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "document",
			"Sec-Fetch-Mode": "navigate",
			"Sec-Fetch-Site": "none",
			"Sec-Fetch-User": "?1",
		}).
		Get("https://dikidi.net/550001?p=0.pi-ssm")
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Fetch main page",
			Err:       err,
		}
	}

	regex := regexp.MustCompile(`name="telegram_csrf" value="([^"]+)"`)
	matches := regex.FindStringSubmatch(resp.String())
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
	var authData APIAuth
	resp, err := client.R().
		SetContext(ctx).
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
		SetSuccessResult(&authData).
		Post("https://auth.dikidi.net/ajax/check/auth/")
	if err != nil {
		return "", &errors.ExternalAPIError{
			Procedure: "AcquireCSRFToken",
			Step:      "Fetch form CSRF token",
			Err:       err,
		}
	}
	fmt.Printf("[AcquireCSRFToken] response body: %s\n", resp.String())

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
	fmt.Println(req)
	_, err := client.R().
		SetContext(ctx).
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
		Post("https://auth.dikidi.net/ajax/user/auth/")
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "SendAuthRequest",
			Step:      "Post auth request",
			Err:       err,
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

func (c *Client) RenewCookies(ctx context.Context, client *req.Client) error {
	_, err := client.R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "document",
			"Sec-Fetch-Mode": "navigate",
			"Sec-Fetch-Site": "none",
			"Sec-Fetch-User": "?1",
		}).
		Get("https://dikidi.net/550001?p=0.pi-ssm")
	if err != nil {
		return &errors.ExternalAPIError{
			Procedure: "RenewCookies",
			Step:      "Fetch main page",
			Err:       err,
		}
	}
	return nil
}

func (c *Client) AcquireClientCookies(client *req.Client) (*ClientCookies, error) {
	var cookies ClientCookies
	all := make(map[string]string)

	rootCookies, err := client.GetCookies("https://dikidi.net")
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireClientCookies",
			Step:      "Get root cookies",
			Err:       err,
		}
	}

	authCookies, err := client.GetCookies("https://auth.dikidi.net")
	if err != nil {
		return nil, &errors.ExternalAPIError{
			Procedure: "AcquireClientCookies",
			Step:      "Get auth cookies",
			Err:       err,
		}
	}

	for _, cookie := range rootCookies {
		all[cookie.Name] = cookie.Value
		if cookie.Name == "cookie_name" {
			cookies.CookieName = &cookie.Value
		}
		if cookie.Name == "token" {
			cookies.Token = &cookie.Value
		}
	}

	for _, cookie := range authCookies {
		all[cookie.Name] = cookie.Value
		if cookie.Name == "cookie_name" {
			cookies.CookieName = &cookie.Value
		}
		if cookie.Name == "token" {
			cookies.Token = &cookie.Value
		}
	}

	allList := make([]string, 0, len(all))
	for name, value := range all {
		allList = append(allList, fmt.Sprintf("%s=%s", name, value))
	}
	cookies.All = strings.Join(allList, ";")
	fmt.Println(cookies.All)

	return &cookies, nil
}
