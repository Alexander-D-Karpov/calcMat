// lab6.go - Численное решение обыкновенных дифференциальных уравнений
// Вариант 4: методы Эйлера, Усовершенствованный Эйлера, Адамса
//
// Запуск: go run lab6.go
package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Структура для хранения результатов решения
type Solution struct {
	X      []float64
	Y      []float64
	Method string
}

// Структура для хранения ОДУ
type ODE struct {
	Name     string
	Function func(x, y float64) float64
	Exact    func(x float64) float64
	Y0       float64
	X0       float64
}

// Доступные ОДУ
var ODEs = []ODE{
	{
		Name:     "y' = y",
		Function: func(x, y float64) float64 { return y },
		Exact:    func(x float64) float64 { return math.Exp(x) },
		Y0:       1.0,
		X0:       0.0,
	},
	{
		Name:     "y' = -y",
		Function: func(x, y float64) float64 { return -y },
		Exact:    func(x float64) float64 { return math.Exp(-x) },
		Y0:       1.0,
		X0:       0.0,
	},
	{
		Name:     "y' = x + y",
		Function: func(x, y float64) float64 { return x + y },
		Exact:    func(x float64) float64 { return 2*math.Exp(x) - x - 1 },
		Y0:       1.0,
		X0:       0.0,
	},
	{
		Name:     "y' = x - y",
		Function: func(x, y float64) float64 { return x - y },
		Exact:    func(x float64) float64 { return x - 1 + 2*math.Exp(-x) },
		Y0:       1.0,
		X0:       0.0,
	},
}

// Цвета для методов
var methodColors = map[string]color.RGBA{
	"Exact":         {R: 0, G: 0, B: 128, A: 255},   // Темно-синий
	"Euler":         {R: 255, G: 0, B: 0, A: 255},   // Красный
	"ImprovedEuler": {R: 0, G: 128, B: 0, A: 255},   // Зеленый
	"Adams":         {R: 255, G: 140, B: 0, A: 255}, // Оранжевый
	"RungeKutta4":   {R: 128, G: 0, B: 128, A: 255}, // Фиолетовый
}

// Названия методов для отображения
var methodNames = map[string]string{
	"Exact":         "Точное решение",
	"Euler":         "Метод Эйлера",
	"ImprovedEuler": "Усовершенствованный метод Эйлера",
	"Adams":         "Метод Адамса",
	"RungeKutta4":   "Метод Рунге-Кутта 4",
}

// Метод Эйлера
func eulerMethod(ode ODE, x0, xEnd, h float64) Solution {
	n := int((xEnd-x0)/h) + 1
	x := make([]float64, n)
	y := make([]float64, n)

	x[0] = x0
	y[0] = ode.Y0

	for i := 1; i < n; i++ {
		x[i] = x[i-1] + h
		y[i] = y[i-1] + h*ode.Function(x[i-1], y[i-1])
	}

	return Solution{X: x, Y: y, Method: "Euler"}
}

// Усовершенствованный метод Эйлера (метод Хейна)
func improvedEulerMethod(ode ODE, x0, xEnd, h float64) Solution {
	n := int((xEnd-x0)/h) + 1
	x := make([]float64, n)
	y := make([]float64, n)

	x[0] = x0
	y[0] = ode.Y0

	for i := 1; i < n; i++ {
		x[i] = x[i-1] + h
		// Предиктор (метод Эйлера)
		yPredict := y[i-1] + h*ode.Function(x[i-1], y[i-1])
		// Корректор (трапецевидальное правило)
		y[i] = y[i-1] + h*(ode.Function(x[i-1], y[i-1])+ode.Function(x[i], yPredict))/2
	}

	return Solution{X: x, Y: y, Method: "ImprovedEuler"}
}

// Метод Рунге-Кутта 4-го порядка (для инициализации метода Адамса)
func rungeKutta4(ode ODE, x0, xEnd, h float64) Solution {
	n := int((xEnd-x0)/h) + 1
	x := make([]float64, n)
	y := make([]float64, n)

	x[0] = x0
	y[0] = ode.Y0

	for i := 1; i < n; i++ {
		x[i] = x[i-1] + h
		k1 := h * ode.Function(x[i-1], y[i-1])
		k2 := h * ode.Function(x[i-1]+h/2, y[i-1]+k1/2)
		k3 := h * ode.Function(x[i-1]+h/2, y[i-1]+k2/2)
		k4 := h * ode.Function(x[i-1]+h, y[i-1]+k3)
		y[i] = y[i-1] + (k1+2*k2+2*k3+k4)/6
	}

	return Solution{X: x, Y: y, Method: "RungeKutta4"}
}

