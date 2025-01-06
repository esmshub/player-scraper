package ssl

import (
	"context"
	"fmt"
	"net/url"
	"player-scraper/internal/core"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type SslWebRosterLoader struct {
	url           *url.URL
	maxConcurrent int
	onLoaded      func(*core.RosterFile)
	onError       func(error)
}

func (l *SslWebRosterLoader) Load(rosters []*core.RosterFile, context context.Context) {
	parser := &core.TextRosterParser{}

	collector := colly.NewCollector(
		colly.AllowedDomains(l.url.Hostname()),
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Use OnRequest callback to associate the context
	collector.OnRequest(func(r *colly.Request) {
		select {
		case <-context.Done(): // If context is cancelled, stop the request
			r.Abort() // Abort the request
		default:
			//
		}
	})

	// triggered when the scraper encounters an error
	collector.OnError(func(_ *colly.Response, e error) {
		if l.onError != nil {
			l.onError(e)
		}
	})

	// fired when the server responds
	collector.OnResponse(func(r *colly.Response) {
		for _, ros := range rosters {
			if strings.HasSuffix(r.Request.URL.String(), ros.FileLocation) {
				rows, err := parser.Parse(strings.NewReader(string(r.Body)))
				if err != nil {
					if l.onError != nil {
						l.onError(err)
					}
				} else {
					ros.Rows = rows
					if l.onLoaded != nil {
						l.onLoaded(ros)
					}
				}
				break
			}
		}
	})

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: l.maxConcurrent,
		RandomDelay: 100 * time.Millisecond,
	})

	for _, r := range rosters {
		absUrl, err := url.JoinPath(fmt.Sprintf("%s://%s", l.url.Scheme, l.url.Host), r.FileLocation)
		if err != nil {
			if l.onError != nil {
				l.onError(err)
			}
		} else {
			collector.Visit(absUrl)
		}
	}

	collector.Wait()
}
