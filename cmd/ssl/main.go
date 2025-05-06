//go:build ssl

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"player-scraper/internal/core"
	"player-scraper/internal/ssl"
	"player-scraper/internal/ui"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/skratchdot/open-golang/open"
)

var (
	flagTeamsUrl      = flag.String("teams-url", "http://www.ssl2001.ukhome.net/teams.htm", "URL to scrape for team information on SSL website")
	flagDownloadFiles = flag.Bool("download-files", false, "Download the latest rosters from the SSL website")
	flagRostersDir    = flag.String("rosters-dir", ".", "Target directory for downloading or sourcing local rosters")
	flagOutputDir     = flag.String("output-dir", ".", "Output directory for CSV files")
	flagMaxParallel   = flag.Int("max-concurrent", 5, "Number of concurrent requests when loading roster files")
	flagStopOnError   = flag.Bool("stop-on-error", false, "Stop all requests on first error")
	flagExcelExport   = flag.Bool("excel-export", true, "Use Excel-compatible formulas instead of raw values")
	flagCiMode        = flag.Bool("ci", false, "Run in CI mode and disable prompts")
)

var appName = "SSL Player Scraper"

func main() {
	flag.Parse()

	parsedUrl, err := url.Parse(*flagTeamsUrl)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}

	opts := core.ScraperOptions{
		LocalOnly:     false,
		DownloadFiles: *flagDownloadFiles,
		RosterDir:     *flagRostersDir,
		OutputDir:     *flagOutputDir,
		ExcelExport:   *flagExcelExport,
	}

	ciMode := true
	if flagCiMode == nil || !*flagCiMode {
		ciMode = false
		var cancelled bool
		opts, cancelled = ui.Run(appName)
		if cancelled {
			os.Exit(0)
		}
	}
	// } else {
	// 	fmt.Println("\n" + appName)
	// 	fmt.Println("------------------")
	// }

	fmt.Print(fmt.Sprintf("\n%s\n", ui.StyleTitle(appName)))

	fmt.Print("Loading clubs")
	provider := ssl.NewTeamProvider(parsedUrl.String())
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
	remoteUrl := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	if opts.LocalOnly {
		remoteUrl = ""
	}
	loader := &core.FileRosterLoader{
		Dir:           opts.RosterDir,
		RemoteUrl:     remoteUrl,
		DownloadFiles: opts.DownloadFiles,
		MaxConcurrent: *flagMaxParallel,
		OnLoaded: func(r *core.RosterFile) {
			tracker.Increment(1)
		},
		OnError: func(e error) {
			errors = append(errors, e)
			if !opts.LocalOnly && *flagStopOnError {
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

	_, err = core.ExportToCsv(rosters, opts.OutputDir, opts.ExcelExport, "ssl_players_", "SSL Player List")
	if err != nil {
		log.Fatalf("Failed to create output CSV file: %v", err)
	}

	if len(errors) > 0 {
		color.Red("Errors occurred while loading rosters:\n")
		for _, e := range errors {
			color.Red(" - %v\n", e)
		}
	}

	if !ciMode {
		fmt.Println("Press enter key to close ...")
		fmt.Scanln()
		open.Start(opts.OutputDir)
	}
}
