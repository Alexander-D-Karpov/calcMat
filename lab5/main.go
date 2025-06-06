// lab5.go - Интерполяция функции (вариант 4)
//
// Запуск: go run lab5.go [--test] [--bonus]
//
//	--test: запускает тестовые наборы данных
//	--bonus: включает дополнительные методы (Стирлинг и Бессель)
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Структура для хранения таблицы и таблицы разностей
type Table struct {
	X, Y []float64
	Diff [][]float64 // Таблица конечных разностей
}

// Вычисляет факториал
func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	return float64(n) * factorial(n-1)
}

// Проверка на дублирующиеся узлы
func checkDuplicates(x []float64) error {
	set := make(map[float64]struct{})
	for _, v := range x {
		if _, ok := set[v]; ok {
			return errors.New("дублирующий узел")
		}
		set[v] = struct{}{}
	}
	return nil
}

// Проверка равномерности сетки
func checkUniformGrid(x []float64) (float64, bool) {
	if len(x) < 2 {
		return 0, false
	}
	h := x[1] - x[0]
	const eps = 1e-9
	for i := 2; i < len(x); i++ {
		if math.Abs((x[i]-x[i-1])-h) > eps {
			return 0, false
		}
	}
	return h, true
}

// Строит таблицу конечных разностей
func buildDifferenceTable(table *Table) {
	n := len(table.Y)
	table.Diff = make([][]float64, n)

	// Первый ряд - сами значения функции
	table.Diff[0] = make([]float64, n)
	copy(table.Diff[0], table.Y)

	// Вычисляем разности
	for i := 1; i < n; i++ {
		table.Diff[i] = make([]float64, n-i)
		for j := 0; j < n-i; j++ {
			table.Diff[i][j] = table.Diff[i-1][j+1] - table.Diff[i-1][j]
		}
	}
}

// Полином Лагранжа
func lagrange(table Table, x float64) float64 {
	n := len(table.X)
	sum := 0.0

	for i := 0; i < n; i++ {
		term := table.Y[i]
		for j := 0; j < n; j++ {
			if j != i {
				term *= (x - table.X[j]) / (table.X[i] - table.X[j])
			}
		}
		sum += term
	}

	return sum
}

// Разделенные разности для полинома Ньютона
func dividedDifferences(table Table) []float64 {
	n := len(table.X)
	coef := make([]float64, n)
	copy(coef, table.Y)

	for j := 1; j < n; j++ {
		for i := n - 1; i >= j; i-- {
			coef[i] = (coef[i] - coef[i-1]) / (table.X[i] - table.X[i-j])
		}
	}

	return coef
}

// Полином Ньютона (разделенные разности)
func newtonDivided(table Table, x float64, coef []float64) float64 {
	n := len(coef)
	result := coef[n-1]

	for i := n - 2; i >= 0; i-- {
		result = result*(x-table.X[i]) + coef[i]
	}

	return result
}

// Первая интерполяционная формула Ньютона (вперед)
func newtonForward(table Table, x float64, h float64) float64 {
	if table.Diff == nil {
		buildDifferenceTable(&table)
	}

	t := (x - table.X[0]) / h
	result := table.Diff[0][0]
	term := 1.0

	for i := 1; i < len(table.X); i++ {
		term *= (t - float64(i-1)) / float64(i)
		result += term * table.Diff[i][0]
	}

	return result
}

// Вторая интерполяционная формула Ньютона (назад)
func newtonBackward(table Table, x float64, h float64) float64 {
	if table.Diff == nil {
		buildDifferenceTable(&table)
	}

	n := len(table.X) - 1
	t := (x - table.X[n]) / h
	result := table.Diff[0][n]
	term := 1.0

	for i := 1; i <= n; i++ {
		term *= (t + float64(i-1)) / float64(i)
		result += term * table.Diff[i][n-i]
	}

	return result
}

