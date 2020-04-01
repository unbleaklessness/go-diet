package main

import (
	"database/sql"
	"time"
)

type day struct {
	id   int
	date string
}

func now() string {
	return time.Now().Format("2006-01-02")
}

func createToday(db *sql.DB) (ie *ierror) {
	_, e := db.Exec(`insert or ignore into days (date) values ($1)`, now())
	if e != nil {
		ie = &ierror{m: "Could not create today", e: e}
		return
	}
	return
}

func selectToday(db *sql.DB) (d day, ie *ierror) {
	rows, e := db.Query(`select id, date from days where date = $1`, now())
	if e != nil {
		ie = &ierror{m: "Could not select today", e: e}
		return
	}
	defer rows.Close()

	if rows.Next() {

		e = rows.Scan(&d.id, &d.date)
		if e != nil {
			ie = &ierror{m: "Could not scan today", e: e}
			return
		}

		return
	}

	ie = &ierror{m: "Could not find today", e: e}
	return
}

func today(db *sql.DB) day {
	exit := func() {
		panic("Could not get today")
	}

	ie := createToday(db)
	if ie != nil {
		exit()
	}

	d, ie := selectToday(db)
	if ie != nil {
		exit()
	}

	return d
}
