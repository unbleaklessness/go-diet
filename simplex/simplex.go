package simplex

import (
	"math"
)

const (
	bigM = 1e10
)

func pivot(matrix [][]float64, baiscColumns []int) (int, int) {

	lastRow := len(matrix) - 1
	lastColumn := len(matrix[0]) - 1

	rowIndex := -1
	columnIndex := -1

outer:
	for i, column := range matrix[lastRow] {

		if i >= lastColumn {
			continue
		}

		for _, j := range baiscColumns {
			if i == j {
				continue outer
			}
		}

		if column >= 0.0 {
			continue
		}

		rowMinimum := math.MaxFloat64

		for j := range matrix {

			if j >= lastRow {
				continue
			}

			if matrix[j][i] <= 0 {
				continue
			}

			x := matrix[j][lastColumn] / matrix[j][i]
			if x < rowMinimum {
				rowMinimum = x
				rowIndex = j
				columnIndex = i
			}
		}

		break
	}

	return rowIndex, columnIndex
}

func nonZeroRowForColumn(matrix [][]float64, column int, excludeRow int) int {

	for i, row := range matrix {
		if i == excludeRow {
			continue
		}
		if row[column] != 0.0 {
			return i
		}
	}

	return -1
}

func addRows(matrix [][]float64, targetRow int, sourceRow int, n float64) {
	for i := range matrix[targetRow] {
		matrix[targetRow][i] += matrix[sourceRow][i] * n
	}
}

func scalarMultiplyRow(matrix [][]float64, targetRow int, n float64) {
	for i := range matrix[targetRow] {
		matrix[targetRow][i] *= n
	}
}

func handlePivot(matrix [][]float64, pivotRow int, pivotColumn int, saveIndexes *[]int) {

	(*saveIndexes)[pivotRow] = pivotColumn

	pivotCell := matrix[pivotRow][pivotColumn]
	pivotCellInverse := 1.0 / pivotCell

	scalarMultiplyRow(matrix, pivotRow, pivotCellInverse)

	for row := range matrix {
		if row == pivotRow {
			continue
		}

		a := matrix[pivotRow][pivotColumn]
		x := matrix[row][pivotColumn]

		n := -x / a

		addRows(matrix, row, pivotRow, n)
	}
}

func findBasics(matrix [][]float64) ([]int, []int) {

	rows := make([]int, 0)
	columns := make([]int, 0)

	ignoredColumn := len(matrix[0]) - 2

outer:
	for i := range matrix[0] {

		if i == ignoredColumn {
			continue
		}

		for _, j := range rows {
			if j == i {
				continue outer
			}
		}

		rowForNonZero := -1
		foundNonZero := false

		for j := range matrix {
			if matrix[j][i] != 0.0 {

				if foundNonZero {
					foundNonZero = false
					break
				}

				foundNonZero = true
				rowForNonZero = j
			}
		}

		if foundNonZero {
			rows = append(rows, rowForNonZero)
			columns = append(columns, i)
		}
	}

	return rows, columns
}

func getResult(matrix [][]float64, basicColumns []int, nVariables int) ([]float64, float64, bool) {

	variables := make([]float64, nVariables)

	lastRow := len(matrix) - 1
	lastColumn := len(matrix[0]) - 1

	for i := range basicColumns {

		basicColumn := basicColumns[i]

		if basicColumn >= nVariables {
			continue
		}

		variables[basicColumn] = matrix[i][lastColumn]
	}

	return variables, matrix[lastRow][lastColumn], true
}

func zerofyCell(matrix [][]float64, cellRow int, cellColumn int) {

	nonZeroRow := nonZeroRowForColumn(matrix, cellColumn, cellRow)
	if nonZeroRow == -1 {
		panic("No non-zero row is found")
	}

	a := matrix[nonZeroRow][cellColumn]
	x := matrix[cellRow][cellColumn]

	n := -x / a

	addRows(matrix, cellRow, nonZeroRow, n)
}

func solutionExists(matrix [][]float64, basicRows []int, basicColumns []int, nConstraints int) bool {

	lastColumn := len(matrix[0]) - 1

	for _, i := range basicRows {
		if matrix[i][lastColumn] < 0.0 {
			return false
		}
	}

	return true
}

