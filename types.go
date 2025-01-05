package main

import "context"

type RosterLoader interface {
	Load(rosters []*RosterFile, context context.Context)
}

type RosterFile struct {
	Name         string
	Code         string
	League       string
	FileLocation string
	Rows         *[][]string
}
