package main

import (
	"database/sql"
)

func createTables(db *sql.DB) (e error) {
	_, e = db.Exec(`create table if not exists products (
		id integer primary key autoincrement,
		name text not null,
		kcals real not null,
		proteins real not null,
		carbs real not null,
		fats real not null
	)`)
	if e != nil {
		return
	}

	_, e = db.Exec(`create table if not exists days (
		id integer primary key autoincrement,
		date text not null unique on conflict ignore
	)`)
	if e != nil {
		return
	}

	_, e = db.Exec(`create table if not exists dayProducts (
		id integer primary key autoincrement,
		dayId integer not null,
		productId integer not null,
		amount real not null
	)`)
	if e != nil {
		return
	}

	_, e = db.Exec(`create table if not exists dailyNorm (
		kcals real not null,
		proteins real not null,
		carbs real not null,
		fats real not null
	)`)
	if e != nil {
		return
	}

	return
}
