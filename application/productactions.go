package main

import (
	"database/sql"
	"fmt"
)

func addProduct(db *sql.DB) (ie *ierror) {
	p, e := scanProduct()
	if e != nil {
		ie = &ierror{m: "Could not add product", e: e}
		return
	}

	db.Exec(`insert into products
		(name, kcals, proteins, carbs, fats)
		values
		($1, $2, $3, $4, $5)
	`, p.name, p.kcals, p.proteins, p.carbs, p.fats)

	return
}

func listProducts(db *sql.DB) (ie *ierror) {
	ps, e := selectProducts(db)
	if e != nil {
		ie = &ierror{m: "Could not list products", e: e}
		return
	}
	n := len(ps) - 1
	for i, p := range ps {
		fmt.Printf("Name: %s \n", p.name)
		fmt.Printf("Kcals: %f \n", p.kcals)
		fmt.Printf("Proteins: %f \n", p.proteins)
		fmt.Printf("Carbs: %f \n", p.carbs)
		fmt.Printf("Fats: %f \n", p.fats)
		if i != n {
			fmt.Println()
		}
	}
	return
}
