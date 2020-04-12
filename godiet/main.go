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
)

type product struct {
	Name     string
	Kcals    float64
	Proteins float64
	Carbs    float64
	Fats     float64
	Portion  float64
	Maximum  float64
}

type productWithAmount struct {
	Amount  float64
	Product product
}

const (
	jsonExtension = ".json"

	normKcals    = 3000.0
	normProteins = 500.0
	normCarbs    = 1600.0
	normFats     = 900.0

	overshootPercentage = 15.0
	overshootKcals      = normKcals + (normKcals/100.0)*overshootPercentage
	overshootProteins   = normProteins + (normProteins/100.0)*overshootPercentage
	overshootCarbs      = normCarbs + (normCarbs/100.0)*overshootPercentage
	overshootFats       = normFats + (normFats/100.0)*overshootPercentage
)

func overshootAmount(amount float64) float64 {
	return amount + (amount/100.0)*overshootPercentage
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

	return p, e
}

func findProducts() []product {

	products := make([]product, 0)

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

func randomProduct(products []product) (product, int) {
	if len(products) < 1 {
		return product{}, -1
	}

	index := rand.Intn(len(products))
	return products[index], index
}

func (p *product) randomAmount() float64 {
	maximumPortions := int(math.Ceil(p.Maximum / p.Portion))
	return float64(rand.Intn(maximumPortions)) * p.Portion
}

func main() {

	rand.Seed(time.Now().UnixNano())

	defaultFloat := float64(math.MaxFloat64)
	// defaultInteger := int64(math.MaxInt64)

	newProduct := flag.String("new-product", "", "Create new product")
	kcals := flag.Float64("kcals", defaultFloat, "Kcals for product")
	proteins := flag.Float64("proteins", defaultFloat, "Proteins for product")
	carbs := flag.Float64("carbs", defaultFloat, "Carbs for product")
	fats := flag.Float64("fats", defaultFloat, "Fats for product")
	portion := flag.Float64("portion", defaultFloat, "Product portion")
	productMaximum := flag.Float64("maximum", defaultFloat, "Product maximum")
	optimize := flag.String("optimize", "", "Create an optimized diet")
	// dayProducts := flag.Int64("-day-products", defaultInteger, "Use with `-optimize` to set maximum number of products per day")

	flag.Parse()

	products := findProducts()

	fmt.Println(products)

	if len(*newProduct) > 0 && *kcals != defaultFloat &&
		*proteins != defaultFloat && *carbs != defaultFloat && *fats != defaultFloat &&
		*portion != defaultFloat && *productMaximum != defaultFloat {

		path := setJSONExtension(filepath.Clean(*newProduct))
		name := cutExtension(filepath.Base(*newProduct))

		p := product{
			Name:     name,
			Kcals:    *kcals,
			Proteins: *proteins,
			Carbs:    *carbs,
			Fats:     *fats,
			Portion:  *portion,
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

		amounts := make([]float64, len(products))

		kcalsCounter := 0.0
		proteinsCounter := 0.0
		carbsCounter := 0.0
		fatsCounter := 0.0

		for {
			p, i := randomProduct(products)
			if i == -1 {
				fmt.Println("No products are found")
				return
			}

			amount := p.randomAmount()

			kcalsCounter += p.Kcals * amount / 100.0
			proteinsCounter += p.Proteins * amount / 100.0
			carbsCounter += p.Carbs * amount / 100.0
			fatsCounter += p.Fats * amount / 100.0

			amounts[i] += amount

			if amounts[i] > overshootAmount(p.Maximum) ||
				kcalsCounter > overshootKcals || proteinsCounter > overshootProteins ||
				carbsCounter > overshootCarbs || fatsCounter > overshootFats {

				amounts = make([]float64, len(products))

				kcalsCounter = 0
				proteinsCounter = 0
				carbsCounter = 0
				fatsCounter = 0

				continue
			}

			if kcalsCounter > normKcals && proteinsCounter > normProteins &&
				carbsCounter > normCarbs && fatsCounter > normFats {

				break
			}
		}

		fmt.Println("Kcals:", kcalsCounter)
		fmt.Println("Proteins:", proteinsCounter)
		fmt.Println("Carbs:", carbsCounter)
		fmt.Println("Fats:", fatsCounter)
		fmt.Println()

		optimizedProducts := make([]productWithAmount, 0)

		for i, amount := range amounts {

			if amount <= 0.5 {
				continue
			}

			p := productWithAmount{
				Amount:  amount,
				Product: products[i],
			}
			optimizedProducts = append(optimizedProducts, p)

			fmt.Printf("%s -> %.2f\n", products[i].Name, amount)
		}

		data, e := json.MarshalIndent(optimizedProducts, "", "    ")
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