// Метод Адамса (4-го порядка)
func adamsMethod(ode ODE, x0, xEnd, h float64) Solution {
	n := int((xEnd-x0)/h) + 1
	x := make([]float64, n)
	y := make([]float64, n)

	// Инициализация первых 4 точек методом Рунге-Кутта
	rk4 := rungeKutta4(ode, x0, x0+3*h, h)

	for i := 0; i < 4 && i < n; i++ {
		x[i] = rk4.X[i]
		y[i] = rk4.Y[i]
	}

	// Применение метода Адамса для остальных точек
	for i := 4; i < n; i++ {
		x[i] = x[i-1] + h

		// Вычисляем значения производной в предыдущих точках
		f1 := ode.Function(x[i-1], y[i-1])
		f2 := ode.Function(x[i-2], y[i-2])
		f3 := ode.Function(x[i-3], y[i-3])
		f4 := ode.Function(x[i-4], y[i-4])

		// Формула Адамса-Башфорта 4-го порядка (предиктор)
		yPredict := y[i-1] + h*(55*f1-59*f2+37*f3-9*f4)/24

		// Формула Адамса-Мултона 4-го порядка (корректор)
		fPredict := ode.Function(x[i], yPredict)
		y[i] = y[i-1] + h*(9*fPredict+19*f1-5*f2+f3)/24
	}

	return Solution{X: x, Y: y, Method: "Adams"}
}

// Точное решение
func exactSolution(ode ODE, x0, xEnd, h float64) Solution {
	n := int((xEnd-x0)/h) + 1
	x := make([]float64, n)
	y := make([]float64, n)

	for i := 0; i < n; i++ {
		x[i] = x0 + float64(i)*h
		y[i] = ode.Exact(x[i])
	}

	return Solution{X: x, Y: y, Method: "Exact"}
}

// Правило Рунге для оценки погрешности
func rungeRule(sol1, sol2 Solution, p int) []float64 {
	errors := make([]float64, len(sol1.Y))
	factor := math.Pow(2, float64(p)) - 1

	for i := 0; i < len(errors) && i < len(sol2.Y); i++ {
		errors[i] = math.Abs(sol2.Y[i]-sol1.Y[i]) / factor
	}

	return errors
}

// Вычисление максимальной погрешности
func maxError(computed, exact []float64) float64 {
	maxErr := 0.0
	for i := 0; i < len(computed) && i < len(exact); i++ {
		err := math.Abs(computed[i] - exact[i])
		if err > maxErr {
			maxErr = err
		}
	}
	return maxErr
}

// Создание точек для графика из решения
func solutionToXYs(sol Solution) plotter.XYs {
	pts := make(plotter.XYs, len(sol.X))
	for i := range sol.X {
		pts[i].X = sol.X[i]
		pts[i].Y = sol.Y[i]
	}
	return pts
}

// График сравнения всех методов
func plotComparisonAllMethods(solutions []Solution, eqInfo map[string]string, h float64, baseFilename string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Сравнение численных методов\n%s, h = %.3f", eqInfo["name"], h)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	// Добавление точного решения (линия)
	for _, sol := range solutions {
		if sol.Method == "Exact" {
			pts := solutionToXYs(sol)
			line, err := plotter.NewLine(pts)
			if err != nil {
				return err
			}
			line.LineStyle.Width = vg.Points(3)
			line.LineStyle.Color = methodColors[sol.Method]
			p.Add(line)
			p.Legend.Add(methodNames[sol.Method], line)
			break
		}
	}

	// Добавление численных методов (точки + линии)
	for _, sol := range solutions {
		if sol.Method != "Exact" {
			pts := solutionToXYs(sol)

			// Линия
			line, err := plotter.NewLine(pts)
			if err != nil {
				return err
			}
			line.LineStyle.Width = vg.Points(2)
			line.LineStyle.Color = methodColors[sol.Method]
			line.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(2)}

			// Точки
			scatter, err := plotter.NewScatter(pts)
			if err != nil {
				return err
			}
			scatter.GlyphStyle.Color = methodColors[sol.Method]
			scatter.GlyphStyle.Radius = vg.Points(3)
			scatter.GlyphStyle.Shape = draw.CircleGlyph{}

			p.Add(line, scatter)
			p.Legend.Add(methodNames[sol.Method], line)
		}
	}

	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	filename := fmt.Sprintf("%s_comparison_all_methods.png", baseFilename)
	err := p.Save(12*vg.Inch, 8*vg.Inch, filename)
	if err == nil {
		fmt.Printf("График сравнения всех методов сохранен: %s\n", filename)
	}
	return err
}

