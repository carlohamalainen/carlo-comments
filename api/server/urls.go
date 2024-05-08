package server

import (
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

	resp, err := http.Get(pageURL)
	if err != nil {
		logger.Error("failed to get page", "error", err)
		return make([]string, 0), err
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
