# GoGoGadget notes

This is my notes app. There are many like it, but this one is mine.

The notes are stored in a sqlite database file stored in `~/.ggg/notes.db`.
Before using the following commands, you should init the database with `ggg init`.

## Create a new note

`ggg n` will open your `$EDITOR` and you can write a new note. The title will be the first line.

## List notes

`ggg l` will list all notes. `ggg l term` will filter note titles matching the term.

## Open an existing note

`ggg o #3` will open the note with id 3 on the defined `$EDITOR`.
`ggg o term` will open the note matching the term if it's unambiguous.

## Delete note

`ggg d #3` will delete the note with id 3.
`ggg d term` wil delete the note matching the term if it's unambiguous.

## Find text

`ggg f term` will present a snippet of the note around the matched term.