// Отдельный график для каждого метода
func plotIndividualMethod(sol Solution, exactSol Solution, eqInfo map[string]string, h float64, baseFilename string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s\n%s, h = %.3f", methodNames[sol.Method], eqInfo["name"], h)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	// Точное решение
	exactPts := solutionToXYs(exactSol)
	exactLine, err := plotter.NewLine(exactPts)
	if err != nil {
		return err
	}
	exactLine.LineStyle.Width = vg.Points(3)
	exactLine.LineStyle.Color = methodColors["Exact"]
	p.Add(exactLine)
	p.Legend.Add(methodNames["Exact"], exactLine)

	// Численный метод
	pts := solutionToXYs(sol)

	// Линия
	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.LineStyle.Width = vg.Points(2)
	line.LineStyle.Color = methodColors[sol.Method]
	line.LineStyle.Dashes = []vg.Length{vg.Points(4), vg.Points(2)}

	// Точки
	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		return err
	}
	scatter.GlyphStyle.Color = methodColors[sol.Method]
	scatter.GlyphStyle.Radius = vg.Points(4)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}

	p.Add(line, scatter)
	p.Legend.Add(methodNames[sol.Method], line)

	// Вычисление и отображение максимальной погрешности
	maxErr := maxError(sol.Y, exactSol.Y)

	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	filename := fmt.Sprintf("%s_%s_individual.png", baseFilename, strings.ToLower(sol.Method))
	err = p.Save(10*vg.Inch, 7*vg.Inch, filename)
	if err == nil {
		fmt.Printf("График %s сохранен: %s (макс. погрешность: %.2e)\n",
			methodNames[sol.Method], filename, maxErr)
	}
	return err
}

// График анализа погрешностей
func plotErrorAnalysis(solutions []Solution, exactSol Solution, eqInfo map[string]string, h float64, baseFilename string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Анализ погрешностей\n%s, h = %.3f", eqInfo["name"], h)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "Абсолютная погрешность"

	for _, sol := range solutions {
		if sol.Method != "Exact" {
			// Вычисление погрешности
			errorPts := make(plotter.XYs, len(sol.X))
			for i := range sol.X {
				if i < len(exactSol.Y) {
					errorPts[i].X = sol.X[i]
					errorPts[i].Y = math.Abs(sol.Y[i] - exactSol.Y[i])
				}
			}

			// Линия погрешности
			line, err := plotter.NewLine(errorPts)
			if err != nil {
				return err
			}
			line.LineStyle.Width = vg.Points(2)
			line.LineStyle.Color = methodColors[sol.Method]

			// Точки погрешности
			scatter, err := plotter.NewScatter(errorPts)
			if err != nil {
				return err
			}
			scatter.GlyphStyle.Color = methodColors[sol.Method]
			scatter.GlyphStyle.Radius = vg.Points(3)
			scatter.GlyphStyle.Shape = draw.CircleGlyph{}

			p.Add(line, scatter)
			p.Legend.Add(methodNames[sol.Method], line)
		}
	}

	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	filename := fmt.Sprintf("%s_error_analysis.png", baseFilename)
	err := p.Save(12*vg.Inch, 8*vg.Inch, filename)
	if err == nil {
		fmt.Printf("График анализа погрешностей сохранен: %s\n", filename)
	}
	return err
}

// График логарифмического анализа погрешностей
func plotLogErrorAnalysis(solutions []Solution, exactSol Solution, eqInfo map[string]string, h float64, baseFilename string) error {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Логарифмический анализ погрешностей\n%s, h = %.3f", eqInfo["name"], h)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "log₁₀(Абсолютная погрешность)"

	for _, sol := range solutions {
		if sol.Method != "Exact" {
			// Вычисление логарифмической погрешности
			logErrorPts := make(plotter.XYs, 0)
			for i := range sol.X {
				if i < len(exactSol.Y) {
					absError := math.Abs(sol.Y[i] - exactSol.Y[i])
					if absError > 1e-16 { // Избегаем log(0)
						logErrorPts = append(logErrorPts, plotter.XY{
							X: sol.X[i],
							Y: math.Log10(absError),
						})
					}
				}
			}

			if len(logErrorPts) > 0 {
				// Линия погрешности
				line, err := plotter.NewLine(logErrorPts)
				if err != nil {
					return err
				}
				line.LineStyle.Width = vg.Points(2)
				line.LineStyle.Color = methodColors[sol.Method]

				// Точки погрешности
				scatter, err := plotter.NewScatter(logErrorPts)
				if err != nil {
					return err
				}
				scatter.GlyphStyle.Color = methodColors[sol.Method]
				scatter.GlyphStyle.Radius = vg.Points(3)
				scatter.GlyphStyle.Shape = draw.CircleGlyph{}

				p.Add(line, scatter)
				p.Legend.Add(methodNames[sol.Method], line)
			}
		}
	}

	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	filename := fmt.Sprintf("%s_log_error_analysis.png", baseFilename)
	err := p.Save(12*vg.Inch, 8*vg.Inch, filename)
	if err == nil {
		fmt.Printf("График логарифмического анализа погрешностей сохранен: %s\n", filename)
	}
	return err
}

