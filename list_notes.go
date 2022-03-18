package main

import (
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func ListNotes(c *cli.Context) error {
	err := FindAndOpenDB()
	if err != nil {
		return ExitWithError(err, "DB file not found. Did you call 'init' yet?")
	}

	var notes []Note
	if c.Args().Len() == 0 {
		notes, err = FindAllNotes()
	} else {
		notes, err = FindNotesBySubstring(c.Args().First())
	}

	if err != nil {
		return ExitWithError(err, "Could not find notes")
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Title", "Created", "Modified"})
	for _, note := range notes {
		ct, _ := time.Parse(time.RFC3339, note.created)
		mt, _ := time.Parse(time.RFC3339, note.modified)
		t.AppendRow(table.Row{note.id, note.title, humanize.Time(ct), humanize.Time(mt)})
	}
	t.SetStyle(table.StyleRounded)
	t.Style().Color.Header = text.Colors{text.FgYellow}
	t.Style().Options.SeparateRows = true
	t.Render()

	return nil
}
