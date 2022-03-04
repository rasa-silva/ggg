package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	db *sql.DB
)

type Note struct {
	id       int
	title    string
	body     string
	created  string
	modified string
}

type SearchResult struct {
	id      int
	title   string
	snippet string
}

func CreateDB(dbPath string) error {
	OpenDB(dbPath)
	_, err := db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS Notes USING fts5(title, body, created, modified, tokenize=trigram)")
	if err != nil {
		return err
	}

	return nil
}

func OpenDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	return nil
}

func FindAllNotes() ([]Note, error) {
	LogOnVerbose("Finding all notes...")
	results, err := db.Query("select rowid, title, created, modified from Notes")
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0, 10)
	for results.Next() {
		var rowid int
		var title, created, modified string
		if err = results.Scan(&rowid, &title, &created, &modified); err != nil {
			return nil, err
		}

		notes = append(notes, Note{id: rowid, title: title, created: created, modified: modified})
	}

	return notes, nil
}

func FindNoteById(id int) (*Note, error) {
	LogOnVerbose(fmt.Sprint("Finding note with id", id, "..."))
	query := fmt.Sprintf("SELECT title, body FROM Notes WHERE rowid = %v", id)
	result, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var title string
	var body string
	if succ := result.Next(); !succ {
		return nil, errors.New("could not find note by Id")
	}
	defer result.Close()

	if err = result.Scan(&title, &body); err != nil {
		return nil, err
	}

	return &Note{id: id, title: title, body: body}, nil
}

func FindNotesBySubstring(substring string) ([]Note, error) {
	LogOnVerbose(fmt.Sprint("Finding notes with substring ", substring, "..."))
	results, err := db.Query("SELECT rowid, title, created, modified FROM Notes WHERE title MATCH ?", substring)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0, 10)
	for results.Next() {
		var rowid int
		var title, created, modified string
		if err = results.Scan(&rowid, &title, &created, &modified); err != nil {
			return nil, err
		}

		notes = append(notes, Note{id: rowid, title: title, created: created, modified: modified})
	}

	return notes, nil
}

func InsertNote(note Note) error {
	LogOnVerbose(fmt.Sprintf("Inserting note '%v'...", note.title))
	_, err := db.Exec("INSERT INTO Notes VALUES (?, ?, ?, ?)", note.title, note.body, note.created, note.modified)
	return err
}

func UpdateNote(note *Note) error {
	_, err := db.Exec("UPDATE Notes SET title = ?, body = ?, modified = ? WHERE rowid = ?", note.title, note.body, note.modified, note.id)
	return err
}

func DeleteNoteByTitle(title string) (int, error) {
	LogOnVerbose(fmt.Sprint("Deleting note titled", title, "..."))
	res, err := db.Exec("DELETE FROM Notes WHERE title MATCH ?", title)
	if err != nil {
		return -1, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}

	return int(rows), nil
}

func DeleteNoteById(id int) error {
	LogOnVerbose(fmt.Sprint("Deleting note with id", id, "..."))
	res, err := db.Exec("DELETE FROM Notes WHERE rowid = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return errors.New("id not found")
	}
	return err
}

func FindMatchesOnAllNotes(term string, matchStart string, matchEnd string, trailing string) ([]SearchResult, error) {
	results, err := db.Query("SELECT rowid, title, snippet(Notes, 1, ?, ?, ?, 200) FROM Notes WHERE body MATCH ?", matchStart, matchEnd, trailing, term)
	if err != nil {
		return nil, err
	}

	matches := make([]SearchResult, 0, 10)
	for results.Next() {
		var rowid int
		var title string
		var snippet string
		if err = results.Scan(&rowid, &title, &snippet); err != nil {
			return nil, err
		}

		matches = append(matches, SearchResult{id: rowid, title: title, snippet: snippet})
	}

	return matches, nil

}
