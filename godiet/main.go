package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unbleaklessness/go-diet/simplex"
)

type product struct {
	ID          uint64
	Kcals       float64
	Proteins    float64
	Carbs       float64
	Fats        float64
	Maximum     float64
	Minimum     float64
	Description string

	name string
}

type dietEntry struct {
	ID       uint64
	Amount   float64
	Consumed float64

	product product
}

type diet = [][]dietEntry

const (
	jsonExtension = ".json"

	normKcals    = 3000.0
	normProteins = 500.0
	normCarbs    = 1600.0
	normFats     = 900.0

	overshootPercentage = 5.0
	overshootKcals      = normKcals + (normKcals/100.0)*overshootPercentage
	overshootProteins   = normProteins + (normProteins/100.0)*overshootPercentage
	overshootCarbs      = normCarbs + (normCarbs/100.0)*overshootPercentage
	overshootFats       = normFats + (normFats/100.0)*overshootPercentage
)

func id() uint64 {
	return rand.Uint64()
}

func cutExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

func setJSONExtension(path string) string {
	return cutExtension(path) + jsonExtension
}

func unmarshalProduct(productBytes []byte) (product, error) {

	decoder := json.NewDecoder(bytes.NewReader(productBytes))
	decoder.DisallowUnknownFields()

	p := product{}

	e := decoder.Decode(&p)
	if e != nil {
		return product{}, e
	}

	return p, nil
}

func readProduct(path string) (product, error) {

	data, e := ioutil.ReadFile(path)
	if e != nil {
		return product{}, e
	}

	p, e := unmarshalProduct(data)
	if e != nil {
		return product{}, e
	}

	p.name = cutExtension(filepath.Base(path))

	return p, e
}

func isProduct(info os.FileInfo) bool {
	return !info.IsDir() && filepath.Ext(info.Name()) == jsonExtension
}

func unmarshalDiet(dietBytes []byte) (diet, error) {

	decoder := json.NewDecoder(bytes.NewReader(dietBytes))
	decoder.DisallowUnknownFields()

	d := diet{}

	e := decoder.Decode(&d)
	if e != nil {
		return diet{}, e
	}

	return d, nil
}

func readDiet(path string) (diet, error) {

	data, e := ioutil.ReadFile(path)
	if e != nil {
		return diet{}, e
	}

	d, e := unmarshalDiet(data)
	if e != nil {
		return diet{}, e
	}

	return d, e
}

func isDiet(path string) bool {
	info, e := os.Stat(path)
	return e == nil && !info.IsDir() && filepath.Ext(path) == jsonExtension
}

func findProducts() []product {

	products := []product{}

	filepath.Walk(".", func(path string, info os.FileInfo, e error) error {

		if !isProduct(info) {
			return nil
		}

		p, e := readProduct(path)
		if e != nil {
			return nil
		}

		products = append(products, p)

		return nil
	})

	return products
}

