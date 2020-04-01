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

func showTodayTotal(db *sql.DB) ierrori {

	var (
		t            day
		e            error
		products     []product
		p            product
		thisError    func(e error) ierrori
		rows         *sql.Rows
		totalProduct product
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not show total for today", e: e}
	}

	t, e = today(db)
	if e != nil {
		return thisError(e)
	}

	rows, e = db.Query(`select kcals, proteins, carbs, fats
		from products where id in
		(select productId from dayProducts
			where dayId = $1)`, t.id)
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	for rows.Next() {
		e = rows.Scan(&p.kcals, &p.proteins, &p.carbs, &p.fats)
		if e != nil {
			return thisError(e)
		}
		products = append(products, p)
	}

	for _, p = range products {
		totalProduct.kcals += p.kcals
		totalProduct.proteins += p.proteins
		totalProduct.carbs += p.carbs
		totalProduct.fats += p.fats
	}

	fmt.Println("Kcals:", totalProduct.kcals)
	fmt.Println("Proteins:", totalProduct.proteins)
	fmt.Println("Carbs:", totalProduct.carbs)
	fmt.Println("Fats:", totalProduct.fats)

	return nil
}

func listTodayProducts(db *sql.DB) ierrori {

	var (
		t            day
		e            error
		productNames []string
		name         string
		thisError    func(e error) ierrori
		rows         *sql.Rows
		i            int
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not list today products", e: e}
	}

	t, e = today(db)
	if e != nil {
		return thisError(e)
	}

	rows, e = db.Query(`select name
		from products where id in
		(select productId from dayProducts
			where dayId = $1)`, t.id)
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	for rows.Next() {
		e = rows.Scan(&name)
		if e != nil {
			return thisError(e)
		}
		productNames = append(productNames, name)
	}

	for i, name = range productNames {
		fmt.Println(i+1, "-", name)
	}

	return nil
}
