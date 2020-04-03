package main

import (
	"database/sql"
	"fmt"
)

func addProduct(db *sql.DB) ierrori {

	var (
		p         product
		e         error
		thisError func(e error) ierrori
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not add product", e: e}
	}

	p, e = scanProduct()
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`insert into products
		(name, kcals, proteins, carbs, fats)
		values
		($1, $2, $3, $4, $5)
	`, p.name, p.kcals, p.proteins, p.carbs, p.fats)
	if e != nil {
		return thisError(e)
	}

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
		fmt.Println("Name:", p.name)
		fmt.Println("Kcals:", p.kcals)
		fmt.Println("Proteins:", p.proteins)
		fmt.Println("Carbs:", p.carbs)
		fmt.Println("Fats:", p.fats)
		if i != n {
			fmt.Println()
		}
	}

	return nil
}

func removeProduct(db *sql.DB) ierrori {

	var (
		name      string
		e         error
		thisError func(e error) ierrori
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not remove product", e: e}
	}

	fmt.Print("Name: ")
	name, e = readLine()
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`delete from dayProducts
		where productId in
		(select id from products where name = $1 limit 1)`, name)
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`delete from products where name = $1`, name)
	if e != nil {
		return thisError(e)
	}

	return nil
}

func findProduct(db *sql.DB) ierrori {

	var (
		p         product
		e         error
		name      string
		thisError func(e error) ierrori
		rows      *sql.Rows
		pattern   string
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not find product", e: e}
	}

	fmt.Print("Name: ")
	name, e = readLine()
	if e != nil {
		return thisError(e)
	}

	pattern = "%" + name + "%"

	rows, e = db.Query(`select id, name, kcals, proteins, carbs, fats
		from products
		where name like $1 collate nocase`, pattern)
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	for rows.Next() {
		e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			return thisError(e)
		}
		fmt.Println()
		printProduct(p)
	}

	return nil
}