func dayDiet(products []product) ([]dietEntry, bool) {

	nOptimizationColumns := len(products)

	nLTConstraints := 4 + len(products)
	ltConstraintsLHS := make([][]float64, nLTConstraints)
	ltConstraintsRHS := make([]float64, nLTConstraints)
	for i := 0; i < nLTConstraints; i++ {
		ltConstraintsLHS[i] = make([]float64, nOptimizationColumns)
	}

	for i := 0; i < len(products); i++ {
		ltConstraintsLHS[0][i] = products[i].Kcals
		ltConstraintsLHS[1][i] = products[i].Proteins * 4.0
		ltConstraintsLHS[2][i] = products[i].Carbs * 4.0
		ltConstraintsLHS[3][i] = products[i].Fats * 9.0
	}
	ltConstraintsRHS[0] = overshootKcals
	ltConstraintsRHS[1] = overshootProteins
	ltConstraintsRHS[2] = overshootCarbs
	ltConstraintsRHS[3] = overshootFats

	for i := 0; i < len(products); i++ {
		j := i + 4
		ltConstraintsLHS[j][i] = 100.0
		ltConstraintsRHS[j] = products[i].Maximum
	}

	nGTConstraints := 4 + len(products)
	gtConstraintsLHS := make([][]float64, nGTConstraints)
	gtConstraintsRHS := make([]float64, nGTConstraints)
	for i := 0; i < nGTConstraints; i++ {
		gtConstraintsLHS[i] = make([]float64, nOptimizationColumns)
	}

	for i := 0; i < len(products); i++ {
		gtConstraintsLHS[0][i] = products[i].Kcals
		gtConstraintsLHS[1][i] = products[i].Proteins * 4.0
		gtConstraintsLHS[2][i] = products[i].Carbs * 4.0
		gtConstraintsLHS[3][i] = products[i].Fats * 9.0
	}
	gtConstraintsRHS[0] = normKcals
	gtConstraintsRHS[1] = normProteins
	gtConstraintsRHS[2] = normCarbs
	gtConstraintsRHS[3] = normFats

	for i := 0; i < len(products); i++ {
		j := i + 4
		gtConstraintsLHS[j][i] = 100.0
		gtConstraintsRHS[j] = products[i].Minimum
	}

	objective := make([]float64, nOptimizationColumns)
	for i := range products {
		proteins := products[i].Proteins * 4.0
		carbs := products[i].Carbs * 4.0
		fats := products[i].Fats * 9.0
		objective[i] = products[i].Kcals + proteins + carbs + fats
	}

	amounts, _, ok := simplex.Simplex(
		objective,
		gtConstraintsLHS, gtConstraintsRHS,
		ltConstraintsLHS, ltConstraintsRHS,
		[][]float64{}, []float64{})
	if !ok {
		return []dietEntry{}, false
	}

	totalKcals := 0.0
	totalProteins := 0.0
	totalCarbs := 0.0
	totalFats := 0.0

	for i, p := range products {
		if amounts[i] <= 0.0 {
			continue
		}
		totalKcals += p.Kcals * amounts[i]
		totalProteins += p.Proteins * amounts[i] * 4.0
		totalCarbs += p.Carbs * amounts[i] * 4.0
		totalFats += p.Fats * amounts[i] * 9.0
	}

	ok = totalKcals <= overshootKcals && totalKcals >= normKcals &&
		totalProteins <= overshootProteins && totalProteins >= normProteins &&
		totalCarbs <= overshootCarbs && totalCarbs >= normCarbs &&
		totalFats <= overshootFats && totalFats >= normFats
	if !ok {
		return []dietEntry{}, false
	}

	dayDiet := make([]dietEntry, len(products))

	for i, amount := range amounts {
		p := dietEntry{
			ID:     products[i].ID,
			Amount: amount,
		}
		dayDiet[i] = p
	}

	return dayDiet, true
}

func pickRandomProducts(products []product, n int) []product {

	pickedProducts := make([]product, n)
	pickedIndexes := make([]int, n)
	index := 0

outer:
	for index < n {
		pickIndex := rand.Intn(len(products))
		for _, i := range pickedIndexes[:index] {
			if i == pickIndex {
				continue outer
			}
		}
		pickedProducts[index] = products[pickIndex]
		pickedIndexes[index] = pickIndex
		index++
	}

	return pickedProducts
}

func weekDay() int {
	day := int(time.Now().Weekday())
	if day == 0 {
		day = 6
	} else {
		day--
	}
	return day
}

func writeJSON(structure interface{}, path string) error {

	data, e := json.MarshalIndent(structure, "", "    ")
	if e != nil {
		return e
	}

	e = ioutil.WriteFile(path, data, os.ModePerm)
	if e != nil {
		return e
	}

	return nil
}

func productWithID(products []product, id uint64) (product, bool) {

	for _, p := range products {
		if p.ID == id {
			return p, true
		}
	}

	return product{}, false
}

func setDietProducts(d diet, products []product) (diet, bool) {

	for _, day := range d {
		for i, entry := range day {
			p, ok := productWithID(products, entry.ID)
			if !ok {
				return diet{}, false
			}
			day[i].product = p
		}
	}

	return d, true
}

func getDiet(path string, products []product) (diet, bool) {

	if !isDiet(path) {
		fmt.Println("Provided file is not a diet")
		return diet{}, false
	}

	d, e := readDiet(path)
	if e != nil {
		fmt.Println("Could not read diet")
		return diet{}, false
	}

	d, ok := setDietProducts(d, products)
	if !ok {
		fmt.Println("Could not find diet product")
		return diet{}, false
	}

	return d, true
}