// Вспомогательные функции
func findMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func findMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Сохранение данных в CSV файл
func saveToCSV(solutions []Solution, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Заголовки
	headers := []string{"x"}
	for _, sol := range solutions {
		headers = append(headers, sol.Method)
	}
	writer.Write(headers)

	// Данные
	if len(solutions) > 0 {
		n := len(solutions[0].X)
		for i := 0; i < n; i++ {
			row := []string{fmt.Sprintf("%.6f", solutions[0].X[i])}
			for _, sol := range solutions {
				if i < len(sol.Y) {
					row = append(row, fmt.Sprintf("%.6f", sol.Y[i]))
				} else {
					row = append(row, "")
				}
			}
			writer.Write(row)
		}
	}

	return nil
}

// Создание всех графиков
func createAllPlots(solutions []Solution, eqInfo map[string]string, h float64, baseFilename string) {
	var exactSol Solution
	for _, sol := range solutions {
		if sol.Method == "Exact" {
			exactSol = sol
			break
		}
	}

	fmt.Println("\nСоздание графиков...")

	// 1. График сравнения всех методов
	if err := plotComparisonAllMethods(solutions, eqInfo, h, baseFilename); err != nil {
		fmt.Printf("Ошибка создания сравнительного графика: %v\n", err)
	}

	// 2. Отдельные графики для каждого метода
	for _, sol := range solutions {
		if sol.Method != "Exact" {
			if err := plotIndividualMethod(sol, exactSol, eqInfo, h, baseFilename); err != nil {
				fmt.Printf("Ошибка создания графика для %s: %v\n", sol.Method, err)
			}
		}
	}

	// 3. Анализ погрешностей
	if err := plotErrorAnalysis(solutions, exactSol, eqInfo, h, baseFilename); err != nil {
		fmt.Printf("Ошибка создания графика анализа погрешностей: %v\n", err)
	}

	// 4. Логарифмический анализ погрешностей
	if err := plotLogErrorAnalysis(solutions, exactSol, eqInfo, h, baseFilename); err != nil {
		fmt.Printf("Ошибка создания логарифмического графика погрешностей: %v\n", err)
	}
}

// Вывод таблицы результатов
func printTable(solutions []Solution) {
	if len(solutions) == 0 {
		return
	}

	fmt.Printf("\n%-10s", "x")
	for _, sol := range solutions {
		fmt.Printf("%-15s", methodNames[sol.Method])
	}
	fmt.Println()

	fmt.Printf("%-10s", "")
	for range solutions {
		fmt.Printf("%-15s", "")
	}
	fmt.Println()

	n := len(solutions[0].X)
	for i := 0; i < n; i++ {
		fmt.Printf("%-10.6f", solutions[0].X[i])
		for _, sol := range solutions {
			if i < len(sol.Y) {
				fmt.Printf("%-15.6f", sol.Y[i])
			} else {
				fmt.Printf("%-15s", "-")
			}
		}
		fmt.Println()
	}
}

// Анализ погрешности
func analyzeErrors(solutions []Solution, exact Solution) {
	fmt.Println("\nАнализ погрешности:")
	fmt.Printf("%-25s %-15s\n", "Метод", "Макс. погрешность")
	fmt.Println(strings.Repeat("-", 40))

	for _, sol := range solutions {
		if sol.Method != "Exact" {
			maxErr := maxError(sol.Y, exact.Y)
			fmt.Printf("%-25s %-15.6e\n", methodNames[sol.Method], maxErr)
		}
	}
}

