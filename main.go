package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

var (
	verbose       bool
	verbosePrefix = text.FgBlue.Sprint("V:")
	errorPrefix   = text.FgHiRed.Sprint("ERROR:")
)

// GetDBPath returns the full path to the database file
func GetDBPath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dbPath := path.Join(hd, ".ggg", "notes.db")
	LogOnVerbose(fmt.Sprint("Using DB at", dbPath))
	return dbPath, nil
}

// InitDB creates and initializes a new database
func InitDB(c *cli.Context) error {
	path, err := GetDBPath()
	if err != nil {
		return err
	}

	appDir := filepath.Dir(path)
	err = os.MkdirAll(appDir, 0700)
	if err != nil {
		return err
	}

	err = CreateDB(path)
	if err != nil {
		return ExitWithError(err, "Could not create DB")
	}

	fmt.Println("Created", path)
	return nil
}

func FindAndOpenDB() error {
	path, err := GetDBPath()
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	return OpenDB(path)
}

func main() {

	app := &cli.App{
		Usage: "GoGoGadget notes!",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Value:       false,
				Usage:       "Verbose output",
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize a new database",
				Action:  InitDB,
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Add a new note",
				Action:  NewNote,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List notes containing the search term. if no option is passed, all notes are shown",
				Action:  ListNotes,
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete note by id or title",
				Action:  DeleteNote,
			},
			{
				Name:    "open",
				Aliases: []string{"o"},
				Usage:   "Open note by id or title",
				Action:  OpenNote,
			},
			{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Find matches on notes",
				Action:  FindMatches,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
