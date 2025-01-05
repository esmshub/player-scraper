package main

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
)

type EsmsFileRosterLoader struct {
	remoteUrl     string
	download      bool
	dir           string
	maxConcurrent int
	onLoaded      func(*RosterFile)
	onError       func(error)
}

func (l *EsmsFileRosterLoader) Load(rosters []*RosterFile, ctx context.Context) {
	if _, err := os.Stat(l.dir); err != nil {
		if os.IsNotExist(err) {
			if l.onError != nil {
				l.onError(fmt.Errorf("directory does not exist: %s", l.dir))
			}
		}
	}

	// Create a semaphore with a buffer size of maxConcurrent
	sem := make(chan struct{}, int(math.Max(1, float64(l.maxConcurrent))))
	defer close(sem)

	errCh := make(chan error)
	defer close(errCh)

	resultCh := make(chan *RosterFile)
	defer close(resultCh)

	var wg sync.WaitGroup

	parser := &EsmsRosterParser{}
	downloadRoster := func(roster *RosterFile) ([]byte, error) {
		fileUrl, err := url.JoinPath(l.remoteUrl, roster.FileLocation)
		if err != nil {
			return nil, err
		}
		res, err := http.Get(fileUrl)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("file download failed: %s - %s (%d)", fileUrl, res.Status, res.StatusCode)
		}

		return io.ReadAll(res.Body)
	}
	loadAndParse := func(roster *RosterFile) (*[][]string, error) {
		localPath := filepath.Join(l.dir, roster.Code+".txt")
		file, err := os.Open(localPath)
		if err == nil {
			defer file.Close()
			return parser.Parse(file)
		} else if os.IsNotExist(err) && l.remoteUrl != "" {
			// check remote
			contents, err := downloadRoster(roster)
			if err != nil {
				return nil, err
			}

			parsed, err := parser.Parse(strings.NewReader(string(contents)))
			if err != nil {
				return nil, err
			}

			if l.download {
				// save roster locally
				err = os.WriteFile(localPath, contents, 0644)
				if err != nil {
					return nil, err
				}
			}

			return parsed, nil
		} else {
			return nil, err
		}
	}

	for _, r := range rosters {
		wg.Add(1)
		go func(roster *RosterFile) {
			defer wg.Done()

			// Acquire a semaphore
			sem <- struct{}{}

			time.Sleep(50 * time.Millisecond)

			select {
			case <-ctx.Done():
				// If the context is cancelled, do nothing
			default:
				rows, err := loadAndParse(roster)
				if err != nil {
					errCh <- err
				} else {
					roster.Rows = rows
					resultCh <- roster
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
				if ok && l.onLoaded != nil {
					l.onLoaded(ros)
				}
			case err, ok := <-errCh:
				if ok && l.onError != nil {
					l.onError(err)
				}
			}
		}
	}()

	wg.Wait()
}
