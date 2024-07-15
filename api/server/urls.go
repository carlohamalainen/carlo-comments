package server

import (
	"fmt"
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/carlohamalainen/carlo-comments/conduit"
)

func (s *Server) InitHost(ctx context.Context, host string) error {
	logger := conduit.GetLogger(ctx)

	logger.Info("initialising host", "host", host)

	urls, err := scrape(logger, "https://"+host)
	if err != nil {
		logger.Error("failed to scrape", "error", err)
		return err
	}
	for _, url := range urls {
		logger.Info("setting known blog post", "host", host, "url", url)
		s.SetKnown(ctx, host, url)
	}
	return nil
}

func scrape(logger *slog.Logger, pageURL string) ([]string, error) {
	// The simple thing here would be to do
	// resp, err := http.Get(pageURL)
	// but gosec then complains about https://cwe.mitre.org/data/definitions/88.html

	// Intead, parse the URL:
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		logger.Error("failed to parse URL", "error", err)
		return nil, fmt.Errorf("invalid URL: %v", err)
	}
	// Ensure that it's http or https:
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		logger.Error("invalid URL scheme", "scheme", parsedURL.Scheme)
		return nil, fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}
	// Construct a GET request using the parsed URL:
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to get page", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		logger.Error("failed to parse body", "error", err)
		return make([]string, 0), err
	}

	if _, err = url.Parse(pageURL); err != nil {
		logger.Error("failed to parse url", "url", pageURL, "error", err)
		return make([]string, 0), err
	}

	var links []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if strings.HasPrefix(a.Val, "/2") {

						if !conduit.IsValidPostID(a.Val) {
							logger.Warn("ignoring link", "page_url", pageURL, "link_url", a.Val)
							continue
						}

						links = append(links, strings.TrimSuffix(a.Val, "/"))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}