func main() {

	rand.Seed(time.Now().UnixNano())

	defaultFloat := float64(math.MaxFloat64)
	defaultInteger := int64(math.MaxInt64)

	newProductFlag := flag.String("new-product", "", "Create new product")
	newDietFlag := flag.String("new-diet", "", "Create optimized diet")
	productsPerDayFlag := flag.Int64("products-per-day", defaultInteger, "Use with `-optimize` to set the number of products per day")
	productsPerWeekFlag := flag.Int64("products-per-week", defaultInteger, "Use with `-optimize` to set the number of products per week")
	dietFlag := flag.String("diet", "", "Diet actions. Show diet if alone")
	productsFlag := flag.Bool("products", false, "Use with `-diet` flag to see products and amounts for the whole week")
	remainingFlag := flag.Bool("remaining", false, "Use with `-diet` flag to see remaining products and amounts for today")
	productFlag := flag.String("product", "", "Use with `-diet` and `-consumed` flags to add a product and consumed amount for today")
	consumedFlag := flag.Float64("consumed", defaultFloat, "Use with `-diet` and `-product` flags to add a product and consumed amount for today")
	resetConsumedFlag := flag.Bool("reset-consumed", false, "Use with `-diet` flag to reset all consumed amounts")

	flag.Parse()

	products := findProducts()

	if len(*newProductFlag) > 0 {

		path := setJSONExtension(filepath.Clean(*newProductFlag))

		p := product{
			ID: id(),
		}

		e := writeJSON(p, path)
		if e != nil {
			fmt.Println("Could not create new product")
			return
		}

		return

	} else if len(*newDietFlag) > 0 {

		if len(products) < 1 {
			fmt.Println("No products found")
			return
		}

		nWeekDays := 7

		productsPerDay := 6
		if *productsPerDayFlag != defaultInteger {
			productsPerDay = int(*productsPerDayFlag)
		}

		productsPerWeek := 15
		if *productsPerWeekFlag != defaultInteger {
			productsPerWeek = int(*productsPerWeekFlag)
		}

		newDiet := make(diet, nWeekDays)
		weekProducts := pickRandomProducts(products, productsPerWeek)

	newDietLoop:
		for {
			newDiet = make(diet, nWeekDays)
			weekProducts = pickRandomProducts(products, productsPerWeek)

			currentDay := 0
			iterations := 0
			for currentDay < nWeekDays {
				if iterations > 10000 {
					continue newDietLoop
				}
				iterations++

				dayProducts := pickRandomProducts(weekProducts, productsPerDay)

				dayDiet, ok := dayDiet(dayProducts)
				if !ok {
					continue
				}

				newDiet[currentDay] = dayDiet
				currentDay++
			}

			break
		}

		path := setJSONExtension(filepath.Clean(*newDietFlag))

		e := writeJSON(newDiet, path)
		if e != nil {
			fmt.Println("Could not save diet")
			return
		}

		return

	} else if len(*dietFlag) > 0 && *productsFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		commonEntries := []dietEntry{}

		for _, day := range diet {
		dietProductsLoop:
			for _, entry := range day {
				for i, common := range commonEntries {
					if common.product.name == entry.product.name {
						commonEntries[i].Amount += entry.Amount
						continue dietProductsLoop
					}
				}
				commonEntries = append(commonEntries, entry)
			}
		}

		for i, p := range commonEntries {
			amount := p.Amount * 100.0
			index := i + 1
			fmt.Printf("%d) %s - %.0f\n", index, p.product.name, amount)
		}

		return

	} else if len(*dietFlag) > 0 && *remainingFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		today := weekDay()

		for i, entry := range diet[today] {
			index := i + 1
			amount := entry.Amount * 100.0
			remaining := (entry.Amount - entry.Consumed) * 100.0
			percentange := (amount - remaining) / amount * 100.0
			fmt.Printf("%d) %s - %.0f%%, %.0f out of %.0f\n", index, entry.product.name, percentange, remaining, amount)
		}

		return

	} else if len(*dietFlag) > 0 && len(*productFlag) > 0 && *consumedFlag != defaultFloat {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		today := weekDay()

		for i, entry := range diet[today] {
			if entry.product.name == *productFlag {

				diet[today][i].Consumed += *consumedFlag / 100.0

				e := writeJSON(diet, *dietFlag)
				if e != nil {
					fmt.Println("Could not save diet with changes")
					return
				}

				return
			}
		}

		fmt.Println("Could not find product with provided name")
		return

	} else if len(*dietFlag) > 0 && *resetConsumedFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		for i := range diet {
			for j := range diet[i] {
				diet[i][j].Consumed = 0.0
			}
		}

		e := writeJSON(diet, *dietFlag)
		if e != nil {
			fmt.Println("Could not save diet with changes")
			return
		}

		return

	} else if len(*dietFlag) > 0 {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		for i, day := range diet {
			index := i + 1
			fmt.Printf("Day %d:\n", index)
			for j, entry := range day {
				index := j + 1
				amount := entry.Amount * 100.0
				fmt.Printf("%d) %s - %.0f\n", index, entry.product.name, amount)
			}
			if i < len(day) {
				fmt.Println()
			}
		}

		return

	} else {
		fmt.Println("Unknown flag combination")
	}
}