func makeSaveIndexes(basicColumns []int, nConstraints int) []int {

	indexes := make([]int, nConstraints)

	for _, i := range indexes {
		indexes[i] = -1
	}

	for i := range basicColumns {
		indexes[i] = i
	}

	return indexes
}

// Simplex maximizes given objective function subject to given constraints.
func Simplex(
	objective []float64,
	gtConstraintsLHS [][]float64, gtConstraintsRHS []float64,
	ltConstraintsLHS [][]float64, ltConstraintsRHS []float64,
	eqConstraintsLHS [][]float64, eqConstraintsRHS []float64) ([]float64, float64, bool) {

	for i := range objective {
		objective[i] *= -1
	}

	nVariables := len(objective)
	nGTConstraints := len(gtConstraintsLHS)
	nLTConstraints := len(ltConstraintsLHS)
	nEQConstraints := len(eqConstraintsLHS)
	nConstraints := nGTConstraints + nLTConstraints + nEQConstraints

	matrix := make([][]float64, 0)

	totalSlacks := nGTConstraints + nLTConstraints
	totalSurpluses := nGTConstraints + nEQConstraints

	slackIndex := nVariables
	slackIndexes := make([]int, totalSlacks)

	surplusIndex := nVariables + totalSlacks
	surplusIndexes := make([]int, totalSurpluses)

	for i := 0; i < totalSlacks; i++ {
		for j := range ltConstraintsLHS {
			element := 0.0
			if i == j {
				element = 1.0
			}
			ltConstraintsLHS[j] = append(ltConstraintsLHS[j], element)
		}
		for j := range gtConstraintsLHS {
			index := j + nLTConstraints
			element := 0.0
			if i == index {
				element = -1.0
			}
			gtConstraintsLHS[j] = append(gtConstraintsLHS[j], element)
		}
		for j := range eqConstraintsLHS {
			eqConstraintsLHS[j] = append(eqConstraintsLHS[j], 0.0)
		}
		objective = append(objective, 0.0)
		slackIndexes[i] = slackIndex
		slackIndex++
	}

	for i := 0; i < totalSurpluses; i++ {
		for j := range eqConstraintsLHS {
			element := 0.0
			if i == j {
				element = 1.0
			}
			eqConstraintsLHS[j] = append(eqConstraintsLHS[j], element)
		}
		for j := range gtConstraintsLHS {
			index := j + nEQConstraints
			element := 0.0
			if i == index {
				element = 1.0
			}
			gtConstraintsLHS[j] = append(gtConstraintsLHS[j], element)
		}
		for j := range ltConstraintsLHS {
			ltConstraintsLHS[j] = append(ltConstraintsLHS[j], 0.0)
		}
		objective = append(objective, bigM)
		surplusIndexes[i] = surplusIndex
		surplusIndex++
	}

	for j := range eqConstraintsLHS {
		eqConstraintsLHS[j] = append(eqConstraintsLHS[j], 0.0, eqConstraintsRHS[j])
	}
	for j := range gtConstraintsLHS {
		gtConstraintsLHS[j] = append(gtConstraintsLHS[j], 0.0, gtConstraintsRHS[j])
	}
	for j := range ltConstraintsLHS {
		ltConstraintsLHS[j] = append(ltConstraintsLHS[j], 0.0, ltConstraintsRHS[j])
	}

	objective = append(objective, 1.0, 0.0)

	matrix = append(matrix, gtConstraintsLHS...)
	matrix = append(matrix, ltConstraintsLHS...)
	matrix = append(matrix, eqConstraintsLHS...)
	matrix = append(matrix, objective)

	lastRow := len(matrix) - 1

	for _, i := range surplusIndexes {
		zerofyCell(matrix, lastRow, i)
	}

	basicRows, basicColumns := findBasics(matrix)

	if !solutionExists(matrix, basicRows, basicColumns, nConstraints) {
		return []float64{}, 0.0, false
	}

	saveIndexes := makeSaveIndexes(basicColumns, nConstraints)

	for {
		pivotRow, pivotColumn := pivot(matrix, basicColumns)
		if pivotRow == -1 {
			break
		}

		handlePivot(matrix, pivotRow, pivotColumn, &saveIndexes)
	}

	return getResult(matrix, saveIndexes, nVariables)
}
