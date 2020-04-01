package main

import (
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

func scanProduct() (product, error) {

	var (
		p         product
		e         error
		thisError func(e error) (product, error)
	)

	thisError = func(e error) (product, error) {
		return p, ierror{m: "Could no scan product", e: e}
	}

	fmt.Print("Name: ")
	_, e = fmt.Scanln(&p.name)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Kcals: ")
	_, e = fmt.Scanln(&p.kcals)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Proteins: ")
	_, e = fmt.Scanln(&p.proteins)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Carbs: ")
	_, e = fmt.Scanln(&p.carbs)
	if e != nil {
		return thisError(e)
	}

	fmt.Print("Fats: ")
	_, e = fmt.Scanln(&p.fats)
	if e != nil {
		return thisError(e)
	}

	return p, nil
}
