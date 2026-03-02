package dikidi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

func (c *Client) ScrapeServices(ctx context.Context, client *req.Client, serviceProviderURL string) ([]int, error) {
	doc, err := c.GetDocument(ctx, client, serviceProviderURL)
	if err != nil {
		return nil, fmt.Errorf("eror scraping document from url %s: %w", serviceProviderURL, err)
	}

	idList := make([]int, 0)
	var parsingErr error
	doc.Find(".newrecord2").Each(func(_ int, s *goquery.Selection) {
		dataOptions, exists := s.Attr("data-options")
		if !exists {
			return
		}
		var pageOptions APIHTMLPageOptions
		err := json.Unmarshal([]byte(dataOptions), &pageOptions)
		if err != nil {
			parsingErr = errors.Join(parsingErr, err)
		}
		list := pageOptions.StepData.List
		for _, list := range list {
			for _, service := range list.Services {
				idList = append(idList, service.ID)
			}
		}
	})

	if parsingErr != nil {
		return nil, parsingErr
	}

	return idList, nil
}

func (c *Client) GetDocument(ctx context.Context, client *req.Client, url string) (*goquery.Document, error) {
	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Sec-Fetch-Dest": "document",
			"Sec-Fetch-Mode": "navigate",
			"Sec-Fetch-Site": "none",
			"Sec-Fetch-User": "?1",
		}).
		Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	return doc, nil
}
