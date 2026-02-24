package dikidi

import (
	"context"
	"fmt"
	"labgrab/internal/shared/errors"
	"net/http"
	"strings"

	"github.com/imroc/req/v3"
)

func (c *Client) RenewCookies(ctx context.Context, client *req.Client) error {
	resp := client.Get("https://dikidi.net/550001?p=0.pi-ssm").
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "document",
			"Sec-Fetch-Mode": "navigate",
			"Sec-Fetch-Site": "none",
			"Sec-Fetch-User": "?1",
		}).
		Do(ctx)
	if resp.Err != nil {
		return &errors.ExternalAPIError{
			Procedure: "AcquireTelegramCSRFToken",
			Step:      "Fetch main page",
			Err:       resp.Err,
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

	allList := make([]string, 0)
	for name, value := range all {
		allList = append(allList, fmt.Sprintf("%s=%s", name, value))
	}
	cookies.All = strings.Join(allList, ";")
	fmt.Println(cookies.All)

	return &cookies, nil
}

func parseCookieString(raw string) []*http.Cookie {
	var cookies []*http.Cookie
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, value, _ := strings.Cut(part, "=")
		cookies = append(cookies, &http.Cookie{
			Name:  strings.TrimSpace(name),
			Value: strings.TrimSpace(value),
		})
	}
	return cookies
}
