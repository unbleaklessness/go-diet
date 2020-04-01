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

func scanDailyNorm() (dailyNorm, error) {

	var (
		d         dailyNorm
		e         error
		thisError func(e error) (dailyNorm, ierrori)
	)

	thisError = func(e error) (dailyNorm, ierrori) {
		return d, ierror{m: "Could not scan daily norm", e: e}
	}

	fmt.Print("Kcals: ")
	_, e = fmt.Scanln(&d.kcals)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Proteins: ")
	_, e = fmt.Scanln(&d.proteins)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Carbs: ")
	_, e = fmt.Scanln(&d.carbs)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Fats: ")
	_, e = fmt.Scanln(&d.fats)
	if e != nil {
		return thisError(e)
	}

	return d, nil
}

func addDailyNorm(db *sql.DB) ierrori {

	var (
		e         error
		thisError func(e error) ierrori
		d         dailyNorm
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not add daily norm", e: e}
	}

	_, e = db.Exec(`delete from dailyNorm`)
	if e != nil {
		return thisError(e)
	}

	d, e = scanDailyNorm()
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`insert into dailyNorm
		(kcals, proteins, carbs, fats)
		values
		($1, $2, $3, $4)
	`, d.kcals, d.proteins, d.carbs, d.fats)
	if e != nil {
		return thisError(e)
	}

	return nil
}
