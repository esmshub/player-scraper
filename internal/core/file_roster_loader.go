package core

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corpix/uarand"
)

type FileRosterLoader struct {
	RemoteUrl     string
	DownloadFiles bool
	Dir           string
	MaxConcurrent int
	OnLoaded      func(*RosterFile)
	OnError       func(error)
}

func (l *FileRosterLoader) Load(rosters []*RosterFile, ctx context.Context) {
	if l.DownloadFiles {
		if _, err := os.Stat(l.Dir); err != nil {
			if os.IsNotExist(err) {
				if l.OnError != nil {
					l.OnError(fmt.Errorf("directory does not exist: %s", l.Dir))
				}
			}
		}
	}

	// Create a semaphore with a buffer size of maxConcurrent
	sem := make(chan struct{}, int(math.Max(1, float64(l.MaxConcurrent))))
	defer close(sem)

	errCh := make(chan error)
	defer close(errCh)

	resultCh := make(chan *RosterFile)
	defer close(resultCh)

	var wg sync.WaitGroup

	parser := &TextRosterParser{}
	// Create a new HTTP client
	client := &http.Client{}
	downloadFile := func(filePath string) ([]byte, error) {
		fileUrl, err := url.JoinPath(l.RemoteUrl, filePath)
		if err != nil {
			return nil, err
		}

		// Create a new request with the URL you want to access
		req, err := http.NewRequest("GET", fileUrl, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", uarand.GetRandom())

		// Send the request
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("file download failed: %s - %s", fileUrl, res.Status)
		}

		return io.ReadAll(res.Body)
	}
	loadCsvFile := func(filePath string) (*[][]string, error) {
		file, err := os.Open(filePath)
		if err == nil {
			defer file.Close()
			return parser.Parse(file)
		} else {
			return nil, err
		}
	}
	loadAcademy := func(roster *RosterFile) error {
		time.Sleep(50 * time.Millisecond)
		content, err := downloadFile(roster.AcademyFileLocation)
		if err != nil {
			return err
		}

		academyRows, err := parser.Parse(strings.NewReader(string(content)))
		if err != nil {
			return err
		}

		if len(*academyRows) > 0 {
			*roster.Rows = append(*roster.Rows, (*academyRows)[1:]...)
		}

		return nil
	}
	loadInfo := func(roster *RosterFile) error {
		time.Sleep(50 * time.Millisecond)
		infoContent, err := downloadFile(roster.InfoFileLocation)
		if err != nil {
			return err
		}

		infoRows, err := parser.Parse(strings.NewReader(string(infoContent)))
		if err != nil {
			return err
		}

		roster.InfoRows = infoRows
		return nil
	}
	loadAndParse := func(roster *RosterFile) error {
		localPath := filepath.Join(l.Dir, roster.Code+".txt")
		if l.RemoteUrl != "" {
			// check remote
			contents, err := downloadFile(roster.FileLocation)
			if err != nil {
				return err
			}

			parsed, err := parser.Parse(strings.NewReader(string(contents)))
			if err != nil {
				return err
			}
			roster.Rows = parsed

			// scrape academy
			if roster.AcademyFileLocation != "" {
				loadAcademy(roster)
			}

			// scrape info remotely
			if roster.InfoFileLocation != "" {
				loadInfo(roster)
			}

			if l.DownloadFiles {
				// save roster locally
				err = os.WriteFile(localPath, contents, 0644)
				if err != nil {
					return err
				}
			}
		} else {
			rows, err := loadCsvFile(localPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			roster.Rows = rows
		}

		// load info locally if exists
		infLocalPath := filepath.Join(l.Dir, fmt.Sprintf("INFO_%s.txt", roster.Code))
		if infRows, err := loadCsvFile(infLocalPath); err == nil {
			roster.InfoRows = infRows
		}

		return nil
	}

	for _, r := range rosters {
		wg.Add(1)
		go func(roster *RosterFile) {
			defer wg.Done()

			// Acquire a semaphore
			sem <- struct{}{}

			var done bool = false
			for roster.Failures < 3 && !done {

				time.Sleep(50 * time.Millisecond)

				select {
				case <-ctx.Done():
					// If the context is cancelled, do nothing
				default:
					if err := loadAndParse(roster); err != nil {
						if roster.Failures < 3 {
							// allow retry
							roster.Failures++
						} else {
							done = true
							errCh <- err
						}
					} else {
						done = true
						resultCh <- roster
					}
				}
			}

			// Release the semaphore
			<-sem
		}(r)
	}

	go func() {
		for {
			select {
			case ros, ok := <-resultCh:
				if ok && l.OnLoaded != nil {
					l.OnLoaded(ros)
				}
			case err, ok := <-errCh:
				if ok && l.OnError != nil {
					l.OnError(err)
				}
			}
		}
	}()

	wg.Wait()
}
