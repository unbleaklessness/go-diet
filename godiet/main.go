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
	Name        string
	Kcals       float64
	Proteins    float64
	Carbs       float64
	Fats        float64
	Maximum     float64
	Minimum     float64
	Description string
}

type productWithAmount struct {
	Amount   float64
	Consumed float64
	Product  product
}

type diet = [][]productWithAmount

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

	return p, e
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

func findProducts() []product {

	products := []product{}

	filepath.Walk(".", func(path string, info os.FileInfo, e error) error {

		if info.IsDir() {
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

func dayDiet(products []product) ([]productWithAmount, bool) {

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
		return []productWithAmount{}, false
	}

	diet := make([]productWithAmount, len(products))

	for i, amount := range amounts {
		p := productWithAmount{
			Product: products[i],
			Amount:  amount,
		}
		diet[i] = p
	}

	return diet, true
}

func dayDietIsValid(products []productWithAmount) bool {

	totalKcals := 0.0
	totalProteins := 0.0
	totalCarbs := 0.0
	totalFats := 0.0

	for _, p := range products {
		if p.Amount <= 0.0 {
			continue
		}
		totalKcals += p.Product.Kcals * p.Amount
		totalProteins += p.Product.Proteins * p.Amount * 4.0
		totalCarbs += p.Product.Carbs * p.Amount * 4.0
		totalFats += p.Product.Fats * p.Amount * 9.0
	}

	return totalKcals <= overshootKcals && totalKcals >= normKcals &&
		totalProteins <= overshootProteins && totalProteins >= normProteins &&
		totalCarbs <= overshootCarbs && totalCarbs >= normCarbs &&
		totalFats <= overshootFats && totalFats >= normFats
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
		pickedIndexes[index] = pickIndex
		pickedProducts[index] = products[pickIndex]
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
		name := cutExtension(filepath.Base(path))

		p := product{
			Name: name,
		}

		data, e := json.MarshalIndent(p, "", "    ")
		if e != nil {
			fmt.Println("Could not create new product")
			return
		}

		e = ioutil.WriteFile(path, data, os.ModePerm)
		if e != nil {
			fmt.Println("Could not write new product")
			return
		}

		return

	} else if len(*newDietFlag) > 0 {

		*newDietFlag = filepath.Clean(*newDietFlag)

		nWeekDays := 7

		nDayProducts := 6
		if *productsPerDayFlag != defaultInteger {
			nDayProducts = int(*productsPerDayFlag)
		}

		nWeekProducts := 15
		if *productsPerWeekFlag != defaultInteger {
			nWeekProducts = int(*productsPerWeekFlag)
		}

		newDiet := make(diet, nWeekDays)
		weekProducts := pickRandomProducts(products, nWeekProducts)

	dietLoop:
		for {
			newDiet = make(diet, nWeekDays)
			weekProducts = pickRandomProducts(products, nWeekProducts)

			currentDay := 0
			iterations := 0
			for currentDay < nWeekDays {
				if iterations > 10000 {
					continue dietLoop
				}
				iterations++

				dayProducts := pickRandomProducts(weekProducts, nDayProducts)

				dayDiet, ok := dayDiet(dayProducts)
				if !ok || !dayDietIsValid(dayDiet) {
					continue
				}

				newDiet[currentDay] = dayDiet
				currentDay++
			}

			break
		}

		e := writeJSON(newDiet, *newDietFlag)
		if e != nil {
			fmt.Println("Could not save diet")
			return
		}

		return

	} else if len(*dietFlag) > 0 && *productsFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, e := readDiet(*dietFlag)
		if e != nil {
			fmt.Println("Could not read diet")
			return
		}

		dietProducts := []productWithAmount{}

		for _, day := range diet {
		dayProductLoop:
			for _, dayProduct := range day {
				for i, dietProduct := range dietProducts {
					if dietProduct.Product.Name == dayProduct.Product.Name {
						dietProducts[i].Amount += dayProduct.Amount
						continue dayProductLoop
					}
				}
				dietProducts = append(dietProducts, dayProduct)
			}
		}

		for i, p := range dietProducts {
			pAmount := p.Amount * 100.0
			index := i + 1
			fmt.Printf("%d) %s - %.0f\n", index, p.Product.Name, pAmount)
		}

		return

	} else if len(*dietFlag) > 0 && *remainingFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, e := readDiet(*dietFlag)
		if e != nil {
			fmt.Println("Could not read diet")
			return
		}

		today := weekDay()

		maximum := 0
		for _, p := range diet[today] {
			length := len(p.Product.Name)
			if length > maximum {
				maximum = length
			}
		}

		for i, p := range diet[today] {
			index := i + 1
			amount := p.Amount * 100.0
			remaining := (p.Amount - p.Consumed) * 100.0
			percentange := (amount - remaining) / amount * 100.0
			fmt.Printf("%d) %s - %.0f%%, %.0f out of %.0f\n", index, p.Product.Name, percentange, remaining, amount)
		}

		return

	} else if len(*dietFlag) > 0 && len(*productFlag) > 0 && *consumedFlag != defaultFloat {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, e := readDiet(*dietFlag)
		if e != nil {
			fmt.Println("Could not read diet")
			return
		}

		today := weekDay()

		ok := false
		for i, p := range diet[today] {
			if p.Product.Name == *productFlag {
				diet[today][i].Consumed += *consumedFlag / 100.0
				ok = true
				break
			}
		}
		if !ok {
			fmt.Println("Could not find product with provided name")
			return
		}

		e = writeJSON(diet, *dietFlag)
		if e != nil {
			fmt.Println("Could not save diet with changes")
			return
		}

		return

	} else if len(*dietFlag) > 0 && *resetConsumedFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, e := readDiet(*dietFlag)
		if e != nil {
			fmt.Println("Could not read diet")
			return
		}

		for i := range diet {
			for j := range diet[i] {
				diet[i][j].Consumed = 0.0
			}
		}

		e = writeJSON(diet, *dietFlag)
		if e != nil {
			fmt.Println("Could not save diet with changes")
			return
		}

		return

	} else if len(*dietFlag) > 0 {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, e := readDiet(*dietFlag)
		if e != nil {
			fmt.Println("Could not read diet")
			return
		}

		for i, day := range diet {
			index := i + 1
			fmt.Printf("Day %d:\n", index)
			for j, p := range day {
				index := j + 1
				amount := p.Amount * 100.0
				fmt.Printf("%d) %s - %.0f\n", index, p.Product.Name, amount)
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