// Формула Гаусса (первая)
func gaussFirst(table Table, x float64, h float64) float64 {
	if table.Diff == nil {
		buildDifferenceTable(&table)
	}

	// Центральный индекс
	m := len(table.X) / 2
	u := (x - table.X[m]) / h

	// Начинаем с центрального значения
	result := table.Diff[0][m]

	// Первая разность: среднее соседних
	result += u * (table.Diff[1][m] + table.Diff[1][m-1]) / 2

	// Вторая разность
	result += u * (u - 1) * table.Diff[2][m-1] / 2

	// Третья разность: среднее
	result += u * (u - 1) * (u + 1) * (table.Diff[3][m-1] + table.Diff[3][m-2]) / 12

	// Четвертая разность
	result += u * (u - 1) * (u*u - 4) * table.Diff[4][m-2] / 24

	return result
}

// Формула Стирлинга
func stirling(table Table, x float64, h float64) float64 {
	if table.Diff == nil {
		buildDifferenceTable(&table)
	}

	m := len(table.X) / 2
	u := (x - table.X[m]) / h

	result := table.Diff[0][m]

	// Первая разность: среднее
	result += u * (table.Diff[1][m] + table.Diff[1][m-1]) / 2

	// Вторая разность
	result += u * u * table.Diff[2][m-1] / 2

	// Третья разность: среднее
	result += u * (u*u - 1) * (table.Diff[3][m-1] + table.Diff[3][m-2]) / 12

	// Четвертая разность
	result += u * u * (u*u - 1) * table.Diff[4][m-2] / 24

	return result
}

// Формула Бесселя
func bessel(table Table, x float64, h float64) float64 {
	if table.Diff == nil {
		buildDifferenceTable(&table)
	}

	// Узел m и m+1 окружают точку
	m := len(table.X)/2 - 1

	// u относительно середины между m и m+1
	u := (x - (table.X[m]+table.X[m+1])/2) / h

	// Начальное значение - среднее
	result := (table.Diff[0][m] + table.Diff[0][m+1]) / 2

	// Первая разность
	result += u * table.Diff[1][m]

	// Вторая разность
	result += (u*u - 1.0/4.0) * (table.Diff[2][m] + table.Diff[2][m-1]) / 4

	// Третья разность
	result += u * (u*u - 1) * table.Diff[3][m-1] / 6

	return result
}

// Построение графиков всех методов
func createPlots(table Table, interpolationPoints []float64, methods map[string]func(float64) float64) {
	// 1. Полный график со всеми методами
	createFullPlot(table, interpolationPoints, methods, "interpolation_full.png")

	// 2. Отдельные графики для каждого метода
	for name, method := range methods {
		createSingleMethodPlot(table, interpolationPoints, name, method,
			fmt.Sprintf("interpolation_%s.png", strings.ToLower(strings.ReplaceAll(name, " ", "_"))))
	}

	// 3. Увеличенный график (zoom) между первыми точками
	if len(table.X) >= 2 {
		createZoomPlot(table, interpolationPoints, methods, "interpolation_zoom.png")
	}
}

