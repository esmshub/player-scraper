package ffo

import (
	"fmt"
	"net/url"
	"path"
	"player-scraper/internal/core"
	"strings"

	"github.com/gocolly/colly"
)

type FfoWebTeamProvider struct {
	url string
}

func (p *FfoWebTeamProvider) Load() ([]*core.RosterFile, error) {
	rosters := []*core.RosterFile{}
	var err error

	rootUrl, err := url.Parse(p.url)
	if err != nil {
		return rosters, err
	}

	// instantiate a new collector object
	c := colly.NewCollector(
		colly.AllowedDomains(rootUrl.Hostname()),
	)

	// triggered when the scraper encounters an error
	c.OnError(func(_ *colly.Response, e error) {
		err = e
	})

	// triggered when a CSS selector matches an element
	c.OnHTML("a", func(e *colly.HTMLElement) {
		u, err := url.Parse(e.Request.AbsoluteURL(e.Attr("href")))
		if err != nil {
			panic(err)
		}

		if strings.HasPrefix(u.Path, "/clubs/club_pages") {
			code := strings.TrimSuffix(path.Base(u.Path), path.Ext(path.Base(u.Path)))
			league := strings.TrimSuffix(path.Base(e.Request.URL.Path), path.Ext(e.Request.URL.Path))
			rosters = append(rosters, &core.RosterFile{
				Code:         code,
				League:       league,
				FileLocation: fmt.Sprintf("text_files/%s/roster/%s.txt", league, code),
			})
		} else if !strings.HasPrefix(e.Attr("href"), rootUrl.String()) {
			c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		}
	})

	c.Visit(rootUrl.String())
	c.Wait()

	return rosters, err
}

func NewTeamProvider(url string) *FfoWebTeamProvider {
	return &FfoWebTeamProvider{
		url: url,
	}
}
