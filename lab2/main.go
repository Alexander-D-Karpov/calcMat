package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Функции для уравнения: x^3 - 1.89x^2 - 2x + 1.76 = 0
func f1(x float64) float64 {
	return math.Pow(x, 3) - 1.89*math.Pow(x, 2) - 2*x + 1.76
}

func df1(x float64) float64 {
	return 3*math.Pow(x, 2) - 3.78*x - 2
}

// Для метода простой итерации (с параметром alpha)
func phi1(x float64, alpha float64) float64 {
	return x - alpha*f1(x)
}

func dphi1(x float64, alpha float64) float64 {
	return 1 - alpha*df1(x)
}

// Дополнительные тестовые функции
func f2(x float64) float64 {
	return math.Sin(x) - 0.5*x
}

func df2(x float64) float64 {
	return math.Cos(x) - 0.5
}

func f3(x float64) float64 {
	return math.Pow(x, 2) - math.Log(x+1)
}

func df3(x float64) float64 {
	return 2*x - 1/(x+1)
}

// Функции для системы нелинейных уравнений:
// tan(xy+0.3)= x², 0.9x²+2y²=1
func sys1F1(x, y float64) float64 {
	return math.Tan(x*y+0.3) - x*x
}

func sys1F2(x, y float64) float64 {
	return 0.9*math.Pow(x, 2) + 2*math.Pow(y, 2) - 1
}

// Частные производные для системы
func sys1DF1dx(x, y float64) float64 {
	// Производная tan(u) равна sec^2(u)*u'
	return y*(1/(math.Cos(x*y+0.3)*math.Cos(x*y+0.3))) - 2*x
}

func sys1DF1dy(x, y float64) float64 {
	return x * (1 / (math.Cos(x*y+0.3) * math.Cos(x*y+0.3)))
}

func sys1DF2dx(x, y float64) float64 {
	return 1.8 * x
}

func sys1DF2dy(x, y float64) float64 {
	return 4 * y
}

// Метод половинного деления
func bisectionMethod(f func(float64) float64, a, b, eps float64) (float64, int, [][]float64) {
	var iterations [][]float64

	fa := f(a)
	fb := f(b)
	if fa*fb >= 0 {
		fmt.Println("Ошибка: f(a) и f(b) должны иметь разные знаки")
		return 0, 0, iterations
	}

	iteration := 0
	for math.Abs(b-a) > eps {
		iteration++
		c := (a + b) / 2
		fc := f(c)
		iterations = append(iterations, []float64{a, b, c, fa, fb, fc, math.Abs(b - a)})

		if fc == 0 {
			return c, iteration, iterations
		}

		if fa*fc < 0 {
			b = c
			fb = fc
		} else {
			a = c
			fa = fc
		}

		if iteration > 100 {
			fmt.Println("Предупреждение: достигнуто максимальное число итераций")
			break
		}
	}

	return (a + b) / 2, iteration, iterations
}

// Метод хорд
func chordMethod(f func(float64) float64, a, b, eps float64) (float64, int, [][]float64) {
	var iterations [][]float64

	fa := f(a)
	fb := f(b)
	if fa*fb >= 0 {
		fmt.Println("Ошибка: f(a) и f(b) должны иметь разные знаки")
		return 0, 0, iterations
	}

	x := a
	var prevX float64
	iteration := 0

	for {
		iteration++
		prevX = x
		x = a - fa*(b-a)/(fb-fa)
		fx := f(x)
		iterations = append(iterations, []float64{a, b, x, fa, fb, fx, math.Abs(x - prevX)})

		if math.Abs(x-prevX) < eps {
			break
		}

		if fa*fx < 0 {
			b = x
			fb = fx
		} else {
			a = x
			fa = fx
		}

		if iteration > 100 {
			fmt.Println("Предупреждение: достигнуто максимальное число итераций")
			break
		}
	}

	return x, iteration, iterations
}

