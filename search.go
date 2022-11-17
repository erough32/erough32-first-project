package main

import (
	"fmt"
	"strings"

	"github.com/yunge/sphinx"
)

var opts *sphinx.Options = &sphinx.Options{
	Host:       "127.0.0.1",
	Port:       9312,
	Timeout:    5000,
	MaxMatches: 1000,
	Limit:      500,
}

func searchAny(text string) []string {
	sc := sphinx.NewClient(opts)
	sc.SetSortMode(sphinx.SPH_SORT_EXTENDED, "time desc")
	res, err := sc.Query(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						text, "/", `\/`,
					), "(", `\(`,
				), ")", `\)`,
			), "!", `\!`,
		), "book1", "Search Query()",
	)
	if err != nil {
		fmt.Println("search err", err)
	}

	var searchInfo []string
	for _, match := range res.Matches {
		searchInfo = append(searchInfo, fmt.Sprint(match.DocId))
	}

	return searchInfo
}
