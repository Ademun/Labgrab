package dikidi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/imroc/req/v3"
)

func acquireTelegramCsrfToken(ctx context.Context, client *req.Client) (string, error) {
	resp := client.Get("https://dikidi.net/550001?p=0.pi-ssm").Do(ctx)
	if resp.Err != nil {
		return "", resp.Err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`name="telegram_csrf" value="([^"]+)"`)
	telegramCsrf := regex.FindStringSubmatch(string(body))[1]
	return telegramCsrf, nil
}

func acquireCsrfToken(ctx context.Context, client *req.Client, req CSRFTokenRequest) (string, error) {
	resp := client.Post("https://auth.dikidi.net/ajax/check/auth/").SetFormData(map[string]string{
		"telegram_csrf": req.TelegramCSRFToken,
		"number":        req.PhoneNumber,
	}).Do(ctx)
	if resp.Err != nil {
		return "", resp.Err
	}
	defer resp.Body.Close()

	var authData AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authData); err != nil {
		return "", err
	}

	regex := regexp.MustCompile(`name="csrf" value="([^"]+)"`)
	csrf := regex.FindStringSubmatch(authData.HTML)[1]

	return csrf, nil
}

func sendAuthRequest(ctx context.Context, client *req.Client, req AuthRequest) error {
	resp := client.Post("https://auth.dikidi.net/ajax/user/auth/").SetFormData(map[string]string{
		"number":        req.PhoneNumber,
		"password":      req.Password,
		"csrf":          req.CSRFToken,
		"telegram_csrf": req.TelegramCSRFToken,
	}).Do(ctx)
	if resp.Err != nil {
		return resp.Err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	return nil
}