// Полный график со всеми методами
func createFullPlot(table Table, interpolationPoints []float64, methods map[string]func(float64) float64, filename string) {
	p := plot.New()

	p.Title.Text = "Интерполяция функции (все методы)"
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	// Исходные точки
	pts := make(plotter.XYs, len(table.X))
	for i := range table.X {
		pts[i].X = table.X[i]
		pts[i].Y = table.Y[i]
	}

	scatter, _ := plotter.NewScatter(pts)
	scatter.GlyphStyle.Color = color.RGBA{0, 0, 0, 255}
	scatter.GlyphStyle.Radius = vg.Points(3)
	p.Add(scatter)

	// Точки интерполяции (если есть)
	if len(interpolationPoints) > 0 {
		intPts := make(plotter.XYs, len(interpolationPoints))
		for i, x := range interpolationPoints {
			// Берем первый метод для вычисления значения
			var y float64
			for _, method := range methods {
				y = method(x)
				break
			}
			intPts[i].X = x
			intPts[i].Y = y
		}

		interpScatter, _ := plotter.NewScatter(intPts)
		interpScatter.GlyphStyle.Color = color.RGBA{255, 0, 0, 255}
		interpScatter.GlyphStyle.Radius = vg.Points(4)
		interpScatter.GlyphStyle.Shape = draw.CrossGlyph{}
		p.Add(interpScatter)
		p.Legend.Add("Точки интерполяции", interpScatter)
	}

	// Кривые интерполяции
	colors := []color.RGBA{
		{R: 0, G: 0, B: 255, A: 255},   // Синий
		{R: 0, G: 128, B: 0, A: 255},   // Зеленый
		{R: 255, G: 165, B: 0, A: 255}, // Оранжевый
		{R: 128, G: 0, B: 128, A: 255}, // Фиолетовый
		{R: 255, G: 0, B: 0, A: 255},   // Красный
		{R: 0, G: 128, B: 128, A: 255}, // Бирюзовый
	}

	lineStyles := []draw.LineStyle{
		{Width: vg.Points(1.5), Color: colors[0], Dashes: []vg.Length{}},                                                       // Сплошная
		{Width: vg.Points(1.5), Color: colors[1], Dashes: []vg.Length{vg.Points(4), vg.Points(2)}},                             // Штриховая
		{Width: vg.Points(1.5), Color: colors[2], Dashes: []vg.Length{vg.Points(2), vg.Points(2)}},                             // Пунктирная
		{Width: vg.Points(1.5), Color: colors[3], Dashes: []vg.Length{vg.Points(6), vg.Points(2)}},                             // Штрих-пунктирная
		{Width: vg.Points(1.5), Color: colors[4], Dashes: []vg.Length{vg.Points(2), vg.Points(2), vg.Points(6), vg.Points(2)}}, // Сложная
		{Width: vg.Points(1.5), Color: colors[5], Dashes: []vg.Length{vg.Points(1), vg.Points(2), vg.Points(4), vg.Points(2)}}, // Еще одна сложная
	}

	// Диапазон для графика
	xMin, xMax := table.X[0], table.X[0]
	for _, x := range table.X {
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
	}

	// Небольшое расширение диапазона для графика
	padding := (xMax - xMin) * 0.05
	xMin -= padding
	xMax += padding

	// Добавление кривых
	i := 0
	for name, method := range methods {
		// Точки для кривой
		n := 500 // Количество точек для плавной кривой
		curve := make(plotter.XYs, n)
		dx := (xMax - xMin) / float64(n-1)

		for j := 0; j < n; j++ {
			x := xMin + float64(j)*dx
			curve[j].X = x
			curve[j].Y = method(x)
		}

		line, _ := plotter.NewLine(curve)
		line.LineStyle = lineStyles[i%len(lineStyles)]
		p.Add(line)
		p.Legend.Add(name, line)

		i++
	}

	// Добавление исходных точек в легенду
	p.Legend.Add("Исходные точки", scatter)

	// Настройки графика
	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	// Сохранение
	if err := p.Save(10*vg.Inch, 6*vg.Inch, filename); err != nil {
		fmt.Printf("Ошибка сохранения графика: %v\n", err)
	} else {
		fmt.Printf("График сохранен: %s\n", filename)
	}
}

