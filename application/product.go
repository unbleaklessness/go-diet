package main

import (
	"database/sql"
	"fmt"
)

type product struct {
	id       int
	name     string
	kcals    float32
	proteins float32
	carbs    float32
	fats     float32
}

func scanProduct() (p product, e error) {

	fmt.Print("Name: ")
	_, e = fmt.Scanln(&p.name)
	if e != nil {
		return
	}

	fmt.Print("Kcals: ")
	_, e = fmt.Scanln(&p.kcals)
	if e != nil {
		return
	}

	fmt.Print("Proteins: ")
	_, e = fmt.Scanln(&p.proteins)
	if e != nil {
		return
	}

	fmt.Print("Carbs: ")
	_, e = fmt.Scanln(&p.carbs)
	if e != nil {
		return
	}

	fmt.Print("Fats: ")
	_, e = fmt.Scanln(&p.fats)
	if e != nil {
		return
	}

	return
}

func selectProducts(db *sql.DB) (products []product, ie *ierror) {
	exit := func(e error) {
		ie = &ierror{m: "Could not fetch products from the database", e: e}
	}

	rows, e := db.Query(`select id, name, kcals, proteins, carbs, fats from products`)
	if e != nil {
		exit(e)
		return
	}
	defer rows.Close()

	for rows.Next() {
		p := product{}
		e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			exit(e)
			return
		}
		products = append(products, p)
	}
	return
}

func selectProductByName(db *sql.DB, name string) (p product, ie *ierror) {
	exit := func(e error) {
		ie = &ierror{m: "Could not find product by name", e: e}
	}

	rows, e := db.Query(`select
		id, name, kcals, proteins, carbs, fats
		from products
		where name = $1
	`, name)
	if e != nil {
		exit(e)
		return
	}
	defer rows.Close()

	if rows.Next() {
		e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			exit(e)
			return
		}
		return
	}

	exit(nil)
	return
}
