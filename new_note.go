package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"time"
)

func NewNote(c *cli.Context) error {
	err := FindAndOpenDB()
	if err != nil {
		return ExitWithError(err, "DB file not found. Did you call 'init' yet?")
	}

	fp, err := OpenTempFileInEditor("")
	if err != nil {
		if verbose {
			fmt.Printf("%v Problem editing note file: %v\n", errorPrefix, err)
		}
		return cli.Exit(fmt.Sprintf("%v Problem editing note file", errorPrefix), 1)
	}

	content, err := ioutil.ReadFile(fp)
	if err != nil {
		return ExitWithError(err, "Could not read file contents")
	}

	body := string(content)
	if body == "" {
		fmt.Println("Note not created since text is empty.")
		return nil
	}

	title := TitleFromBody(body)

	now := time.Now().UTC().Format(time.RFC3339)
	err = InsertNote(Note{title: title, body: body, created: now, modified: now})
	if err != nil {
		return err
	}

	fmt.Printf("'%v' added.\n", title)
	return nil
}
