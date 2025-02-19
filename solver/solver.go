package solver

import (
	"fmt"
	"math"
)

type Result struct {
	Solution   []float64
	Iterations int
	Errors     []float64
	MatrixNorm float64
}

func SolveSystem(A [][]float64, b []float64, precision float64) (*Result, error) {
	n := len(A)
	if n == 0 || len(b) != n {
		return nil, fmt.Errorf("invalid matrix or vector dimensions")
	}

	dominantA, dominantB, ok := enforceDiagonalDominance(A, b)
	if !ok {
		return nil, fmt.Errorf("unable to achieve diagonal dominance")
	}

	x := make([]float64, n)      // Current solution
	xPrev := make([]float64, n)  // Previous iteration
	errors := make([]float64, n) // Error vector

	iterations := 0
	maxIter := 10000

	for {
		iterations++

		// Perform iteration
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				if j != i {
					sum += dominantA[i][j] * xPrev[j]
				}
			}
			x[i] = (dominantB[i] - sum) / dominantA[i][i]
		}

		// Calculate error and check convergence
		maxError := 0.0
		for i := 0; i < n; i++ {
			errors[i] = math.Abs(x[i] - xPrev[i])
			if errors[i] > maxError {
				maxError = errors[i]
			}
		}

		if maxError < precision {
			break
		}

		if iterations >= maxIter {
			return nil, fmt.Errorf("solution did not converge within %d iterations", maxIter)
		}

		copy(xPrev, x)
	}

	return &Result{
		Solution:   x,
		Iterations: iterations,
		Errors:     errors,
		MatrixNorm: calculateNorm(dominantA),
	}, nil
}

func enforceDiagonalDominance(A [][]float64, b []float64) ([][]float64, []float64, bool) {
	n := len(A)
	used := make([]bool, n)
	newA := make([][]float64, n)
	newB := make([]float64, n)

	for i := 0; i < n; i++ {
		maxDiagonal := -1.0
		maxRow := -1

		for row := 0; row < n; row++ {
			if used[row] {
				continue
			}

			diagonal := math.Abs(A[row][i])
			sum := 0.0
			for col := 0; col < n; col++ {
				if col != i {
					sum += math.Abs(A[row][col])
				}
			}

			if diagonal > sum && diagonal > maxDiagonal {
				maxDiagonal = diagonal
				maxRow = row
			}
		}

		if maxRow == -1 {
			return A, b, false
		}

		newA[i] = make([]float64, n)
		copy(newA[i], A[maxRow])
		newB[i] = b[maxRow]
		used[maxRow] = true
	}

	return newA, newB, true
}

func calculateNorm(A [][]float64) float64 {
	n := len(A)
	maxSum := 0.0

	for i := 0; i < n; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += math.Abs(A[i][j])
		}
		if sum > maxSum {
			maxSum = sum
		}
	}

	return maxSum
}
