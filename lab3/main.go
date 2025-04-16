package main

import (
	"fmt"
	"math"
)

// Подынтегральная функция для варианта 4: f(x) = -2x^3 - 4x^2 + 8x - 4
func f(x float64) float64 {
	return -2*x*x*x - 4*x*x + 8*x - 4
}

// Аналитическая первообразная:
// F(x) = -1/2*x^4 - (4/3)*x^3 + 4*x^2 - 4*x
func F(x float64) float64 {
	return -0.5*math.Pow(x, 4) - (4.0/3.0)*math.Pow(x, 3) + 4*x*x - 4*x
}

// Вычисление точного значения интеграла по формуле Ньютона-Лейбница
func analyticalIntegral(a, b float64) float64 {
	return F(b) - F(a)
}

// Метод Ньютона-Котеса с n = 6 (7 узлов) и весовыми коэффициентами
func newtonCotes(a, b float64) float64 {
	n := 6
	h := (b - a) / float64(n)
	// Весовые коэффициенты для 7 узлов
	weights := []float64{41, 216, 27, 272, 27, 216, 41}
	sum := 0.0
	for i := 0; i <= n; i++ {
		x := a + float64(i)*h
		sum += weights[i] * f(x)
	}
	// Фактор: (b - a) / 840, где 840 — сумма весов для формулы данного порядка
	return (b - a) * sum / 840.0
}

// Метод средних прямоугольников с n разбиениями
func midpoint(a, b float64, n int) float64 {
	h := (b - a) / float64(n)
	sum := 0.0
	for i := 0; i < n; i++ {
		// Середина каждого отрезка
		mid := a + (float64(i)+0.5)*h
		sum += f(mid)
	}
	return h * sum
}

// Метод трапеций с n разбиениями
func trapezoidal(a, b float64, n int) float64 {
	h := (b - a) / float64(n)
	sum := (f(a) + f(b)) / 2.0
	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		sum += f(x)
	}
	return h * sum
}

// Метод Симпсона с n разбиениями (n должно быть чётным)
func simpson(a, b float64, n int) float64 {
	h := (b - a) / float64(n)
	sum := f(a) + f(b)
	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		if i%2 != 0 {
			sum += 4 * f(x)
		} else {
			sum += 2 * f(x)
		}
	}
	return h * sum / 3.0
}

func main() {
	// Границы интегрирования для варианта 4
	a := -3.0
	b := -1.0

	// Вычисление аналитического значения интеграла
	I_exact := analyticalIntegral(a, b)

	// Вычисление интеграла численными методами
	I_nc := newtonCotes(a, b)
	I_mid := midpoint(a, b, 10)
	I_trap := trapezoidal(a, b, 10)
	I_simp := simpson(a, b, 10)

	fmt.Printf("Аналитическое значение интеграла: %.6f\n", I_exact)
	fmt.Printf("Метод Ньютона-Котеса (n=6):      %.6f (Ошибка: %.2f%%)\n", I_nc, math.Abs((I_nc-I_exact)/I_exact)*100)
	fmt.Printf("Метод средних прямоугольников:  %.6f (Ошибка: %.2f%%)\n", I_mid, math.Abs((I_mid-I_exact)/I_exact)*100)
	fmt.Printf("Метод трапеций:                %.6f (Ошибка: %.2f%%)\n", I_trap, math.Abs((I_trap-I_exact)/I_exact)*100)
	fmt.Printf("Метод Симпсона:                %.6f (Ошибка: %.2f%%)\n", I_simp, math.Abs((I_simp-I_exact)/I_exact)*100)
}