// График для отдельного метода
func createSingleMethodPlot(table Table, interpolationPoints []float64, methodName string,
	method func(float64) float64, filename string) {
	p := plot.New()

	p.Title.Text = fmt.Sprintf("Интерполяция: %s", methodName)
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	// Исходные точки
	pts := make(plotter.XYs, len(table.X))
	for i := range table.X {
		pts[i].X = table.X[i]
		pts[i].Y = table.Y[i]
	}

	scatter, _ := plotter.NewScatter(pts)
	scatter.GlyphStyle.Color = color.RGBA{0, 0, 0, 255}
	scatter.GlyphStyle.Radius = vg.Points(3)
	p.Add(scatter)

	// Точки интерполяции (если есть)
	if len(interpolationPoints) > 0 {
		intPts := make(plotter.XYs, len(interpolationPoints))
		for i, x := range interpolationPoints {
			intPts[i].X = x
			intPts[i].Y = method(x)
		}

		interpScatter, _ := plotter.NewScatter(intPts)
		interpScatter.GlyphStyle.Color = color.RGBA{255, 0, 0, 255}
		interpScatter.GlyphStyle.Radius = vg.Points(4)
		interpScatter.GlyphStyle.Shape = draw.CrossGlyph{}
		p.Add(interpScatter)
		p.Legend.Add("Точки интерполяции", interpScatter)
	}

	// Диапазон для графика
	xMin, xMax := table.X[0], table.X[0]
	for _, x := range table.X {
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
	}

	// Небольшое расширение диапазона для графика
	padding := (xMax - xMin) * 0.05
	xMin -= padding
	xMax += padding

	// Точки для кривой
	n := 500 // Количество точек для плавной кривой
	curve := make(plotter.XYs, n)
	dx := (xMax - xMin) / float64(n-1)

	for j := 0; j < n; j++ {
		x := xMin + float64(j)*dx
		curve[j].X = x
		curve[j].Y = method(x)
	}

	line, _ := plotter.NewLine(curve)
	line.LineStyle.Width = vg.Points(1.5)
	line.LineStyle.Color = color.RGBA{0, 0, 255, 255}
	p.Add(line)

	// Легенда
	p.Legend.Add(methodName, line)
	p.Legend.Add("Исходные точки", scatter)

	// Настройки графика
	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	// Сохранение
	if err := p.Save(8*vg.Inch, 6*vg.Inch, filename); err != nil {
		fmt.Printf("Ошибка сохранения графика: %v\n", err)
	} else {
		fmt.Printf("График сохранен: %s\n", filename)
	}
}