// Метод Ньютона для одного уравнения
func newtonMethod(f, df func(float64) float64, x0, eps float64) (float64, int, [][]float64) {
	var iterations [][]float64

	x := x0
	iteration := 0

	for {
		iteration++
		fx := f(x)
		dfx := df(x)
		if math.Abs(dfx) < 1e-10 {
			fmt.Println("Предупреждение: производная близка к нулю, метод может не сходиться")
			return x, iteration, iterations
		}

		xNew := x - fx/dfx
		iterations = append(iterations, []float64{x, fx, dfx, xNew, math.Abs(xNew - x)})

		if math.Abs(xNew-x) < eps {
			x = xNew
			break
		}

		x = xNew

		if iteration > 100 {
			fmt.Println("Предупреждение: достигнуто максимальное число итераций")
			break
		}
	}

	return x, iteration, iterations
}

// Метод секущих
func secantMethod(f func(float64) float64, x0, x1, eps float64) (float64, int, [][]float64) {
	var iterations [][]float64

	iterCount := 0
	for {
		iterCount++
		fx0 := f(x0)
		fx1 := f(x1)
		if math.Abs(fx1-fx0) < 1e-10 {
			fmt.Println("Предупреждение: разница значений функции мала, метод может не сходиться")
			return x1, iterCount, iterations
		}

		xNew := x1 - fx1*(x1-x0)/(fx1-fx0)
		iterations = append(iterations, []float64{x0, x1, xNew, f(xNew), math.Abs(xNew - x1)})

		if math.Abs(xNew-x1) < eps {
			x1 = xNew
			break
		}
		x0, x1 = x1, xNew

		if iterCount > 100 {
			fmt.Println("Предупреждение: достигнуто максимальное число итераций")
			break
		}
	}

	return x1, iterCount, iterations
}

// Метод простой итерации для одного уравнения
func simpleIterationMethod(f func(float64) float64, x0, eps, alpha float64) (float64, int, [][]float64, bool) {
	// Проверка условия сходимости
	a := x0 - 1
	b := x0 + 1
	maxDPhi := 0.0
	for x := a; x <= b; x += (b - a) / 100 {
		val := math.Abs(dphi1(x, alpha))
		if val > maxDPhi {
			maxDPhi = val
		}
	}
	convergent := maxDPhi < 1.0

	var iterations [][]float64
	x := x0
	iterCount := 0

	for {
		iterCount++
		xNew := phi1(x, alpha)
		iterations = append(iterations, []float64{x, xNew, f(xNew), math.Abs(xNew - x)})

		if math.Abs(xNew-x) < eps {
			x = xNew
			break
		}
		x = xNew

		if iterCount > 100 || math.IsNaN(x) || math.Abs(x) > 1e10 {
			fmt.Println("Предупреждение: метод расходится или достигнуто максимальное число итераций")
			break
		}
	}

	return x, iterCount, iterations, convergent
}

// Метод Ньютона для системы нелинейных уравнений
func newtonSystemMethod(
	f1 func(float64, float64) float64,
	f2 func(float64, float64) float64,
	df1dx func(float64, float64) float64,
	df1dy func(float64, float64) float64,
	df2dx func(float64, float64) float64,
	df2dy func(float64, float64) float64,
	x0, y0, eps float64) (float64, float64, int, [][]float64) {

	var iterations [][]float64

	x, y := x0, y0
	iterCount := 0

	for {
		iterCount++
		fVal1 := f1(x, y)
		fVal2 := f2(x, y)

		j11 := df1dx(x, y)
		j12 := df1dy(x, y)
		j21 := df2dx(x, y)
		j22 := df2dy(x, y)

		det := j11*j22 - j12*j21
		if math.Abs(det) < 1e-10 {
			fmt.Println("Предупреждение: матрица Якоби вырождена, метод может не сходиться")
			return x, y, iterCount, iterations
		}

		// Вычисляем шаг (dx, dy)
		dx := (-j22*fVal1 + j12*fVal2) / det
		dy := (j21*fVal1 - j11*fVal2) / det

		xNew := x + dx
		yNew := y + dy

		iterations = append(iterations, []float64{x, y, xNew, yNew, math.Abs(dx), math.Abs(dy)})

		if math.Hypot(dx, dy) < eps {
			x, y = xNew, yNew
			break
		}

		x, y = xNew, yNew

		if iterCount > 100 {
			fmt.Println("Предупреждение: достигнуто максимальное число итераций")
			break
		}
	}

	return x, y, iterCount, iterations
}

