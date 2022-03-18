package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func OpenNote(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return ExitWithMessage("Need note id or (partial) title")
	}

	err := FindAndOpenDB()
	if err != nil {
		return ExitWithError(err, "DB file not found. Did you call 'init' yet?")
	}

	filter := c.Args().First()
	var note *Note
	if strings.HasPrefix(filter, "#") {
		id, err := strconv.Atoi(filter[1:])
		if err != nil {
			return ExitWithError(err, "Invalid Id")
		}

		note, err = FindNoteById(id)
		if err != nil {
			return ExitWithError(err, "Could not find note")
		}

	} else {
		notes, _ := FindNotesBySubstring(filter)
		if len(notes) > 1 {
			fmt.Printf("Ambiguous note title: %v matching notes.\n", len(notes))
		}

		note = &notes[0]
	}

	filename, err := OpenTempFileInEditor(note.body)
	if err != nil {
		return ExitWithError(err, "Could not open file")
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return ExitWithError(err, "Could not read file contents")
	}

	if note.body == string(contents) {
		fmt.Println("No changes.")
		return nil
	}

	note.body = string(contents)
	note.title = TitleFromBody(note.body)
	note.modified = time.Now().UTC().Format(time.RFC3339)
	err = UpdateNote(note)
	if err != nil {
		return ExitWithError(err, "Could not update note")
	}

	fmt.Printf("'%v' updated!\n", note.title)
	return nil
}
