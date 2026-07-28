// Package domain holds shared LOLA artifact entities for scrape and graph ingest.
package domain

type Artifact struct {
	Name            string
	Description     string
	OS              []string
	Commands        []Command
	Paths           []string
	Detection       Detection
	Resources       []Resource
	Acknowledgement []Person
	MitreID         string
	Category        string
	Privileges      string
}

type Command struct {
	Command     string
	Description string
	Usecase     string
	Category    string
	Privileges  string
	MitreID     string
	OS          []string
	Tags        []string
}

type Detection struct {
	Sigma []string
	Yara  []string
}

type Resource struct {
	Link string
}

type Person struct {
	Name   string
	Handle string
}