// Увеличенный график части интервала
func createZoomPlot(table Table, interpolationPoints []float64, methods map[string]func(float64) float64, filename string) {
	p := plot.New()

	p.Title.Text = "Увеличенный график интерполяции"
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	// Выбираем диапазон для увеличения (первые 30% интервала)
	xMin, xMax := table.X[0], table.X[0]
	for _, x := range table.X {
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
	}

	// Диапазон увеличения - начало интервала
	xZoomMax := xMin + (xMax-xMin)*0.3

	// Подбираем точки, попадающие в диапазон увеличения
	var zoomPts plotter.XYs
	for i, x := range table.X {
		if x <= xZoomMax {
			zoomPts = append(zoomPts, struct{ X, Y float64 }{x, table.Y[i]})
		}
	}

	scatter, _ := plotter.NewScatter(zoomPts)
	scatter.GlyphStyle.Color = color.RGBA{0, 0, 0, 255}
	scatter.GlyphStyle.Radius = vg.Points(3)
	p.Add(scatter)

	// Точки интерполяции в диапазоне
	var intPtsInZoom plotter.XYs
	for _, x := range interpolationPoints {
		if x <= xZoomMax {
			// Берем первый метод для значения
			var y float64
			for _, method := range methods {
				y = method(x)
				break
			}
			intPtsInZoom = append(intPtsInZoom, struct{ X, Y float64 }{x, y})
		}
	}

	if len(intPtsInZoom) > 0 {
		interpScatter, _ := plotter.NewScatter(intPtsInZoom)
		interpScatter.GlyphStyle.Color = color.RGBA{255, 0, 0, 255}
		interpScatter.GlyphStyle.Radius = vg.Points(4)
		interpScatter.GlyphStyle.Shape = draw.CrossGlyph{}
		p.Add(interpScatter)
		p.Legend.Add("Точки интерполяции", interpScatter)
	}

	// Кривые интерполяции
	colors := []color.RGBA{
		{R: 0, G: 0, B: 255, A: 255},   // Синий
		{R: 0, G: 128, B: 0, A: 255},   // Зеленый
		{R: 255, G: 165, B: 0, A: 255}, // Оранжевый
		{R: 128, G: 0, B: 128, A: 255}, // Фиолетовый
		{R: 255, G: 0, B: 0, A: 255},   // Красный
		{R: 0, G: 128, B: 128, A: 255}, // Бирюзовый
	}

	lineStyles := []draw.LineStyle{
		{Width: vg.Points(1.5), Color: colors[0], Dashes: []vg.Length{}},                                                       // Сплошная
		{Width: vg.Points(1.5), Color: colors[1], Dashes: []vg.Length{vg.Points(4), vg.Points(2)}},                             // Штриховая
		{Width: vg.Points(1.5), Color: colors[2], Dashes: []vg.Length{vg.Points(2), vg.Points(2)}},                             // Пунктирная
		{Width: vg.Points(1.5), Color: colors[3], Dashes: []vg.Length{vg.Points(6), vg.Points(2)}},                             // Штрих-пунктирная
		{Width: vg.Points(1.5), Color: colors[4], Dashes: []vg.Length{vg.Points(2), vg.Points(2), vg.Points(6), vg.Points(2)}}, // Сложная
		{Width: vg.Points(1.5), Color: colors[5], Dashes: []vg.Length{vg.Points(1), vg.Points(2), vg.Points(4), vg.Points(2)}}, // Еще одна сложная
	}

	// Небольшое расширение диапазона для графика
	padding := (xZoomMax - xMin) * 0.05
	xPlotMin := xMin - padding
	xPlotMax := xZoomMax + padding

	// Добавление кривых
	i := 0
	for name, method := range methods {
		// Точки для кривой
		n := 300 // Количество точек для плавной кривой
		curve := make(plotter.XYs, n)
		dx := (xPlotMax - xPlotMin) / float64(n-1)

		for j := 0; j < n; j++ {
			x := xPlotMin + float64(j)*dx
			curve[j].X = x
			curve[j].Y = method(x)
		}

		line, _ := plotter.NewLine(curve)
		line.LineStyle = lineStyles[i%len(lineStyles)]
		p.Add(line)
		p.Legend.Add(name, line)

		i++
	}

	// Добавление исходных точек в легенду
	p.Legend.Add("Исходные точки", scatter)

	// Настройки графика
	p.Legend.Top = true
	p.Legend.Left = true
	p.Add(plotter.NewGrid())

	// Сохранение
	if err := p.Save(8*vg.Inch, 6*vg.Inch, filename); err != nil {
		fmt.Printf("Ошибка сохранения графика: %v\n", err)
	} else {
		fmt.Printf("График сохранен: %s\n", filename)
	}
}

// Чтение данных интерактивно
func readTableInteractive() (Table, error) {
	var table Table

	fmt.Println("Введите точки (x y), пустая строка для завершения:")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return table, errors.New("ошибка: необходимо два числа")
		}

		x, err1 := strconv.ParseFloat(fields[0], 64)
		y, err2 := strconv.ParseFloat(fields[1], 64)

		if err1 != nil || err2 != nil {
			return table, errors.New("ошибка: неправильный формат чисел")
		}

		table.X = append(table.X, x)
		table.Y = append(table.Y, y)
	}

	if len(table.X) < 2 {
		return table, errors.New("необходимо минимум 2 точки")
	}

	// Сортировка по X
	indices := make([]int, len(table.X))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return table.X[indices[i]] < table.X[indices[j]]
	})

	sortedX := make([]float64, len(table.X))
	sortedY := make([]float64, len(table.Y))

	for i, idx := range indices {
		sortedX[i] = table.X[idx]
		sortedY[i] = table.Y[idx]
	}

	table.X = sortedX
	table.Y = sortedY

	// Проверка на дубликаты
	if err := checkDuplicates(table.X); err != nil {
		return table, err
	}

	return table, nil
}

