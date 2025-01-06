package core

import (
	"bufio"
	"io"
	"strings"
)

type TextRosterParser struct{}

func (p *TextRosterParser) Parse(data io.Reader) (*[][]string, error) {
	scanner := bufio.NewScanner(data)
	rows := [][]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "---") {
			continue
		}

		rows = append(rows, strings.Fields(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &rows, nil
}
