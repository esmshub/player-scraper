package core

import "context"

type ScraperOptions struct {
	LocalOnly     bool
	DownloadFiles bool
	RosterDir     string
	OutputDir     string
	ExcelExport   bool
}

type RosterLoader interface {
	Load(rosters []*RosterFile, context context.Context)
}

type RosterFile struct {
	Name                string
	Code                string
	League              string
	FileLocation        string
	InfoFileLocation    string
	AcademyFileLocation string
	Rows                *[][]string
	InfoRows            *[][]string
	Failures            int
}
