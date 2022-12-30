package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const filename = "versions.json"

var aliases = []string{"dev", "latest"}

type Version struct {
	Version string   `json:"version"`
	Title   string   `json:"title"`
	Aliases []string `json:"aliases"`
}

func (v *Version) ContainsAlias(alias string) bool {
	for _, a := range v.Aliases {
		if a == alias {
			return true
		}
	}
	return false
}

func main() {
	var versions []*Version
	b, err := os.ReadFile(filename)
	checkErr(err)
	err = json.Unmarshal(b, &versions)
	checkErr(err)

	for _, v := range versions {
		v.Title = v.Version
		for _, alias := range aliases {
			if v.ContainsAlias(alias) {
				v.Title = fmt.Sprintf("%s (%s)", v.Version, alias)
			}
		}
	}

	b, err = json.Marshal(&versions)
	checkErr(err)
	err = os.WriteFile(filename, b, 664)
	checkErr(err)
}

func checkErr(msg interface{}) {
	if msg != nil {
		fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(1)
	}
}
