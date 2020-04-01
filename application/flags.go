package main

import "flag"

type flags struct {
	product bool
	add     bool
	name    string
	list    bool
	today   bool
	norm    bool
	rest    []string
}

func initializeFlags() flags {

	productFlagUsage := "Action with products"
	productLong := flag.Bool("product", false, productFlagUsage)
	productShort := flag.Bool("p", false, productFlagUsage)

	addFlagUsage := "Add an item"
	addLong := flag.Bool("add", false, addFlagUsage)
	addShort := flag.Bool("a", false, addFlagUsage)

	nameFlagUsage := "Name of an item"
	nameLong := flag.String("name", "", nameFlagUsage)
	nameShort := flag.String("n", "", nameFlagUsage)

	listFlagUsage := "List items"
	listLong := flag.Bool("list", false, listFlagUsage)
	listShort := flag.Bool("l", false, listFlagUsage)

	todayFlagUsage := "Action with today"
	todayLong := flag.Bool("today", false, todayFlagUsage)
	todayShort := flag.Bool("t", false, todayFlagUsage)

	normUsage := "Action with daily norm"
	norm := flag.Bool("norm", false, normUsage)

	flag.Parse()

	product := *productLong || *productShort
	add := *addLong || *addShort
	list := *listLong || *listShort
	today := *todayLong || *todayShort

	var name string
	if len(*nameLong) > 0 {
		name = *nameLong
	} else if len(*nameShort) > 0 {
		name = *nameShort
	}

	return flags{
		product: product,
		add:     add,
		name:    name,
		list:    list,
		today:   today,
		norm:    *norm,
		rest:    flag.Args(),
	}
}
