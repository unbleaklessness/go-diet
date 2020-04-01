package main

import (
	"database/sql"
	"fmt"
)

func addProduct(db *sql.DB) ierrori {

	var (
		p product
		e error
	)

	p, e = scanProduct()
	if e != nil {
		return ierror{m: "Could not add product", e: e}
	}

	db.Exec(`insert into products
		(name, kcals, proteins, carbs, fats)
		values
		($1, $2, $3, $4, $5)
	`, p.name, p.kcals, p.proteins, p.carbs, p.fats)

	return nil
}

func listProducts(db *sql.DB) ierrori {

	var (
		e         error
		rows      *sql.Rows
		n         int
		i         int
		thisError func(e error) ierrori
		p         product
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not list products", e: e}
	}

	rows, e = db.Query(`select id, name, kcals, proteins, carbs, fats from products`)
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	var products []product
	for rows.Next() {
		e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			return thisError(e)
		}
		products = append(products, p)
	}

	n = len(products) - 1
	for i, p = range products {
		fmt.Printf("Name: %s \n", p.name)
		fmt.Printf("Kcals: %f \n", p.kcals)
		fmt.Printf("Proteins: %f \n", p.proteins)
		fmt.Printf("Carbs: %f \n", p.carbs)
		fmt.Printf("Fats: %f \n", p.fats)
		if i != n {
			fmt.Println()
		}
	}

	return nil
}
