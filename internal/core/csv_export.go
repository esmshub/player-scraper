package core

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

func ExportToCsv(rosters []*RosterFile, outputDir, fileNamePrefix string, title string) (string, error) {

	if _, err := os.Stat(outputDir); err != nil {
		if os.IsNotExist(err) {
			color.New(color.FgBlue).Println("Output directory does not exist, it will be created.")
			os.Mkdir(outputDir, 0755)
		}
	}

	// open the CSV file
	outputFile := path.Join(outputDir, fmt.Sprintf("%s%d.csv", fileNamePrefix, time.Now().Unix()))
	file, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{fmt.Sprintf("%s (scraped on %s)", title, time.Now().Format(time.DateTime))})
	writer.Write([]string{})
	headers := []string{
		"Team",
		"Code",
		"League",
		"Name",
		"Age",
		"Nat",
		"St",
		"Tk",
		"Ps",
		"Sh",
		"Ag",
		"KAb",
		"TAb",
		"PAb",
		"SAb",
		"Gam",
		"Sub",
		"Min",
		"Mom",
		"Sav",
		"Con",
		"Ktk",
		"Kps",
		"Sht",
		"Gls",
		"Ass",
		"DP",
		"Inj",
		"Sus",
	}

	writer.Write(headers)

	// write each club as a CSV row
	records := [][]string{}
	for _, r := range rosters {
		for i, row := range *r.Rows {
			if i > 0 { // skip header row
				fields := append([]string{r.Name, r.Code, r.League}, row...)
				records = append(records, fields)
			}
		}
	}

	color.Blue("Finished\t\t ... Players=%d, Clubs=%d\n", len(records), len(rosters))
	writer.WriteAll(records)
	absPath, err := filepath.Abs(file.Name())
	if err != nil {
		color.Yellow("Failed to get absolute path: %v", err)
		absPath = file.Name()
	}

	color.Green("Export file\t\t ... %s.", absPath)
	return absPath, nil
}
