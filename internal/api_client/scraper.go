package api_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) ScrapeSlotSourcesIDs(ctx context.Context, sourcesIDsProviderURL string) ([]int, error) {
	doc, err := c.ScrapeDocument(ctx, sourcesIDsProviderURL)
	if err != nil {
		return nil, fmt.Errorf("eror scraping document from url %s: %w", sourcesIDsProviderURL, err)
	}

	idList := make([]int, 0)
	var parsingErr error
	doc.Find(".newrecord2").Each(func(_ int, s *goquery.Selection) {
		dataOptions, exists := s.Attr("data-options")
		if !exists {
			return
		}
		var pageOptions HTMLPageOptions
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

func (c *Client) ScrapeDocument(ctx context.Context, sourcesIDsProviderURL string) (*goquery.Document, error) {
	res, err := c.httpClient.Get(ctx, sourcesIDsProviderURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