// Метод простой итерации для системы нелинейных уравнений
func simpleIterationSystemMethod(
	f1 func(float64, float64) float64,
	f2 func(float64, float64) float64,
	phi1 func(float64, float64) float64,
	phi2 func(float64, float64) float64,
	x0, y0, eps float64) (float64, float64, int, [][]float64) {

	var iterations [][]float64

	x, y := x0, y0
	iterCount := 0

	for {
		iterCount++
		xNew := phi1(x, y)
		yNew := phi2(x, y)

		iterations = append(iterations, []float64{x, y, xNew, yNew, math.Abs(xNew - x), math.Abs(yNew - y)})

		if math.Hypot(xNew-x, yNew-y) < eps {
			x, y = xNew, yNew
			break
		}

		x, y = xNew, yNew

		if iterCount > 100 || math.IsNaN(x) || math.IsNaN(y) || math.Abs(x) > 1e10 || math.Abs(y) > 1e10 {
			fmt.Println("Предупреждение: метод расходится или достигнуто максимальное число итераций")
			break
		}
	}

	return x, y, iterCount, iterations
}

// Функция для проверки существования корня на интервале
func rootExists(f func(float64) float64, a, b float64) bool {
	return f(a)*f(b) <= 0
}

// Функция записи результатов в файл
func writeToFile(filename string, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// Функция для форматирования двумерного слайса в таблицу
func formatTable(headers []string, data [][]float64) string {
	var sb strings.Builder

	// Заголовки
	for _, h := range headers {
		sb.WriteString(fmt.Sprintf("%-15s", h))
	}
	sb.WriteString("\n")
	// Разделительная линия
	for range headers {
		sb.WriteString("---------------")
	}
	sb.WriteString("\n")
	// Данные
	for _, row := range data {
		for _, val := range row {
			sb.WriteString(fmt.Sprintf("%-15.6f", val))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Функция для чтения числа с плавающей точкой с поддержкой значения по умолчанию
func readFloat(reader *bufio.Reader, prompt string, defaultValue float64) float64 {
	fmt.Printf("%s (по умолчанию %.6f): ", prompt, defaultValue)
	inputStr, _ := reader.ReadString('\n')
	inputStr = strings.TrimSpace(inputStr)

	// Если ввод пустой, используем значение по умолчанию
	if inputStr == "" {
		return defaultValue
	}

	// Замена запятой на точку для поддержки разных локалей
	inputStr = strings.Replace(inputStr, ",", ".", -1)

	value, err := strconv.ParseFloat(inputStr, 64)
	if err != nil || value <= 0 {
		fmt.Printf("Некорректный ввод! Используется значение по умолчанию: %.6f\n", defaultValue)
		return defaultValue
	}

	return value
}

// Функция для чтения целого числа с поддержкой значения по умолчанию
func readInt(reader *bufio.Reader, prompt string, defaultValue, minValue, maxValue int) int {
	fmt.Printf("%s (по умолчанию %d): ", prompt, defaultValue)
	inputStr, _ := reader.ReadString('\n')
	inputStr = strings.TrimSpace(inputStr)

	// Если ввод пустой, используем значение по умолчанию
	if inputStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(inputStr)
	if err != nil || value < minValue || value > maxValue {
		fmt.Printf("Некорректный ввод! Используется значение по умолчанию: %d\n", defaultValue)
		return defaultValue
	}

	return value
}

// Функция для чтения двух чисел (интервала) с поддержкой значений по умолчанию
func readInterval(reader *bufio.Reader, prompt string, defaultA, defaultB float64) (float64, float64) {
	fmt.Printf("%s (по умолчанию [%.6f, %.6f]): ", prompt, defaultA, defaultB)
	inputStr, _ := reader.ReadString('\n')
	inputStr = strings.TrimSpace(inputStr)

	// Если ввод пустой, используем значения по умолчанию
	if inputStr == "" {
		return defaultA, defaultB
	}

	parts := strings.Split(inputStr, " ")
	if len(parts) != 2 {
		fmt.Printf("Некорректный формат ввода! Используются значения по умолчанию: [%.6f, %.6f]\n", defaultA, defaultB)
		return defaultA, defaultB
	}

	// Замена запятой на точку для поддержки разных локалей
	parts[0] = strings.Replace(parts[0], ",", ".", -1)
	parts[1] = strings.Replace(parts[1], ",", ".", -1)

	a, err1 := strconv.ParseFloat(parts[0], 64)
	b, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		fmt.Printf("Некорректные значения! Используются значения по умолчанию: [%.6f, %.6f]\n", defaultA, defaultB)
		return defaultA, defaultB
	}

	return a, b
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Решение нелинейных уравнений и систем")
	fmt.Println("=======================================")
	fmt.Println("1. Решить нелинейное уравнение")
	fmt.Println("2. Решить систему нелинейных уравнений")

	choice := readInt(reader, "Введите ваш выбор (1 или 2)", 1, 1, 2)

	if choice == 1 {
		fmt.Println("\nВыберите уравнение:")
		fmt.Println("1. x^3 - 1.89x^2 - 2x + 1.76 = 0")
		fmt.Println("2. sin(x) - 0.5x = 0")
		fmt.Println("3. x^2 - ln(x+1) = 0")

		eqChoice := readInt(reader, "Введите ваш выбор (1-3)", 1, 1, 3)

		var f func(float64) float64
		var df func(float64) float64
		var eqStr string
		switch eqChoice {
		case 1:
			f = f1
			df = df1
			eqStr = "x^3 - 1.89x^2 - 2x + 1.76 = 0"
		case 2:
			f = f2
			df = df2
			eqStr = "sin(x) - 0.5x = 0"
		case 3:
			f = f3
			df = df3
			eqStr = "x^2 - ln(x+1) = 0"
		}

		fmt.Println("\nВыберите метод решения:")
		fmt.Println("1. Метод половинного деления")
		fmt.Println("2. Метод хорд")
		fmt.Println("3. Метод Ньютона")
		fmt.Println("4. Метод секущих")
		fmt.Println("5. Метод простой итерации")

		methodChoice := readInt(reader, "Введите ваш выбор (1-5)", 1, 1, 5)

		if methodChoice != 4 {
			// Для методов, требующих интервал
			a, b := readInterval(reader, "Введите интервал [a, b] (через пробел)", -1.0, 1.0)
			eps := readFloat(reader, "Введите точность", 0.0001)

			if methodChoice != 5 && !rootExists(f, a, b) {
				fmt.Println("На данном интервале нет корня! Значения функции на концах интервала имеют одинаковый знак.")
				fmt.Println("Попробуйте другой интервал.")
				a, b = readInterval(reader, "Введите интервал [a, b] (через пробел)", -2.0, 2.0)
				if !rootExists(f, a, b) {
					fmt.Println("На данном интервале тоже нет корня. Использую метод простой итерации.")
					methodChoice = 5 // Переход к методу простой итерации
				}
			}

			var root float64
			var iterCount int
			var iterData [][]float64
			var resultStr, tableStr string
			var headers []string

			switch methodChoice {
			case 1:
				fmt.Println("\nРешение методом половинного деления...")
				root, iterCount, iterData = bisectionMethod(f, a, b, eps)
				headers = []string{"a", "b", "x", "f(a)", "f(b)", "f(x)", "|b-a|"}
				resultStr = fmt.Sprintf("Уравнение: %s\nМетод: Метод половинного деления\nИнтервал: [%.6f, %.6f]\nКорень: %.6f\nf(корень): %.10f\nЧисло итераций: %d",
					eqStr, a, b, root, f(root), iterCount)
				tableStr = formatTable(headers, iterData)
			case 2:
				fmt.Println("\nРешение методом хорд...")
				root, iterCount, iterData = chordMethod(f, a, b, eps)
				headers = []string{"a", "b", "x", "f(a)", "f(b)", "f(x)", "|xₖ₊₁-xₖ|"}
				resultStr = fmt.Sprintf("Уравнение: %s\nМетод: Метод хорд\nИнтервал: [%.6f, %.6f]\nКорень: %.6f\nf(корень): %.10f\nЧисло итераций: %d",
					eqStr, a, b, root, f(root), iterCount)
				tableStr = formatTable(headers, iterData)
			case 3:
				fmt.Println("\nРешение методом Ньютона...")
				x0 := a
				if math.Abs(f(b)/df(b)) < math.Abs(f(a)/df(a)) {
					x0 = b
				}
				root, iterCount, iterData = newtonMethod(f, df, x0, eps)
				headers = []string{"xₖ", "f(xₖ)", "f'(xₖ)", "xₖ₊₁", "|xₖ₊₁-xₖ|"}
				resultStr = fmt.Sprintf("Уравнение: %s\nМетод: Метод Ньютона\nНачальное приближение: %.6f\nКорень: %.6f\nf(корень): %.10f\nЧисло итераций: %d",
					eqStr, x0, root, f(root), iterCount)
				tableStr = formatTable(headers, iterData)
			case 5:
				fmt.Println("\nРешение методом простой итерации...")
				alpha := 0.1
				x0 := (a + b) / 2
				root, iterCount, iterData, convergent := simpleIterationMethod(f, x0, eps, alpha)
				headers = []string{"xₖ", "xₖ₊₁", "f(xₖ₊₁)", "|xₖ₊₁-xₖ|"}
				convStr := "Условие сходимости выполнено"
				if !convergent {
					convStr = "Внимание: условие сходимости не выполнено"
				}
				resultStr = fmt.Sprintf("Уравнение: %s\nМетод: Метод простой итерации\n%s\nНачальное приближение: %.6f\nКорень: %.6f\nf(корень): %.10f\nЧисло итераций: %d",
					eqStr, convStr, x0, root, f(root), iterCount)
				tableStr = formatTable(headers, iterData)
			}

			fmt.Println("\nРезультаты:")
			fmt.Println(resultStr)
			fmt.Println("\nТаблица итераций:")
			fmt.Println(tableStr)

			fmt.Print("\nСохранить результаты в файл? (y/n, по умолчанию y): ")
			saveChoice, _ := reader.ReadString('\n')
			saveChoice = strings.TrimSpace(saveChoice)
			if saveChoice == "" || strings.ToLower(saveChoice) == "y" {
				filename := "nonlinear_equation_results.txt"
				content := resultStr + "\n\nТаблица итераций:\n" + tableStr
				if err := writeToFile(filename, content); err != nil {
					fmt.Println("Ошибка сохранения в файл:", err)
				} else {
					fmt.Println("Результаты сохранены в", filename)
				}
			}
		} else { // Метод секущих
			x0, x1 := readInterval(reader, "Введите два начальных приближения (через пробел)", 0.0, 1.0)
			eps := readFloat(reader, "Введите точность", 0.0001)

			fmt.Println("\nРешение методом секущих...")
			root, iterCount, iterData := secantMethod(f, x0, x1, eps)
			headers := []string{"xₖ₋₁", "xₖ", "xₖ₊₁", "f(xₖ₊₁)", "|xₖ₊₁-xₖ|"}
			resultStr := fmt.Sprintf("Уравнение: %s\nМетод: Метод секущих\nНачальные приближения: %.6f, %.6f\nКорень: %.6f\nf(корень): %.10f\nЧисло итераций: %d",
				eqStr, x0, x1, root, f(root), iterCount)
			tableStr := formatTable(headers, iterData)
			fmt.Println("\nРезультаты:")
			fmt.Println(resultStr)
			fmt.Println("\nТаблица итераций:")
			fmt.Println(tableStr)

			fmt.Print("\nСохранить результаты в файл? (y/n, по умолчанию y): ")
			saveChoice, _ := reader.ReadString('\n')
			saveChoice = strings.TrimSpace(saveChoice)
			if saveChoice == "" || strings.ToLower(saveChoice) == "y" {
				filename := "nonlinear_equation_results.txt"
				content := resultStr + "\n\nТаблица итераций:\n" + tableStr
				if err := writeToFile(filename, content); err != nil {
					fmt.Println("Ошибка сохранения в файл:", err)
				} else {
					fmt.Println("Результаты сохранены в", filename)
				}
			}
		}
	} else {
		// Решение системы нелинейных уравнений
		fmt.Println("\nВыберите систему уравнений:")
		fmt.Println("1. {tan(xy + 0.3) = x², 0.9x² + 2y² = 1}")
		fmt.Println("2. {sin(x+y) - 1.2x = 0, x² + y² = 1}")

		sysChoice := readInt(reader, "Введите ваш выбор (1-2)", 1, 1, 2)

		fmt.Println("\nВыберите метод решения:")
		fmt.Println("1. Метод Ньютона")
		fmt.Println("2. Метод простой итерации")

		sysMethodChoice := readInt(reader, "Введите ваш выбор (1-2)", 1, 1, 2)

		x0, y0 := readInterval(reader, "Введите начальные приближения x0 и y0 (через пробел)", 0.5, 0.5)
		eps := readFloat(reader, "Введите точность", 0.0001)

		var resultStr, tableStr string
		if sysMethodChoice == 1 {
			fmt.Println("\nРешение системы методом Ньютона...")
			if sysChoice == 1 {
				x, y, iterCount, iterData := newtonSystemMethod(
					sys1F1, sys1F2,
					sys1DF1dx, sys1DF1dy,
					sys1DF2dx, sys1DF2dy,
					x0, y0, eps)
				headers := []string{"xₖ", "yₖ", "xₖ₊₁", "yₖ₊₁", "|dx|", "|dy|"}
				sysName := "tan(xy + 0.3) = x², 0.9x² + 2y² = 1"
				resultStr = fmt.Sprintf("Система: %s\nМетод: Метод Ньютона\nНачальные приближения: (%.6f, %.6f)\nРешение: (%.6f, %.6f)\nНевязки: %.10f, %.10f\nЧисло итераций: %d",
					sysName, x0, y0, x, y, sys1F1(x, y), sys1F2(x, y), iterCount)
				tableStr = formatTable(headers, iterData)
			} else {
				// Пример второй системы (реализация по аналогии)
				sinxyF1 := func(x, y float64) float64 { return math.Sin(x+y) - 1.2*x }
				sinxyF2 := func(x, y float64) float64 { return math.Pow(x, 2) + math.Pow(y, 2) - 1 }
				sinxyDF1dx := func(x, y float64) float64 { return math.Cos(x+y) - 1.2 }
				sinxyDF1dy := func(x, y float64) float64 { return math.Cos(x + y) }
				sinxyDF2dx := func(x, y float64) float64 { return 2 * x }
				sinxyDF2dy := func(x, y float64) float64 { return 2 * y }

				x, y, iterCount, iterData := newtonSystemMethod(
					sinxyF1, sinxyF2,
					sinxyDF1dx, sinxyDF1dy,
					sinxyDF2dx, sinxyDF2dy,
					x0, y0, eps)
				headers := []string{"xₖ", "yₖ", "xₖ₊₁", "yₖ₊₁", "|dx|", "|dy|"}
				sysName := "sin(x+y) - 1.2x = 0, x² + y² = 1"
				resultStr = fmt.Sprintf("Система: %s\nМетод: Метод Ньютона\nНачальные приближения: (%.6f, %.6f)\nРешение: (%.6f, %.6f)\nНевязки: %.10f, %.10f\nЧисло итераций: %d",
					sysName, x0, y0, x, y, sinxyF1(x, y), sinxyF2(x, y), iterCount)
				tableStr = formatTable(headers, iterData)
			}
		} else {
			fmt.Println("\nРешение системы методом простой итерации...")
			if sysChoice == 1 {
				// Преобразования для системы 1 (выбраны по примеру из задания)
				phi1 := func(x, y float64) float64 {
					// Вычисляем y по второму уравнению
					return math.Sqrt((1 - 0.9*x*x) / 2)
				}
				phi2 := func(x, y float64) float64 {
					if math.Abs(x) < 1e-10 {
						return 0
					}
					return (math.Atan(x*x) - 0.3) / x
				}
				x, y, iterCount, iterData := simpleIterationSystemMethod(
					sys1F1, sys1F2,
					phi1, phi2,
					x0, y0, eps)
				headers := []string{"xₖ", "yₖ", "xₖ₊₁", "yₖ₊₁", "|xₖ₊₁-xₖ|", "|yₖ₊₁-yₖ|"}
				sysName := "tan(xy + 0.3) = x², 0.9x² + 2y² = 1"
				resultStr = fmt.Sprintf("Система: %s\nМетод: Метод простой итерации\nНачальные приближения: (%.6f, %.6f)\nРешение: (%.6f, %.6f)\nНевязки: %.10f, %.10f\nЧисло итераций: %d",
					sysName, x0, y0, x, y, sys1F1(x, y), sys1F2(x, y), iterCount)
				tableStr = formatTable(headers, iterData)
			} else {
				phi1 := func(x, y float64) float64 { return math.Sin(x+y) / 1.2 }
				phi2 := func(x, y float64) float64 { return math.Sqrt(1 - x*x) }
				sinxyF1 := func(x, y float64) float64 { return math.Sin(x+y) - 1.2*x }
				sinxyF2 := func(x, y float64) float64 { return math.Pow(x, 2) + math.Pow(y, 2) - 1 }
				x, y, iterCount, iterData := simpleIterationSystemMethod(
					sinxyF1, sinxyF2,
					phi1, phi2,
					x0, y0, eps)
				headers := []string{"xₖ", "yₖ", "xₖ₊₁", "yₖ₊₁", "|xₖ₊₁-xₖ|", "|yₖ₊₁-yₖ|"}
				sysName := "sin(x+y) - 1.2x = 0, x² + y² = 1"
				resultStr = fmt.Sprintf("Система: %s\nМетод: Метод простой итерации\nНачальные приближения: (%.6f, %.6f)\nРешение: (%.6f, %.6f)\nНевязки: %.10f, %.10f\nЧисло итераций: %d",
					sysName, x0, y0, x, y, sinxyF1(x, y), sinxyF2(x, y), iterCount)
				tableStr = formatTable(headers, iterData)
			}
		}

		fmt.Println("\nРезультаты:")
		fmt.Println(resultStr)
		fmt.Println("\nТаблица итераций:")
		fmt.Println(tableStr)

		fmt.Print("\nСохранить результаты в файл? (y/n, по умолчанию y): ")
		saveChoice, _ := reader.ReadString('\n')
		saveChoice = strings.TrimSpace(saveChoice)
		if saveChoice == "" || strings.ToLower(saveChoice) == "y" {
			filename := "nonlinear_system_results.txt"
			content := resultStr + "\n\nТаблица итераций:\n" + tableStr
			if err := writeToFile(filename, content); err != nil {
				fmt.Println("Ошибка сохранения в файл:", err)
			} else {
				fmt.Println("Результаты сохранены в", filename)
			}
		}
	}
}
