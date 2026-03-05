package mask

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"
)

func CreateRandomHTTPClient() *req.Client {
	client := req.C().
		ImpersonateChrome().
		SetCommonHeaders(map[string]string{
			"User-Agent":                "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Mobile Safari/537.36",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
			"Accept-Encoding":           "br, zstd",
			"Sec-Ch-Ua":                 `"Chromium";v="145", "Google Chrome";v="145", "Not/A)Brand";v="99"`,
			"Sec-Ch-Ua-Mobile":          "?1",
			"Sec-Ch-Ua-Platform":        `"Android"`,
			"Upgrade-Insecure-Requests": "1",
		}).
		OnAfterResponse(func(_ *req.Client, resp *req.Response) error {
			if resp.IsErrorState() {
				return fmt.Errorf("unexpected status code: %d for url %s", resp.StatusCode, resp.Request.URL)
			}
			return nil
		})

	return client
}

func CreateClientWithCookies(rawCookies *string) (*req.Client, error) {
	client := CreateRandomHTTPClient()
	if rawCookies == nil {
		return client, nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse("https://dikidi.net")
	if err != nil {
		return nil, err
	}

	jar.SetCookies(parsedURL, parseCookieString(*rawCookies))
	client.SetCookieJar(jar)
	return client, nil
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
