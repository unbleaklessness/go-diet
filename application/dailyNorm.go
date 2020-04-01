package main

import (
	"database/sql"
	"fmt"
)

type dailyNorm struct {
	kcals    float32
	proteins float32
	carbs    float32
	fats     float32
}

func scanDailyNorm() (d dailyNorm, e error) {

	fmt.Print("Kcals: ")
	_, e = fmt.Scanln(&d.kcals)
	if e != nil {
		return
	}

	fmt.Print("Proteins: ")
	_, e = fmt.Scanln(&d.proteins)
	if e != nil {
		return
	}

	fmt.Print("Carbs: ")
	_, e = fmt.Scanln(&d.carbs)
	if e != nil {
		return
	}

	fmt.Print("Fats: ")
	_, e = fmt.Scanln(&d.fats)
	if e != nil {
		return
	}

	return
}

func removeDailyNorm(db *sql.DB) (e error) {
	_, e = db.Exec(`delete from dailyNorm`)
	return
}

func addDailyNorm(db *sql.DB) (ie *ierror) {
	e := removeDailyNorm(db)
	if e != nil {
		ie = &ierror{m: "Could not delete old daily norm", e: e}
		return
	}

	d, e := scanDailyNorm()
	if e != nil {
		ie = &ierror{m: "Failed to read daily norm", e: e}
		return
	}

	_, e = db.Exec(`insert into dailyNorm
		(kcals, proteins, carbs, fats)
		values
		($1, $2, $3, $4)
	`, d.kcals, d.proteins, d.carbs, d.fats)
	if e != nil {
		ie = &ierror{m: "Could not create a new daily norm", e: e}
		return
	}

	return
}
