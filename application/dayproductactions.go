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
		amount    float32
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

	fmt.Print("Amount: ")
	_, e = fmt.Scanln(&amount)
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

	if !rows.Next() {
		rows.Close()
		return thisError(nil)
	}

	e = rows.Scan(&p.id, &p.name, &p.kcals, &p.proteins, &p.carbs, &p.fats)
	if e != nil {
		return thisError(e)
	}
	rows.Close()

	d, e = today(db)
	if e != nil {
		return thisError(e)
	}

	_, e = db.Exec(`insert into dayProducts (dayId, productId, amount) values ($1, $2, $3)`, d.id, p.id, amount)
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
		d         day
		e         error
		products  []product
		amounts   []float32
		amount    float32
		p         product
		thisError func(e error) ierrori
		rows      *sql.Rows
		total     product
		norm      dailyNorm
		i         int
	)

	thisError = func(e error) ierrori {
		return ierror{m: "Could not show total for today", e: e}
	}

	d, e = today(db)
	if e != nil {
		return thisError(e)
	}

	rows, e = db.Query(`select p.kcals, p.proteins, p.carbs, p.fats, dp.amount
		from products p
		inner join dayProducts dp
		on dp.productId = p.id
		and (dp.dayId = $1)`, d.id)
	if e != nil {
		return thisError(e)
	}

	for rows.Next() {
		e = rows.Scan(&p.kcals, &p.proteins, &p.carbs, &p.fats, &amount)
		if e != nil {
			rows.Close()
			return thisError(e)
		}
		products = append(products, p)
		amounts = append(amounts, amount)
	}
	rows.Close()

	rows, e = db.Query(`select kcals, proteins, carbs, fats from dailyNorm limit 1`)
	if e != nil {
		return thisError(e)
	}
	defer rows.Close()

	if !rows.Next() {
		return thisError(nil)
	}

	e = rows.Scan(&norm.kcals, &norm.proteins, &norm.carbs, &norm.fats)
	if e != nil {
		return thisError(e)
	}

	if len(amounts) != len(products) {
		return thisError(nil)
	}

	for i, p = range products {
		total.kcals += (p.kcals / 100) * amounts[i]
		total.proteins += (p.proteins / 100) * amounts[i]
		total.carbs += (p.carbs / 100) * amounts[i]
		total.fats += (p.fats / 100) * amounts[i]
	}

	fmt.Printf("Kcals: %.2f, %.2f%% \n", total.kcals, (norm.kcals*100)/total.kcals)
	fmt.Printf("Proteins: %.2f, %.2f%% \n", total.proteins, (norm.proteins*100)/total.proteins)
	fmt.Printf("Carbs: %.2f, %.2f%% \n", total.carbs, (norm.carbs*100)/total.carbs)
	fmt.Printf("Fats: %.2f, %.2f%% \n", total.fats, (norm.fats*100)/total.fats)

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
