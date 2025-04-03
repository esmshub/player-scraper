package core

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var cashRegex = regexp.MustCompile(`[^\d.]`)

func getRowByVal(rows [][]string, colIndex int, value string) []string {
	var result []string
	for _, r := range rows {
		for i, v := range r {
			if i == colIndex {
				if strings.EqualFold(v, value) {
					result = r
				}

				break
			}
		}

		if len(result) > 0 {
			break
		}
	}

	return result
}

func getColByValue(row []string, value string) int {
	return slices.IndexFunc(row, func(v string) bool {
		return strings.HasSuffix(v, value)
	})
}

func normalizePlayerWage(wage string) string {
	return extractNumbers(wage)
}

func normalizePlayerValue(value string) string {
	normalized := extractNumbers(value)
	if strings.HasSuffix(value, "K") {
		if val, err := strconv.ParseFloat(normalized, 32); err == nil {
			normalized = fmt.Sprintf("%g", val/1000)
		}
	}
	return normalized
}

func extractNumbers(input string) string {
	// Remove any non-numeric characters (including commas and letters)
	return cashRegex.ReplaceAllString(input, "")
}

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

	hasInfo := false
	// write each club as a CSV row
	records := [][]string{}
	clubCount := 0
	for _, r := range rosters {
		if r.Rows != nil {
			clubCount++
			wageIndex, valueIndex := -1, -1
			if r.InfoRows != nil {
				wageIndex = getColByValue((*r.InfoRows)[0], "Wage")
				valueIndex = getColByValue((*r.InfoRows)[0], "Value")
			}
			for i, row := range *r.Rows {
				if i > 0 { // skip header row
					fields := append([]string{r.Name, r.Code, r.League}, row...)
					if r.InfoRows != nil {
						inf := getRowByVal(*r.InfoRows, 0, row[0])
						if len(inf) > max(wageIndex, valueIndex) {
							hasInfo = true
							wage := normalizePlayerWage(inf[wageIndex])
							val := normalizePlayerValue(inf[valueIndex])
							fields = append(fields, wage, val)
						}
					}
					records = append(records, fields)
				}
			}
		}
	}

	if hasInfo {
		headers = append(headers, "Wage (K)", "Value (M)")
	}
	writer.Write(headers)

	color.Blue("Finished\t\t ... Players=%d, Clubs=%d\n", len(records), clubCount)
	writer.WriteAll(records)
	absPath, err := filepath.Abs(file.Name())
	if err != nil {
		color.Yellow("Failed to get absolute path: %v", err)
		absPath = file.Name()
	}

	color.Green("Export file\t\t ... %s.", absPath)
	return absPath, nil
}
