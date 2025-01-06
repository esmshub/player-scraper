//go:build ssl

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"player-scraper/internal/core"
	"player-scraper/internal/ssl"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

var (
	flagUrl           = flag.String("teams-url", "http://www.ssl2001.ukhome.net/teams.htm", "URL for teams page on the SSL website")
	flagDownloadFiles = flag.Bool("download-files", true, "Download rosters from the SSL website if missing")
	flagDir           = flag.String("dir", ".", "Local directory of rosters")
	flagOutputDir     = flag.String("output-dir", ".", "Output directory for CSV files")
	flagMaxParallel   = flag.Int("max-concurrent", 5, "Number of concurrent requests")
	flagStopOnError   = flag.Bool("stop-on-error", true, "Stop all requests on first error")
)

func main() {
	fmt.Println("\nSSL Player Scraper")
	fmt.Println("------------------")

	flag.Parse()

	parsedUrl, err := url.Parse(*flagUrl)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}

	fmt.Print("Loading clubs")
	provider := ssl.NewTeamProvider(*flagUrl)
	rosters, err := provider.Load()
	if err != nil {
		log.Fatalf("Failed to load rosters: %v", err)
	}
	fmt.Println("\t\t ... done!")

	tracker := &progress.Tracker{
		DeferStart:         false,
		RemoveOnCompletion: false,
		Message:            "Scraping rosters",
		Total:              int64(len(rosters)),
		Units:              progress.UnitsDefault,
	}
	tracker.SetValue(0)

	errors := []error{}
	ctx, cancel := context.WithCancel(context.Background())
	loader := &core.FileRosterLoader{
		Dir:           *flagDir,
		RemoteUrl:     fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host),
		DownloadFiles: *flagDownloadFiles,
		MaxConcurrent: *flagMaxParallel,
		OnLoaded: func(r *core.RosterFile) {
			tracker.Increment(1)
		},
		OnError: func(e error) {
			errors = append(errors, e)
			if *flagStopOnError {
				cancel()
			} else {
				tracker.IncrementWithError(1)
			}
		},
	}
	// instantiate a Progress Writer and set up the options
	pw := progress.NewWriter()
	pw.AppendTracker(tracker)
	pw.SetAutoStop(false)
	pw.SetMessageLength(24)
	pw.SetNumTrackersExpected(1)
	pw.SetTrackerLength(len(rosters))
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleDefault)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"
	// render async
	go pw.Render()

	loader.Load(rosters, ctx)

	ticker := time.Tick(time.Millisecond * 100)
	for !tracker.IsDone() {
		select {
		case <-ctx.Done():
			tracker.MarkAsErrored()
		case <-ticker:
			if tracker.Value() >= tracker.Total {
				tracker.MarkAsDone()
			}
		}
	}

	pw.Stop()
	if !tracker.IsErrored() {
		_, err = core.ExportToCsv(rosters, *flagOutputDir, "ssl_players_", "SSL Player List")
		if err != nil {
			log.Fatalf("Failed to create output CSV file: %v", err)
		}

		fmt.Println("Press enter key to close ...")
		fmt.Scanln()
	} else if len(errors) > 0 {
		fmt.Println("Errors occurred while loading rosters:")
		for _, e := range errors {
			fmt.Printf(" - %v\n", e)
		}
	}
}
