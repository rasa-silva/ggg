package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

func DeleteNote(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return ExitWithMessage("Need note id or (partial) title")
	}

	err := FindAndOpenDB()
	if err != nil {
		return ExitWithError(err, "DB file not found. Did you call 'init' yet?")
	}

	filter := c.Args().First()
	if strings.HasPrefix(filter, "#") {
		id, err := strconv.Atoi(filter[1:])
		if err != nil {
			return ExitWithError(err, "Invalid Id")
		}
		err = DeleteNoteById(id)
		if err != nil {
			return ExitWithError(err, "Could not delete note")
		}
	} else {
		notes, err := FindNotesBySubstring(filter)
		if err != nil {
			return ExitWithError(err, "Could not find notes")
		}
		if len(notes) > 1 {
			fmt.Printf("This will delete %v notes. Proceed? [yN] ", len(notes))
			var confirm string
			fmt.Scanf("%s", &confirm)
			if confirm != "y" {
				fmt.Println("Delete canceled.")
				return nil
			}
		}

		n, err := DeleteNoteByTitle(filter)
		if err != nil {
			return ExitWithError(err, "Could not delete note")
		}
		fmt.Println("Deleted", n, "notes.")
		return nil
	}

	return nil
}
