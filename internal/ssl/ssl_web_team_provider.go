package ssl

import (
	"fmt"
	"net/url"
	"player-scraper/internal/core"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type SslWebTeamProvider struct {
	url string
}

func (p *SslWebTeamProvider) Load() ([]*core.RosterFile, error) {
	rosters := []*core.RosterFile{}
	var err error

	url, err := url.Parse(p.url)
	if err != nil {
		return rosters, err
	}

	c := colly.NewCollector(
		colly.AllowedDomains(url.Hostname()),
	)

	c.OnError(func(_ *colly.Response, e error) {
		err = e
	})

	c.OnHTML("table", func(e *colly.HTMLElement) {
		subTables := e.DOM.Find("table") // Select the desired child elements
		subTables.Last().Find("tr").Each(func(i int, selection *goquery.Selection) {
			fields := selection.Find("td") // Select the desired child elements
			if fields.Length() == 6 && strings.HasSuffix(fields.Eq(2).Find("a").AttrOr("href", ""), ".txt") {
				rosters = append(rosters, &core.RosterFile{
					Name:         strings.TrimSpace(fields.Eq(0).Text()),
					Code:         strings.TrimSpace(fields.Eq(2).Text()),
					League:       strings.TrimSpace(fields.Eq(1).Text()),
					FileLocation: fields.Eq(2).Find("a").AttrOr("href", fmt.Sprintf("%s.txt", fields.Eq(2).Text())),
				})
				rosters = append(rosters, &core.RosterFile{
					Name:         strings.TrimSpace(fmt.Sprintf("%s Youth", fields.Eq(0).Text())),
					Code:         strings.TrimSpace(fields.Eq(4).Text()),
					League:       strings.TrimSpace(fmt.Sprintf("Youth %s", fields.Eq(3).Text())),
					FileLocation: fields.Eq(4).Find("a").AttrOr("href", fmt.Sprintf("%s.txt", fields.Eq(4).Text())),
				})
			}
		})
	})

	c.Visit(url.String())
	c.Wait()

	return rosters, err
}

func NewTeamProvider(url string) *SslWebTeamProvider {
	return &SslWebTeamProvider{
		url: url,
	}
}
