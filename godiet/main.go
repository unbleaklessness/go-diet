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
	"sync/atomic"
	"time"

	"github.com/unbleaklessness/go-diet/simplex"
)

type product struct {
	ID uint64

	Maximum float64
	Minimum float64

	Description string

	Kcals    float64
	Proteins float64
	Carbs    float64
	Fats     float64

	VitaminA        float64
	Thiamin         float64
	Riboflavin      float64
	Niacin          float64
	PantothenicAcid float64
	VitaminB6       float64
	Folate          float64
	VitaminB12      float64
	VitaminC        float64
	VitaminD        float64
	VitaminE        float64
	VitaminK        float64

	Calcium    float64
	Magnesium  float64
	Phosphorus float64
	Potassium  float64
	Sodium     float64
	Copper     float64
	Iron       float64
	Manganese  float64
	Zinc       float64

	Omega3 float64
	Omega6 float64

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

	lowerKcals    = 3000.0
	lowerProteins = lowerKcals * 0.15
	lowerCarbs    = lowerKcals * 0.55
	lowerFats     = lowerKcals * 0.3

	upperPercentage = 0.1
	upperKcals      = lowerKcals + lowerKcals*upperPercentage
	upperProteins   = lowerProteins + lowerProteins*upperPercentage
	upperCarbs      = lowerCarbs + lowerCarbs*upperPercentage
	upperFats       = lowerFats + lowerFats*upperPercentage

	lowerVitaminA = 3000.0 // IU/Day.
	upperVitaminA = 7000.0

	lowerThiamin = 1.2                    // MG/Day.
	upperThiamin = lowerThiamin * 10000.0 // No upper bound.

	lowerRiboflavin = 1.3                       // MG/Day.
	upperRiboflavin = lowerRiboflavin * 10000.0 // No upper bound.

	lowerNiacin = 16.0 // MG/Day.
	upperNiacin = 35.0

	lowerPantothenicAcid = 5.0                            // MG/Day.
	upperPantothenicAcid = lowerPantothenicAcid * 10000.0 // No upper bound.

	lowerVitaminB6 = 1.3 // MG/Day.
	upperVitaminB6 = 100.0

	lowerFolate = 400.0 // MCG/Day.
	upperFolate = 800.0

	lowerVitaminB12 = 2.4   // MCG/Day.
	upperVitaminB12 = 600.0 // Clear upper bound is unkown.

	lowerVitaminC = 90.0   // MG/Day.
	upperVitaminC = 1500.0 // Upper bound is 2000.0.

	lowerVitaminD = 150.0 // IU/Day, should be 600.0.
	upperVitaminD = 4000.0

	lowerVitaminE = 5.0   // MG/Day, should be 15.0.
	upperVitaminE = 125.0 // Clear upper bound is unkown, somewhere around 150.0.

	lowerVitaminK = 120.0                   // MCG/Day.
	upperVitaminK = lowerVitaminK * 10000.0 // No upper bound.

	lowerCalcium = 1000.0 // MG/Day.
	upperCalcium = 2500.0

	lowerMagnesium = 420.0                    // MG/Day.
	upperMagnesium = lowerMagnesium * 10000.0 // Clear upper bound is unkown.

	lowerPhosphorus = 700.0 // MG/Day.
	upperPhosphorus = 4000.0

	lowerPotassium = 4700.0                   // MG/Day.
	upperPotassium = lowerPotassium * 10000.0 // Clear upper bound is unkown.

	lowerSodium = 1500.0 // MG/Day.
	upperSodium = 2300.0

	lowerCopper = 0.9 // MG/Day.
	upperCopper = 10.0

	lowerIron = 8.0 // MG/Day.
	upperIron = 45.0

	lowerManganese = 2.3 // MG/Day.
	upperManganese = 10.0

	lowerZinc = 11.0 // MG/Day.
	upperZinc = 40.0
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

func isJSONPath(path string) bool {
	info, e := os.Stat(path)
	return e == nil && !info.IsDir() && filepath.Ext(path) == jsonExtension
}

func isJSONInfo(info os.FileInfo) bool {
	return !info.IsDir() && filepath.Ext(info.Name()) == jsonExtension
}

func findProducts() []product {

	products := []product{}

	filepath.Walk(".", func(path string, info os.FileInfo, e error) error {

		if !isJSONInfo(info) {
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

func totalMacronutrients(entries []dietEntry, isConsumed bool) (float64, float64, float64, float64) {

	totalKcals := 0.0
	totalProteins := 0.0
	totalCarbs := 0.0
	totalFats := 0.0

	if isConsumed {
		for _, entry := range entries {
			totalKcals += entry.product.Kcals * entry.Consumed
			totalProteins += entry.product.Proteins * entry.Consumed * 4.0
			totalCarbs += entry.product.Carbs * entry.Consumed * 4.0
			totalFats += entry.product.Fats * entry.Consumed * 9.0
		}
	} else {
		for _, entry := range entries {
			totalKcals += entry.product.Kcals * entry.Amount
			totalProteins += entry.product.Proteins * entry.Amount * 4.0
			totalCarbs += entry.product.Carbs * entry.Amount * 4.0
			totalFats += entry.product.Fats * entry.Amount * 9.0
		}
	}

	return totalKcals, totalProteins, totalCarbs, totalFats
}

type micronutrient = int

const (
	vitaminA micronutrient = iota + 1
	thiamin
	riboflavin
	niacin
	pantothenicAcid
	vitaminB6
	folate
	vitaminB12
	vitaminC
	vitaminD
	vitaminE
	vitaminK
	calcium
	magnesium
	phosphorus
	potassium
	sodium
	copper
	iron
	manganese
	zinc
)

func totalMicronutrients(entries []dietEntry, isConsumed bool) map[micronutrient]float64 {

	micronutrients := make(map[micronutrient]float64, 0)
	micronutrients[vitaminA] = 0.0
	micronutrients[thiamin] = 0.0
	micronutrients[riboflavin] = 0.0
	micronutrients[niacin] = 0.0
	micronutrients[pantothenicAcid] = 0.0
	micronutrients[vitaminB6] = 0.0
	micronutrients[folate] = 0.0
	micronutrients[vitaminB12] = 0.0
	micronutrients[vitaminC] = 0.0
	micronutrients[vitaminD] = 0.0
	micronutrients[vitaminE] = 0.0
	micronutrients[vitaminK] = 0.0
	micronutrients[calcium] = 0.0
	micronutrients[magnesium] = 0.0
	micronutrients[phosphorus] = 0.0
	micronutrients[potassium] = 0.0
	micronutrients[sodium] = 0.0
	micronutrients[copper] = 0.0
	micronutrients[iron] = 0.0
	micronutrients[manganese] = 0.0
	micronutrients[zinc] = 0.0

	if isConsumed {
		for _, entry := range entries {
			micronutrients[vitaminA] += entry.product.VitaminA * entry.Consumed
			micronutrients[thiamin] += entry.product.Thiamin * entry.Consumed
			micronutrients[riboflavin] += entry.product.Riboflavin * entry.Consumed
			micronutrients[niacin] += entry.product.Niacin * entry.Consumed
			micronutrients[pantothenicAcid] += entry.product.PantothenicAcid * entry.Consumed
			micronutrients[vitaminB6] += entry.product.VitaminB6 * entry.Consumed
			micronutrients[folate] += entry.product.Folate * entry.Consumed
			micronutrients[vitaminB12] += entry.product.VitaminB12 * entry.Consumed
			micronutrients[vitaminC] += entry.product.VitaminC * entry.Consumed
			micronutrients[vitaminD] += entry.product.VitaminD * entry.Consumed
			micronutrients[vitaminE] += entry.product.VitaminE * entry.Consumed
			micronutrients[vitaminK] += entry.product.VitaminK * entry.Consumed
			micronutrients[calcium] += entry.product.Calcium * entry.Consumed
			micronutrients[magnesium] += entry.product.Magnesium * entry.Consumed
			micronutrients[phosphorus] += entry.product.Phosphorus * entry.Consumed
			micronutrients[potassium] += entry.product.Potassium * entry.Consumed
			micronutrients[sodium] += entry.product.Sodium * entry.Consumed
			micronutrients[copper] += entry.product.Copper * entry.Consumed
			micronutrients[iron] += entry.product.Iron * entry.Consumed
			micronutrients[manganese] += entry.product.Manganese * entry.Consumed
			micronutrients[zinc] += entry.product.Zinc * entry.Consumed
		}
	} else {
		for _, entry := range entries {
			micronutrients[vitaminA] += entry.product.VitaminA * entry.Amount
			micronutrients[thiamin] += entry.product.Thiamin * entry.Amount
			micronutrients[riboflavin] += entry.product.Riboflavin * entry.Amount
			micronutrients[niacin] += entry.product.Niacin * entry.Amount
			micronutrients[pantothenicAcid] += entry.product.PantothenicAcid * entry.Amount
			micronutrients[vitaminB6] += entry.product.VitaminB6 * entry.Amount
			micronutrients[folate] += entry.product.Folate * entry.Amount
			micronutrients[vitaminB12] += entry.product.VitaminB12 * entry.Amount
			micronutrients[vitaminC] += entry.product.VitaminC * entry.Amount
			micronutrients[vitaminD] += entry.product.VitaminD * entry.Amount
			micronutrients[vitaminE] += entry.product.VitaminE * entry.Amount
			micronutrients[vitaminK] += entry.product.VitaminK * entry.Amount
			micronutrients[calcium] += entry.product.Calcium * entry.Amount
			micronutrients[magnesium] += entry.product.Magnesium * entry.Amount
			micronutrients[phosphorus] += entry.product.Phosphorus * entry.Amount
			micronutrients[potassium] += entry.product.Potassium * entry.Amount
			micronutrients[sodium] += entry.product.Sodium * entry.Amount
			micronutrients[copper] += entry.product.Copper * entry.Amount
			micronutrients[iron] += entry.product.Iron * entry.Amount
			micronutrients[manganese] += entry.product.Manganese * entry.Amount
			micronutrients[zinc] += entry.product.Zinc * entry.Amount
		}
	}

	return micronutrients
}

func dayDiet(products []product) ([]dietEntry, bool) {

	nOptimizationColumns := len(products)

	commonOffset := 25

	ltOffset := commonOffset
	nLTConstraints := ltOffset + len(products)
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
		ltConstraintsLHS[4][i] = products[i].VitaminA
		ltConstraintsLHS[5][i] = products[i].Thiamin
		ltConstraintsLHS[6][i] = products[i].Riboflavin
		ltConstraintsLHS[7][i] = products[i].Niacin
		ltConstraintsLHS[8][i] = products[i].PantothenicAcid
		ltConstraintsLHS[9][i] = products[i].VitaminB6
		ltConstraintsLHS[10][i] = products[i].Folate
		ltConstraintsLHS[11][i] = products[i].VitaminB12
		ltConstraintsLHS[12][i] = products[i].VitaminC
		ltConstraintsLHS[13][i] = products[i].VitaminD
		ltConstraintsLHS[14][i] = products[i].VitaminE
		ltConstraintsLHS[15][i] = products[i].VitaminK
		ltConstraintsLHS[16][i] = products[i].Calcium
		ltConstraintsLHS[17][i] = products[i].Magnesium
		ltConstraintsLHS[18][i] = products[i].Phosphorus
		ltConstraintsLHS[19][i] = products[i].Potassium
		ltConstraintsLHS[20][i] = products[i].Sodium
		ltConstraintsLHS[21][i] = products[i].Copper
		ltConstraintsLHS[22][i] = products[i].Iron
		ltConstraintsLHS[23][i] = products[i].Manganese
		ltConstraintsLHS[24][i] = products[i].Zinc
	}
	ltConstraintsRHS[0] = upperKcals
	ltConstraintsRHS[1] = upperProteins
	ltConstraintsRHS[2] = upperCarbs
	ltConstraintsRHS[3] = upperFats
	ltConstraintsRHS[4] = upperVitaminA
	ltConstraintsRHS[5] = upperThiamin
	ltConstraintsRHS[6] = upperRiboflavin
	ltConstraintsRHS[7] = upperNiacin
	ltConstraintsRHS[8] = upperPantothenicAcid
	ltConstraintsRHS[9] = upperVitaminB6
	ltConstraintsRHS[10] = upperFolate
	ltConstraintsRHS[11] = upperVitaminB12
	ltConstraintsRHS[12] = upperVitaminC
	ltConstraintsRHS[13] = upperVitaminD
	ltConstraintsRHS[14] = upperVitaminE
	ltConstraintsRHS[15] = upperVitaminK
	ltConstraintsRHS[16] = upperCalcium
	ltConstraintsRHS[17] = upperMagnesium
	ltConstraintsRHS[18] = upperPhosphorus
	ltConstraintsRHS[19] = upperPotassium
	ltConstraintsRHS[20] = upperSodium
	ltConstraintsRHS[21] = upperCopper
	ltConstraintsRHS[22] = upperIron
	ltConstraintsRHS[23] = upperManganese
	ltConstraintsRHS[24] = upperZinc

	for i := 0; i < len(products); i++ {
		j := i + ltOffset
		ltConstraintsLHS[j][i] = 100.0
		ltConstraintsRHS[j] = products[i].Maximum
	}

	gtOffset := commonOffset
	nGTConstraints := gtOffset + len(products)
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
		gtConstraintsLHS[4][i] = products[i].VitaminA
		gtConstraintsLHS[5][i] = products[i].Thiamin
		gtConstraintsLHS[6][i] = products[i].Riboflavin
		gtConstraintsLHS[7][i] = products[i].Niacin
		gtConstraintsLHS[8][i] = products[i].PantothenicAcid
		gtConstraintsLHS[9][i] = products[i].VitaminB6
		gtConstraintsLHS[10][i] = products[i].Folate
		gtConstraintsLHS[11][i] = products[i].VitaminB12
		gtConstraintsLHS[12][i] = products[i].VitaminC
		gtConstraintsLHS[13][i] = products[i].VitaminD
		gtConstraintsLHS[14][i] = products[i].VitaminE
		gtConstraintsLHS[15][i] = products[i].VitaminK
		gtConstraintsLHS[16][i] = products[i].Calcium
		gtConstraintsLHS[17][i] = products[i].Magnesium
		gtConstraintsLHS[18][i] = products[i].Phosphorus
		gtConstraintsLHS[19][i] = products[i].Potassium
		gtConstraintsLHS[20][i] = products[i].Sodium
		gtConstraintsLHS[21][i] = products[i].Copper
		gtConstraintsLHS[22][i] = products[i].Iron
		gtConstraintsLHS[23][i] = products[i].Manganese
		gtConstraintsLHS[24][i] = products[i].Zinc
	}
	gtConstraintsRHS[0] = lowerKcals
	gtConstraintsRHS[1] = lowerProteins
	gtConstraintsRHS[2] = lowerCarbs
	gtConstraintsRHS[3] = lowerFats
	gtConstraintsRHS[4] = lowerVitaminA
	gtConstraintsRHS[5] = lowerThiamin
	gtConstraintsRHS[6] = lowerRiboflavin
	gtConstraintsRHS[7] = lowerNiacin
	gtConstraintsRHS[8] = lowerPantothenicAcid
	gtConstraintsRHS[9] = lowerVitaminB6
	gtConstraintsRHS[10] = lowerFolate
	gtConstraintsRHS[11] = lowerVitaminB12
	gtConstraintsRHS[12] = lowerVitaminC
	gtConstraintsRHS[13] = lowerVitaminD
	gtConstraintsRHS[14] = lowerVitaminE
	gtConstraintsRHS[15] = lowerVitaminK
	gtConstraintsRHS[16] = lowerCalcium
	gtConstraintsRHS[17] = lowerMagnesium
	gtConstraintsRHS[18] = lowerPhosphorus
	gtConstraintsRHS[19] = lowerPotassium
	gtConstraintsRHS[20] = lowerSodium
	gtConstraintsRHS[21] = lowerCopper
	gtConstraintsRHS[22] = lowerIron
	gtConstraintsRHS[23] = lowerManganese
	gtConstraintsRHS[24] = lowerZinc

	for i := 0; i < len(products); i++ {
		j := i + gtOffset
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

	dayDiet := []dietEntry{}

	for i, amount := range amounts {
		if amount <= 0.0 {
			continue
		}
		p := dietEntry{
			ID:      products[i].ID,
			Amount:  amount,
			product: products[i],
		}
		dayDiet = append(dayDiet, p)
	}

	dl := 0.0001

	totalKcals, totalProteins, totalCarbs, totalFats := totalMacronutrients(dayDiet, false)
	ok = totalKcals <= upperKcals+dl && totalKcals >= lowerKcals-dl &&
		totalProteins <= upperProteins+dl && totalProteins >= lowerProteins-dl &&
		totalCarbs <= upperCarbs+dl && totalCarbs >= lowerCarbs-dl &&
		totalFats <= upperFats+dl && totalFats >= lowerFats-dl
	if !ok {
		return []dietEntry{}, false
	}

	micronutrients := totalMicronutrients(dayDiet, false)
	ok = micronutrients[vitaminA] <= upperVitaminA+dl && micronutrients[vitaminA] >= lowerVitaminA-dl &&
		micronutrients[thiamin] <= upperThiamin+dl && micronutrients[thiamin] >= lowerThiamin-dl &&
		micronutrients[riboflavin] <= upperRiboflavin+dl && micronutrients[riboflavin] >= lowerRiboflavin-dl &&
		micronutrients[niacin] <= upperNiacin+dl && micronutrients[niacin] >= lowerNiacin-dl &&
		micronutrients[pantothenicAcid] <= upperPantothenicAcid+dl && micronutrients[pantothenicAcid] >= lowerPantothenicAcid-dl &&
		micronutrients[vitaminB6] <= upperVitaminB6+dl && micronutrients[vitaminB6] >= lowerVitaminB6-dl &&
		micronutrients[folate] <= upperFolate+dl && micronutrients[folate] >= lowerFolate-dl &&
		micronutrients[vitaminB12] <= upperVitaminB12+dl && micronutrients[vitaminB12] >= lowerVitaminB12-dl &&
		micronutrients[vitaminC] <= upperVitaminC+dl && micronutrients[vitaminC] >= lowerVitaminC-dl &&
		micronutrients[vitaminD] <= upperVitaminD+dl && micronutrients[vitaminD] >= lowerVitaminD-dl &&
		micronutrients[vitaminE] <= upperVitaminE+dl && micronutrients[vitaminE] >= lowerVitaminE-dl &&
		micronutrients[vitaminK] <= upperVitaminK+dl && micronutrients[vitaminK] >= lowerVitaminK-dl &&
		micronutrients[calcium] <= upperCalcium+dl && micronutrients[calcium] >= lowerCalcium-dl &&
		micronutrients[magnesium] <= upperMagnesium+dl && micronutrients[magnesium] >= lowerMagnesium-dl &&
		micronutrients[phosphorus] <= upperPhosphorus+dl && micronutrients[phosphorus] >= lowerPhosphorus-dl &&
		micronutrients[potassium] <= upperPotassium+dl && micronutrients[potassium] >= lowerPotassium-dl &&
		micronutrients[sodium] <= upperSodium+dl && micronutrients[sodium] >= lowerSodium-dl &&
		micronutrients[copper] <= upperCopper+dl && micronutrients[copper] >= lowerCopper-dl &&
		micronutrients[iron] <= upperIron+dl && micronutrients[iron] >= lowerIron-dl &&
		micronutrients[manganese] <= upperManganese+dl && micronutrients[manganese] >= lowerManganese-dl &&
		micronutrients[zinc] <= upperZinc+dl && micronutrients[zinc] >= lowerZinc-dl
	if !ok {
		return []dietEntry{}, false
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
				fmt.Println(entry.ID)
				return diet{}, false
			}
			day[i].product = p
		}
	}

	return d, true
}

func getDiet(path string, products []product) (diet, bool) {

	if !isJSONPath(path) {
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
	totalFlag := flag.Bool("total", false, "Use with `-diet` command to see total nutrients for today")
	detailedFlag := flag.Bool("detailed", false, "Use with `-diet` and `-total` flags to see detailed total nutrients for today")

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

		productsPerDay := 8
		if *productsPerDayFlag != defaultInteger {
			productsPerDay = int(*productsPerDayFlag)
		}

		productsPerWeek := 15
		if *productsPerWeekFlag != defaultInteger {
			productsPerWeek = int(*productsPerWeekFlag)
		}

		newDiet := make(diet, nWeekDays)

		weekIterations := 0

		var finished int32 = 0

		for i := 0; i < 8; i++ {
			go func() {
			newDietLoop:
				for {
					if atomic.LoadInt32(&finished) == 1 {
						break
					}

					thisNewDiet := make(diet, nWeekDays)

					weekProducts := pickRandomProducts(products, productsPerWeek)

					currentDay := 0
					dayIterations := 0
					for currentDay < nWeekDays {
						if dayIterations > 7500 {
							fmt.Printf("Iteration %d\n", weekIterations)
							weekIterations++
							continue newDietLoop
						}
						dayIterations++

						dayProducts := pickRandomProducts(weekProducts, productsPerDay)

						dayDiet, ok := dayDiet(dayProducts)
						if !ok {
							continue
						}

						thisNewDiet[currentDay] = dayDiet
						currentDay++
					}

					atomic.StoreInt32(&finished, 1)
					newDiet = thisNewDiet
					break
				}
			}()
		}

		for atomic.LoadInt32(&finished) == 0 {
			time.Sleep(1000 * time.Millisecond)
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

	} else if len(*dietFlag) > 0 && *totalFlag {

		*dietFlag = filepath.Clean(*dietFlag)

		diet, ok := getDiet(*dietFlag, products)
		if !ok {
			return
		}

		today := weekDay()

		totalKcals, totalProteins, totalCarbs, totalFats := totalMacronutrients(diet[today], true)

		fmt.Printf("Kcals: %f\n", totalKcals)
		fmt.Printf("Proteins: %f\n", totalProteins)
		fmt.Printf("Carbs: %f\n", totalCarbs)
		fmt.Printf("Fats: %f\n", totalFats)

		if *detailedFlag {
			micronutrients := totalMicronutrients(diet[today], true)
			fmt.Println()
			fmt.Printf("Vitamin A: %f\n", micronutrients[vitaminA])
			fmt.Printf("Thiamin: %f\n", micronutrients[thiamin])
			fmt.Printf("Riboflavin: %f\n", micronutrients[riboflavin])
			fmt.Printf("Niacin: %f\n", micronutrients[niacin])
			fmt.Printf("Pantothenic Acid: %f\n", micronutrients[pantothenicAcid])
			fmt.Printf("Vitamin B6: %f\n", micronutrients[vitaminB6])
			fmt.Printf("Folate: %f\n", micronutrients[folate])
			fmt.Printf("Vitamin B12: %f\n", micronutrients[vitaminB12])
			fmt.Printf("Vitamin C: %f\n", micronutrients[vitaminC])
			fmt.Printf("Vitamin D: %f\n", micronutrients[vitaminD])
			fmt.Printf("Vitamin E: %f\n", micronutrients[vitaminE])
			fmt.Printf("Vitamin K: %f\n", micronutrients[vitaminK])
			fmt.Printf("Calcium: %f\n", micronutrients[calcium])
			fmt.Printf("Magnesium: %f\n", micronutrients[magnesium])
			fmt.Printf("Phosphorus: %f\n", micronutrients[phosphorus])
			fmt.Printf("Potassium: %f\n", micronutrients[potassium])
			fmt.Printf("Sodium: %f\n", micronutrients[sodium])
			fmt.Printf("Copper: %f\n", micronutrients[copper])
			fmt.Printf("Iron: %f\n", micronutrients[iron])
			fmt.Printf("Manganese: %f\n", micronutrients[manganese])
			fmt.Printf("Zinc: %f\n", micronutrients[zinc])
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
			if index < len(day) {
				fmt.Println()
			}
		}

		return

	} else {
		fmt.Println("Unknown flag combination")
	}
}
