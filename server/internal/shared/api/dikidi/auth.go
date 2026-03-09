package dikidi

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
)

func (c *Client) AcquireTelegramCSRFToken(ctx context.Context, client *req.Client) (string, error) {
	var resp *req.Response
	var err error
	c.limitCall(func() {
		resp, err = client.R().
			SetContext(ctx).
			SetHeaders(map[string]string{
				"Sec-Fetch-Dest": "document",
				"Sec-Fetch-Mode": "navigate",
				"Sec-Fetch-Site": "none",
				"Sec-Fetch-User": "?1",
			}).
			Get("https://dikidi.net/550001?p=0.pi-ssm")
	})
	if err != nil {
		return "", fmt.Errorf("api client: acquire telegram CSRF token: request failed: %w", err)
	}

	regex := regexp.MustCompile(`name="telegram_csrf" value="([^"]+)"`)
	matches := regex.FindStringSubmatch(resp.String())
	if matches == nil {
		return "", errors.New("api client: acquire telegram CSRF token: no token found")
	}

	return matches[1], nil
}

func (c *Client) AcquireCSRFToken(ctx context.Context, client *req.Client, req CSRFTokenRequest) (string, error) {
	var authData APIAuth
	var err error
	c.limitCall(func() {
		_, err = client.R().
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

	})
	if err != nil {
		return "", fmt.Errorf("api client: acquire CSRF token: request failed: %w", err)
	}

	regex := regexp.MustCompile(`name="csrf" value="([^"]+)"`)
	matches := regex.FindStringSubmatch(authData.HTML)
	if matches == nil {
		return "", errors.New("api client: acquire CSRF token: no token found")
	}

	return matches[1], nil
}

func (c *Client) SendAuthRequest(ctx context.Context, client *req.Client, req AuthRequest) error {
	var err error
	c.limitCall(func() {
		_, err = client.R().
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
	})
	if err != nil {
		return fmt.Errorf("api client: send auth request: request failed: %w", err)
	}

	return nil
}

func (c *Client) AcquireSessionID(cookieName string) (string, error) {
	parts := strings.Split(cookieName, "~")
	if len(parts) != 2 {
		return "", errors.New("api client: acquire session id: bad cookie")
	}
	return parts[1], nil
}

func (c *Client) RenewCookies(ctx context.Context, client *req.Client) error {
	var err error
	c.limitCall(func() {
		_, err = client.R().
			SetContext(ctx).
			SetHeaders(map[string]string{
				"Sec-Fetch-Dest": "document",
				"Sec-Fetch-Mode": "navigate",
				"Sec-Fetch-Site": "none",
				"Sec-Fetch-User": "?1",
			}).
			Get("https://dikidi.net/550001?p=0.pi-ssm")
	})
	if err != nil {
		return fmt.Errorf("api client: renew cookies: request failed: %w", err)
	}
	return nil
}

func (c *Client) AcquireClientCookies(client *req.Client) (*ClientCookies, error) {
	var cookies ClientCookies
	all := make(map[string]string)

	rootCookies, err := client.GetCookies("https://dikidi.net")
	if err != nil {
		return nil, fmt.Errorf("api client: acquire client cookies: failed to get root cookies: %w", err)
	}

	authCookies, err := client.GetCookies("https://auth.dikidi.net")
	if err != nil {
		return nil, fmt.Errorf("api client: acquire client cookies: failed to get auth cookies: %w", err)
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

	return &cookies, nil
}
