package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func FindMatches(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return ExitWithMessage("Need search term")
	}

	err := FindAndOpenDB()
	if err != nil {
		return ExitWithError(err, "DB file not found. Did you call 'init' yet?")
	}

	filter := c.Args().First()

	matchStart := text.FgHiYellow.EscapeSeq()
	matchEnd := text.FgHiWhite.EscapeSeq()
	trailing := text.Faint.Sprint(" (...) ")
	matches, err := FindMatchesOnAllNotes(filter, matchStart, matchEnd, trailing)
	if err != nil {
		return cli.Exit("ERROR: Problem finding matches: "+err.Error(), 1)
	}

	if len(matches) == 0 {
		fmt.Println("No matches found.")
	} else {
		for _, m := range matches {
			println(text.FgHiBlue.Sprintf("#%v - '%v':", m.id, m.title))
			println(m.snippet)
			println()
		}
	}

	return nil
}