// Основная функция
func main() {
	// Парсинг аргументов
	testFlag := flag.Bool("test", false, "Запустить тестовые наборы данных")
	bonusFlag := flag.Bool("bonus", false, "Включить Стирлинг и Бессель")
	flag.Parse()

	// Проведение тестов
	if *testFlag {
		runTests(*bonusFlag)
		return
	}

	// Таблица варианта 4 по умолчанию
	defaultTable := Table{
		X: []float64{1.05, 1.15, 1.25, 1.35, 1.45, 1.55, 1.65},
		Y: []float64{0.1213, 1.1316, 2.1459, 3.1565, 4.1571, 5.1819, 6.1969},
	}

	// Точки интерполяции из задания
	X1 := 1.051 // Первая формула Ньютона
	X2 := 1.277 // Первая формула Гаусса

	var table Table

	fmt.Print("Использовать таблицу 1.4 (вариант 4) по умолчанию? (y/n): ")
	var choice string
	fmt.Scan(&choice)

	if strings.ToLower(choice) == "y" || strings.ToLower(choice) == "д" {
		table = defaultTable
		fmt.Println("Используется таблица 1.4 (вариант 4)")
	} else {
		// Чтение таблицы с клавиатуры
		var err error
		table, err = readTableInteractive()
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}
	}

	// Построение таблицы конечных разностей
	buildDifferenceTable(&table)

	// Вывод исходной таблицы
	fmt.Println("\nИсходная таблица:")
	fmt.Println("   i       x          y")
	for i := range table.X {
		fmt.Printf("%4d %10.6f %10.6f\n", i, table.X[i], table.Y[i])
	}

	// Вывод таблицы конечных разностей
	fmt.Println("\nТаблица конечных разностей:")
	fmt.Printf("%10s ", "x\\Δ^k")
	for k := 0; k < len(table.X); k++ {
		fmt.Printf("%10s ", fmt.Sprintf("Δ^%d", k))
	}
	fmt.Println()

	for i := 0; i < len(table.X); i++ {
		fmt.Printf("%10.6f ", table.X[i])
		for j := 0; j < len(table.Diff); j++ {
			if i < len(table.Diff[j]) {
				fmt.Printf("%10.6f ", table.Diff[j][i])
			} else {
				fmt.Printf("%10s ", "-")
			}
		}
		fmt.Println()
	}

	// Подготовка методов интерполяции
	methods := make(map[string]func(float64) float64)

	// Метод Лагранжа
	methods["Лагранж"] = func(x float64) float64 {
		return lagrange(table, x)
	}

	// Метод Ньютона (разделенные разности)
	coef := dividedDifferences(table)
	methods["Ньютон (div)"] = func(x float64) float64 {
		return newtonDivided(table, x, coef)
	}

	// Проверка на равномерную сетку
	h, uniformGrid := checkUniformGrid(table.X)
	if uniformGrid {
		fmt.Printf("\nРавномерная сетка с шагом h = %.6f\n", h)

		// Метод Ньютона (вперед)
		methods["Ньютон (вперед)"] = func(x float64) float64 {
			return newtonForward(table, x, h)
		}

		// Первая формула Гаусса (если достаточно точек)
		if len(table.X) >= 5 {
			methods["Гаусс I"] = func(x float64) float64 {
				return gaussFirst(table, x, h)
			}

			// Дополнительные методы (если запрошены)
			if *bonusFlag && len(table.X)%2 == 1 { // Нечетное число точек для центрального узла
				methods["Стирлинг"] = func(x float64) float64 {
					return stirling(table, x, h)
				}

				methods["Бессель"] = func(x float64) float64 {
					return bessel(table, x, h)
				}
			}
		}
	} else {
		fmt.Println("\nНеравномерная сетка - методы конечных разностей недоступны")
	}

	// Запрос точки интерполяции или использование X1, X2
	fmt.Print("\nВведите точку интерполяции (или Enter для использования X1=1.051, X2=1.277): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	var interpolationPoints []float64

	if input == "" {
		interpolationPoints = []float64{X1, X2}
		fmt.Printf("Используются точки X1=%.3f, X2=%.3f\n", X1, X2)

		// Вывод результатов интерполяции для X1
		fmt.Printf("\nИнтерполяция в точке X1 = %.6f:\n", X1)
		for name, method := range methods {
			fmt.Printf("  %-15s: %.12f\n", name, method(X1))
		}

		// Вывод результатов интерполяции для X2
		fmt.Printf("\nИнтерполяция в точке X2 = %.6f:\n", X2)
		for name, method := range methods {
			fmt.Printf("  %-15s: %.12f\n", name, method(X2))
		}
	} else {
		x, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("Ошибка в вводе числа. Используется X1=1.051")
			x = X1
		}

		interpolationPoints = []float64{x}

		// Вывод результатов
		fmt.Printf("\nИнтерполяция в точке X = %.6f:\n", x)
		for name, method := range methods {
			fmt.Printf("  %-15s: %.12f\n", name, method(x))
		}
	}

	// Построение графиков
	createPlots(table, interpolationPoints, methods)
}

