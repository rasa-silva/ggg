package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/urfave/cli/v2"

	_ "modernc.org/sqlite"
)

var (
	verbose       bool
	verbosePrefix = text.FgBlue.Sprint("V:")
	errorPrefix   = text.FgHiRed.Sprint("ERROR:")
)

func GetDBPath() (string, error) {
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dbPath := path.Join(hd, ".ggg", "notes.db")
	LogOnVerbose(fmt.Sprint("Using DB at", dbPath))
	return dbPath, nil
}

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