// Чтение входных данных
func readInput() (int, float64, float64, float64) {
	// Выбор ОДУ
	fmt.Println("Доступные ОДУ:")
	for i, ode := range ODEs {
		fmt.Printf("%d. %s, y(%.1f) = %.1f\n", i+1, ode.Name, ode.X0, ode.Y0)
	}
	fmt.Print("Выберите ОДУ (1-4): ")

	var odeChoice int
	fmt.Scan(&odeChoice)
	if odeChoice < 1 || odeChoice > len(ODEs) {
		odeChoice = 1
	}
	odeChoice-- // Преобразование к индексу массива

	// Интервал интегрирования
	fmt.Print("Введите конечную точку интервала (x_end): ")
	var xEnd float64
	fmt.Scan(&xEnd)

	// Шаг интегрирования
	fmt.Print("Введите шаг интегрирования (h): ")
	var h float64
	fmt.Scan(&h)

	// Точность (для правила Рунге)
	fmt.Print("Введите требуемую точность (eps): ")
	var eps float64
	fmt.Scan(&eps)

	return odeChoice, xEnd, h, eps
}

func main() {
	fmt.Println("Лабораторная работа №6")
	fmt.Println("Численное решение ОДУ")
	fmt.Println("Вариант 4: Методы Эйлера, Усовершенствованный Эйлера, Адамса")
	fmt.Println(strings.Repeat("=", 60))

	// Чтение входных данных
	odeChoice, xEnd, h, eps := readInput()
	ode := ODEs[odeChoice]

	fmt.Printf("\nВыбранное ОДУ: %s\n", ode.Name)
	fmt.Printf("Начальные условия: y(%.1f) = %.1f\n", ode.X0, ode.Y0)
	fmt.Printf("Интервал: [%.1f, %.1f]\n", ode.X0, xEnd)
	fmt.Printf("Шаг: h = %.4f\n", h)
	fmt.Printf("Точность: eps = %.2e\n", eps)

	// Решение численными методами
	var solutions []Solution

	// Метод Эйлера
	eulerSol := eulerMethod(ode, ode.X0, xEnd, h)
	solutions = append(solutions, eulerSol)

	// Усовершенствованный метод Эйлера
	improvedEulerSol := improvedEulerMethod(ode, ode.X0, xEnd, h)
	solutions = append(solutions, improvedEulerSol)

	// Метод Адамса
	adamsSol := adamsMethod(ode, ode.X0, xEnd, h)
	solutions = append(solutions, adamsSol)

	// Точное решение
	exactSol := exactSolution(ode, ode.X0, xEnd, h)
	solutions = append(solutions, exactSol)

	// Сохранение данных в CSV
	csvFilename := fmt.Sprintf("ode_data_eq%d_h%.3f.csv", odeChoice+1, h)
	if err := saveToCSV(solutions, csvFilename); err != nil {
		fmt.Printf("Ошибка сохранения CSV: %v\n", err)
	} else {
		fmt.Printf("Данные сохранены в файл: %s\n", csvFilename)
	}

	// Вывод таблицы результатов
	fmt.Println("\nТаблица результатов:")
	printTable(solutions)

	// Анализ погрешности
	analyzeErrors(solutions, exactSol)

	// Правило Рунге для одношаговых методов
	fmt.Println("\nПравило Рунге для одношаговых методов:")

	h2 := h / 2

	// Эйлер с шагом h/2
	eulerSolH2 := eulerMethod(ode, ode.X0, xEnd, h2)
	eulerErrors := rungeRule(eulerSol, eulerSolH2, 1) // Порядок метода Эйлера = 1
	maxEulerError := func() float64 {
		max := 0.0
		for _, err := range eulerErrors {
			if err > max {
				max = err
			}
		}
		return max
	}()
	fmt.Printf("Метод Эйлера - макс. оценка погрешности: %.6e\n", maxEulerError)

	// Усовершенствованный Эйлер с шагом h/2
	improvedEulerSolH2 := improvedEulerMethod(ode, ode.X0, xEnd, h2)
	improvedEulerErrors := rungeRule(improvedEulerSol, improvedEulerSolH2, 2) // Порядок = 2
	maxImprovedEulerError := func() float64 {
		max := 0.0
		for _, err := range improvedEulerErrors {
			if err > max {
				max = err
			}
		}
		return max
	}()
	fmt.Printf("Усовершенствованный метод Эйлера - макс. оценка погрешности: %.6e\n", maxImprovedEulerError)

	// Информация об уравнении для графиков
	eqInfo := map[string]string{
		"name": ode.Name,
	}

	// Базовое имя для файлов графиков
	baseFilename := fmt.Sprintf("ode_data_eq%d_h%.3f", odeChoice+1, h)

	// Создание всех графиков
	createAllPlots(solutions, eqInfo, h, baseFilename)

	if maxEulerError > eps {
		newH := h * math.Sqrt(eps/maxEulerError)
		fmt.Printf("Для достижения точности %.2e рекомендуется шаг h ≤ %.6f\n", eps, newH)
	} else {
		fmt.Printf("Текущий шаг h = %.6f обеспечивает требуемую точность\n", h)
	}
}
