package main

import (
	"bufio"
	"io"
	"strings"
)

type EsmsRosterParser struct{}

func (p *EsmsRosterParser) Parse(data io.Reader) (*[][]string, error) {
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
