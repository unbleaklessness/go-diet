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
	Name     string
	Kcals    float64
	Proteins float64
	Carbs    float64
	Fats     float64
	Maximum  float64
	Minimum  float64
}

type productWithAmount struct {
	Amount  float64
	Product product
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

func main() {

	rand.Seed(time.Now().UnixNano())

	defaultFloat := float64(math.MaxFloat64)

	newProduct := flag.String("new-product", "", "Create new product")
	kcals := flag.Float64("kcals", defaultFloat, "Kcals for product")
	proteins := flag.Float64("proteins", defaultFloat, "Proteins for product")
	carbs := flag.Float64("carbs", defaultFloat, "Carbs for product")
	fats := flag.Float64("fats", defaultFloat, "Fats for product")
	productMaximum := flag.Float64("maximum", defaultFloat, "Product maximum")
	optimize := flag.String("optimize", "", "Create an optimized diet")

	flag.Parse()

	products := findProducts()

	if len(*newProduct) > 0 && *kcals != defaultFloat &&
		*proteins != defaultFloat && *carbs != defaultFloat && *fats != defaultFloat &&
		*productMaximum != defaultFloat {

		path := setJSONExtension(filepath.Clean(*newProduct))
		name := cutExtension(filepath.Base(*newProduct))

		p := product{
			Name:     name,
			Kcals:    *kcals,
			Proteins: *proteins,
			Carbs:    *carbs,
			Fats:     *fats,
			Maximum:  *productMaximum,
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

	} else if len(*optimize) > 0 {

		*optimize = filepath.Clean(*optimize)

		nWeekDays := 7
		nWeekProducts := 15
		nDayProducts := 6

		weekDiet := make([][]productWithAmount, nWeekDays)
		weekProducts := pickRandomProducts(products, nWeekProducts)

	outer:
		for {
			weekDiet = make([][]productWithAmount, nWeekDays)
			weekProducts = pickRandomProducts(products, nWeekProducts)

			currentDay := 0
			iterations := 0
			for currentDay < nWeekDays {
				if iterations > 10000 {
					continue outer
				}
				iterations++

				dayProducts := pickRandomProducts(weekProducts, nDayProducts)

				dayDiet, ok := dayDiet(dayProducts)
				if !ok || !dayDietIsValid(dayDiet) {
					continue
				}

				weekDiet[currentDay] = dayDiet
				currentDay++
			}

			break
		}

		data, e := json.MarshalIndent(weekDiet, "", "    ")
		if e != nil {
			fmt.Println("Could not save optimized diet")
			return
		}

		e = ioutil.WriteFile(*optimize, data, os.ModePerm)
		if e != nil {
			fmt.Println("Could not save optimized diet")
			return
		}

		return

	} else {
		fmt.Println("Unknown flag combination")
	}
}