// Тестовые наборы
func runTests(bonus bool) {
	tests := []struct {
		name  string
		table Table
		x     float64
	}{
		{
			"Вариант 4",
			Table{
				X: []float64{1.05, 1.15, 1.25, 1.35, 1.45, 1.55, 1.65},
				Y: []float64{0.1213, 1.1316, 2.1459, 3.1565, 4.1571, 5.1819, 6.1969},
			},
			1.277,
		},
		{
			"Синус-таблица",
			func() Table {
				var t Table
				t.X = make([]float64, 9)
				t.Y = make([]float64, 9)

				for i := range t.X {
					t.X[i] = -0.4 + float64(i)*0.1
					t.Y[i] = math.Sin(t.X[i])
				}
				return t
			}(),
			math.Pi / 4,
		},
		{
			"Неравный шаг",
			Table{
				X: []float64{0, 0.3, 0.55, 0.8, 1.0},
				Y: []float64{0, 0.3, 0.522, 0.717, 0.842},
			},
			0.4,
		},
		{
			"Дубликат x",
			Table{
				X: []float64{0, 0.3, 0.3, 0.6},
				Y: []float64{0, 0.3, 0.3, 0.564},
			},
			0.32,
		},
	}

	for _, test := range tests {
		fmt.Printf("\n--- %s ---\n", test.name)

		// Проверка на дубликаты
		if err := checkDuplicates(test.table.X); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			continue
		}

		// Построение таблицы разностей
		buildDifferenceTable(&test.table)

		// Вычисление по методу Лагранжа
		lagrangeResult := lagrange(test.table, test.x)
		fmt.Printf("Лагранж      : %.12f\n", lagrangeResult)

		// Вычисление по методу Ньютона (разделенные разности)
		coef := dividedDifferences(test.table)
		newtonDivResult := newtonDivided(test.table, test.x, coef)
		fmt.Printf("Ньютон (div) : %.12f\n", newtonDivResult)

		// Проверка на равномерную сетку
		h, uniformGrid := checkUniformGrid(test.table.X)

		if uniformGrid {
			// Метод Ньютона (вперед)
			fmt.Printf("Ньютон (вперед) : %.12f\n", newtonForward(test.table, test.x, h))

			// Метод Гаусса (если достаточно точек)
			if len(test.table.X) >= 5 {
				fmt.Printf("Гаусс I       : %.12f\n", gaussFirst(test.table, test.x, h))

				// Дополнительные методы
				if bonus && len(test.table.X)%2 == 1 {
					fmt.Printf("Стирлинг     : %.12f\n", stirling(test.table, test.x, h))
					fmt.Printf("Бессель      : %.12f\n", bessel(test.table, test.x, h))
				}
			}
		} else {
			fmt.Println("(неравный шаг - методы конечных разностей недоступны)")
		}
	}
}
