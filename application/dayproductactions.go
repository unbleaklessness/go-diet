package main

import (
	"database/sql"
	"fmt"
)

func addTodayProduct(db *sql.DB) (ie *ierror) {
	var name string
	fmt.Print("Name: ")
	_, e := fmt.Scanln(&name)
	if e != nil {
		ie = &ierror{m: "Could not read product name", e: e}
		return
	}

	p, ie := selectProductByName(db, name)
	if ie != nil {
		return
	}

	t := today(db)
	_, e = db.Exec(`insert into dayProducts (dayId, productId) values ($1, $2)`, t.id, p.id)
	if e != nil {
		ie = &ierror{m: "Could not create today product", e: e}
		return
	}

	return
}
