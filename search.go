package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

// note is a Bear note
type note struct {
	Identifier       string // The identifier is an UUID that can be used to open this note in Bear if selected
	Title            string
	CreationDate     time.Time
	ModificationDate time.Time
}

const (
	// bearDbLocationWithinHomedir is the location of Bear's SQLite database.
	// See: https://bear.app/faq/where-are-bears-notes-located
	bearDbLocationWithinHomedir = "Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data/database.sqlite"

	// noteQuery is the primary query to search for notes in Bear's SQLite database.
	// Bear uses Core Data under the hood, which is why table and column names are a bit strange.
	// Core Data encodes dates as  floats with a different base date from UNIX timestamps, so
	// some date fiddling is required (see: https://stackoverflow.com/a/2923127/300664).
	// The query turns them back into regular UNIX timestamps. Bear sets a ZTRASHEDDATE for
	// notes that have been deleted, so we exclude those.
	noteQuery = `
	SELECT 
		ZUNIQUEIDENTIFIER, 
		ZTITLE, 
		strftime('%s', ZCREATIONDATE, 'unixepoch', '31 years'),
		strftime('%s', ZMODIFICATIONDATE, 'unixepoch', '31 years')
	FROM ZSFNOTE 
	WHERE 
	(ZTITLE LIKE ? OR ZTEXT LIKE ?)`

	noteQueryOrderby     = `ORDER BY ZMODIFICATIONDATE DESC`
	noteQueryNotTrashed  = `AND ZTRASHEDDATE IS NULL`
	noteQueryNotArchived = `AND ZARCHIVEDDATE IS NULL`
)

// search the user's notes for the given query.
func search(query string) ([]note, error) {
	db, err := openDB()
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite db")
	}
	defer db.Close()

	// col LIKE '%abc%' will search for abc anywhere in the VARCHAR column, ignoring case. Reformat
	// the query to be surrounded by wildcard % chars.
	query = fmt.Sprintf("%%%s%%", query)

	queryParts := []string{noteQuery}
	if ignoreArchived {
		queryParts = append(queryParts, noteQueryNotArchived)
	}
	if ignoreTrashed {
		queryParts = append(queryParts, noteQueryNotTrashed)
	}
	queryParts = append(queryParts, noteQueryOrderby)

	// execute query with escaped parameters
	rows, err := db.Query(strings.Join(queryParts, " "), query, query)
	if err != nil {
		return nil, errors.Wrap(err, "execute sqlite query")
	}

	notes, err := parseNotes(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parse sqlite rows")
	}
	return notes, nil
}

// openDB opens Bear's SQLite database in readonly mode
func openDB() (*sql.DB, error) {
	dbLoc := fmt.Sprintf("file:%s/%s?mode=ro", userHomeDir, bearDbLocationWithinHomedir)
	return sql.Open("sqlite", dbLoc)
}

// parseNotes from SQL query result rows
func parseNotes(rows *sql.Rows) ([]note, error) {
	defer rows.Close()

	var notes []note

	for rows.Next() {
		note := note{}

		var creationDate int64
		var modificationDate int64

		// Columns need to be scanned in the order they are selected in the NOTE_QUERY
		err := rows.Scan(&note.Identifier, &note.Title, &creationDate, &modificationDate)
		if err != nil {
			return nil, errors.Wrap(err, "scan note DB column")
		}

		note.CreationDate = time.Unix(creationDate, 0)
		note.ModificationDate = time.Unix(modificationDate, 0)

		notes = append(notes, note)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return notes, nil
}
