package main

import "flag"

type flags struct {
	product bool
	add     bool
	name    string
	list    bool
	today   bool
	norm    bool
	remove  bool
	total   bool
	rest    []string
}

func initializeFlags() flags {

	var (
		f flags

		productFlagUsage string
		productLong      *bool
		productShort     *bool

		addFlagUsage string
		addLong      *bool
		addShort     *bool

		nameFlagUsage string
		nameLong      *string
		nameShort     *string

		listFlagUsage string
		listLong      *bool
		listShort     *bool

		todayFlagUsage string
		todayLong      *bool
		todayShort     *bool

		normUsage string
		norm      *bool

		removeFlagUsage string
		removeLong      *bool
		removeShort     *bool

		totalFlagUsage string
		total          *bool

		product bool
		add     bool
		list    bool
		today   bool
		remove  bool
	)

	productFlagUsage = "Action with products"
	productLong = flag.Bool("product", false, productFlagUsage)
	productShort = flag.Bool("p", false, productFlagUsage)

	addFlagUsage = "Add an item"
	addLong = flag.Bool("add", false, addFlagUsage)
	addShort = flag.Bool("a", false, addFlagUsage)

	nameFlagUsage = "Name of an item"
	nameLong = flag.String("name", "", nameFlagUsage)
	nameShort = flag.String("n", "", nameFlagUsage)

	listFlagUsage = "List items"
	listLong = flag.Bool("list", false, listFlagUsage)
	listShort = flag.Bool("l", false, listFlagUsage)

	todayFlagUsage = "Action with today"
	todayLong = flag.Bool("today", false, todayFlagUsage)
	todayShort = flag.Bool("t", false, todayFlagUsage)

	normUsage = "Action with daily norm"
	norm = flag.Bool("norm", false, normUsage)

	removeFlagUsage = "Remove an item"
	removeLong = flag.Bool("remove", false, removeFlagUsage)
	removeShort = flag.Bool("r", false, removeFlagUsage)

	totalFlagUsage = "Show total"
	total = flag.Bool("total", false, totalFlagUsage)

	flag.Parse()

	product = *productLong || *productShort
	add = *addLong || *addShort
	list = *listLong || *listShort
	today = *todayLong || *todayShort
	remove = *removeLong || *removeShort

	var name string
	if len(*nameLong) > 0 {
		name = *nameLong
	} else if len(*nameShort) > 0 {
		name = *nameShort
	}

	f.product = product
	f.add = add
	f.name = name
	f.list = list
	f.today = today
	f.norm = *norm
	f.remove = remove
	f.total = *total
	f.rest = flag.Args()

	return f
}
