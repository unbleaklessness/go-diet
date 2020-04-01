package main

import (
	"database/sql"
	"fmt"
)

func addTodayProduct(db *sql.DB) ierrori {

	var (
		e         error
		rows      *sql.Rows
		d         day
		p         product
		name      string
		thisError func(e error) ierrori
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not add a product for today", e: e}
	}

	fmt.Print("Name: ")
	_, e = fmt.Scanln(&name)
	if e != nil {
		return thisError(e)
	}

	rows, e = db.Query(`select
		id, name, kcals, proteins, carbs, fats
		from products
		where name = $1
	`, name)
	if e != nil {
		return thisError(e)
	}

	if rows.Next() {
		e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			return thisError(e)
		}
	} else {
		rows.Close()
		return thisError(nil)
	}
	rows.Close()

	d, e = today(db)
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`insert into dayProducts (dayId, productId) values ($1, $2)`, d.id, p.id)
	if e != nil {
		return thisError(e)
	}

	return nil
}

func removeTodayProduct(db *sql.DB) ierrori {

	var (
		name      string
		e         error
		thisError func(e error) ierrori
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not remove product from today", e: e}
	}

	fmt.Print("Name: ")
	_, e = fmt.Scanln(&name)
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`delete from dayProducts
		where id in
		(select id from dayProducts
			where productId in
			(select id from products where name = $1)
		limit 1)`, name)
	if e != nil {
		return thisError(e)
	}

	return nil
}
