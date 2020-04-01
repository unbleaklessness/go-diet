package main

import (
	"database/sql"
)

type day struct {
	id   int
	date string
}

func today(db *sql.DB) (day, ierrori) {

	var (
		e         error
		rows      *sql.Rows
		d         day
		thisError func(e error) (day, ierrori)
	)

	thisError = func(e error) (day, ierrori) {
		return d, ierror{m: "Could not get today", e: e}
	}

	_, e = db.Exec(`insert or ignore into days (date) values ($1)`, now())
	if e != nil {
		return thisError(e)
	}

	rows, e = db.Query(`select id, date from days where date = $1`, now())
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	if rows.Next() {
		e = rows.Scan(&d.id, &d.date)
		if e != nil {
			return thisError(e)
		}
	}

	return d, nil
}
