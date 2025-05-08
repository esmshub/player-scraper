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

func getLetterForCol(col int) string {
	if col < 1 {
		return ""
	}

	result := ""
	for col > 0 {
		col-- // Decrement to make it 0-indexed
		result = string('A'+(col%26)) + result
		col /= 26
	}

	return result
}

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
	idx := -1
	for i, v := range row {
		if strings.HasSuffix(v, value) {
			idx = i
			break
		}
	}
	return idx
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

func getColInt(row []string, index int) int {
	val, err := strconv.Atoi(row[index])
	if err != nil {
		color.Yellow("Failed to convert column to integer: %v", err)
	}

	return val
}

func ExportToCsv(rosters []*RosterFile, outputDir string, useExcelFormulas bool, fileNamePrefix string, title string) (string, error) {

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

	// adjust headers for <stat>/min columns
	statCols := []string{"Sav", "Ktk", "Kps", "Gls"}
	for _, colName := range statCols {
		colIdx := slices.Index(headers, colName)
		if colIdx > -1 {
			headers = slices.Insert(headers, colIdx+1, fmt.Sprintf("%s/min", colName))
		}
	}

	minColIdx := slices.Index(headers, "Min")
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
					// add <stat>/min col values
					if minColIdx > -1 && len(fields) > minColIdx {
						minsPlayed := getColInt(fields, minColIdx)
						for _, statName := range statCols {
							statColIdx := slices.Index(headers, statName)
							if statColIdx > -1 && len(fields) > statColIdx {
								val := "0"
								if useExcelFormulas {
									val = fmt.Sprintf("=IFERROR(INDEX(%s:%[1]s, ROW()) / INDEX(%s:%[2]s, ROW()), 0)", getLetterForCol(statColIdx+1), getLetterForCol(minColIdx+1))
								} else {
									statVal := getColInt(fields, statColIdx)
									if statVal != 0 && minsPlayed != 0 {

										val = fmt.Sprintf("%f", float64(statVal)/float64(minsPlayed))
									}
								}
								fields = slices.Insert(fields, statColIdx+1, val)
							}
						}
					}
					// add wage and value columns
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
		headers = append(headers, "Wage", "Mkt Value")
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
