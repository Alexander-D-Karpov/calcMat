package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"calcMat/solver"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateMenu state = iota
	stateDimension
	stateMatrixInput
	stateFileInput
	statePrecision
	stateProcessing
	stateResult
)

type model struct {
	state       state
	inputMethod string
	dimension   int
	currentRow  int
	matrix      [][]float64
	vector      []float64
	precision   float64
	inputBuffer string
	errorMsg    string
	result      *solver.Result
	err         error
	blink       bool
}

func NewProgram() *tea.Program {
	return tea.NewProgram(model{
		state:  stateMenu,
		matrix: make([][]float64, 0),
		vector: make([]float64, 0),
		blink:  true,
	})
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*530, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.blink = !m.blink
		return m, tea.Tick(time.Millisecond*530, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.handleMenuInput(msg)
		case stateDimension:
			return m.handleDimensionInput(msg)
		case stateMatrixInput:
			return m.handleMatrixInput(msg)
		case stateFileInput:
			return m.handleFileInput(msg)
		case statePrecision:
			return m.handlePrecisionInput(msg)
		case stateResult:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		default:
			panic("unhandled state")
		}

	case solverMsg:
		m.result = msg.result
		m.err = msg.err
		m.state = stateResult
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case stateMenu:
		s.WriteString("╭──────────────────────────────────────────╮\n")
		s.WriteString("│         Simple Iteration Method          │\n")
		s.WriteString("╰──────────────────────────────────────────╯\n\n")
		s.WriteString("Choose input method:\n")
		s.WriteString("1 - Interactive input\n")
		s.WriteString("2 - File input\n\n")
		s.WriteString("Press 'q' to quit")

	case stateDimension:
		s.WriteString("Enter matrix dimension (1-20):\n")
		s.WriteString("Example: 2 for a 2x2 matrix\n\n")
		s.WriteString(m.inputBuffer)
		if m.blink {
			s.WriteString("█")
		}
		if m.errorMsg != "" {
			s.WriteString("\n\nError: " + m.errorMsg)
		}

	case stateMatrixInput:
		s.WriteString(fmt.Sprintf("Enter row %d of %d\n", m.currentRow+1, m.dimension))
		s.WriteString(fmt.Sprintf("Format: enter %d numbers separated by spaces\n", m.dimension+1))
		s.WriteString(fmt.Sprintf("Example: for 2x2 matrix, enter: 4 1 1\n\n"))

		// Draw matrix frame
		s.WriteString("Current matrix:\n")
		s.WriteString("╭" + strings.Repeat("─", m.dimension*8+1) + "╮ ╭" + strings.Repeat("─", 8) + "╮\n")

		for i := 0; i < m.dimension; i++ {
			s.WriteString("│")
			if i < m.currentRow {
				// Show completed rows
				for j := 0; j < m.dimension; j++ {
					s.WriteString(fmt.Sprintf(" %6.2f", m.matrix[i][j]))
				}
				s.WriteString(" │ │")
				s.WriteString(fmt.Sprintf(" %6.2f", m.vector[i]))
				s.WriteString(" │\n")
			} else if i == m.currentRow {
				// Current input row
				nums := strings.Fields(m.inputBuffer)
				for j := 0; j < m.dimension; j++ {
					if j < len(nums) {
						s.WriteString(fmt.Sprintf(" %6s", nums[j]))
					} else {
						s.WriteString("      ·")
					}
				}
				s.WriteString(" │ │")
				if len(nums) > m.dimension {
					s.WriteString(fmt.Sprintf(" %6s", nums[m.dimension]))
				} else {
					s.WriteString("      ·")
				}
				s.WriteString(" │")
				if m.blink {
					s.WriteString("█")
				}
				s.WriteString("\n")
			} else {
				// Future rows
				s.WriteString(strings.Repeat("      ·", m.dimension))
				s.WriteString(" │ │      ·")
				s.WriteString(" │\n")
			}
		}
		s.WriteString("╰" + strings.Repeat("─", m.dimension*8+1) + "╯ ╰" + strings.Repeat("─", 8) + "╯\n")

		if m.errorMsg != "" {
			s.WriteString("\nError: " + m.errorMsg)
		}

	case statePrecision:
		s.WriteString("Enter desired precision:\n")
		s.WriteString("Example: 0.0001 or 1e-4\n\n")
		s.WriteString(m.inputBuffer)
		if m.blink {
			s.WriteString("█")
		}
		if m.errorMsg != "" {
			s.WriteString("\n\nError: " + m.errorMsg)
		}

	case stateProcessing:
		s.WriteString("Processing...")

	case stateResult:
		if m.err != nil {
			s.WriteString("Error: " + m.err.Error() + "\n\n")
			if m.err.Error() == "unable to achieve diagonal dominance" {
				s.WriteString("╭──────────────────────────────────────────╮\n")
				s.WriteString("│           Diagonal Dominance            │\n")
				s.WriteString("╰──────────────────────────────────────────╯\n\n")
				s.WriteString("Your matrix doesn't satisfy this condition\n")
				s.WriteString("even after attempting row reordering.\n\n")
			}
		} else {
			s.WriteString("╭──────────────────────────────────────────╮\n")
			s.WriteString("│              Solution                    │\n")
			s.WriteString("╰──────────────────────────────────────────╯\n\n")
			s.WriteString("Solution vector:\n")
			for i, val := range m.result.Solution {
				s.WriteString(fmt.Sprintf("x%d = %12.6f\n", i+1, val))
			}
			s.WriteString(fmt.Sprintf("\nConverged in %d iterations\n", m.result.Iterations))
			s.WriteString("\nFinal errors:\n")
			for i, err := range m.result.Errors {
				s.WriteString(fmt.Sprintf("e%d = %12.6e\n", i+1, err))
			}
			s.WriteString(fmt.Sprintf("\nMatrix norm: %.6f\n", m.result.MatrixNorm))
		}
		s.WriteString("\nPress 'q' to quit")
	case stateFileInput:
		s.WriteString("╭──────────────────────────────────────────╮\n")
		s.WriteString("│              File Input                  │\n")
		s.WriteString("╰──────────────────────────────────────────╯\n\n")
		s.WriteString("Enter file path (press Enter to load): ")
		s.WriteString(m.inputBuffer)
		if m.blink {
			s.WriteString("█")
		}
		if m.errorMsg != "" {
			s.WriteString("\n\nError: " + m.errorMsg)
		}
	}

	return s.String()
}

type inputData struct {
	matrix    [][]float64
	vector    []float64
	precision float64
}

func parseInput(input string) (*inputData, error) {
	lines := strings.Split(strings.TrimSpace(input), "\n")

	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient input data")
	}

	n, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid dimension: %v", err)
	}

	if n > 20 || n <= 0 {
		return nil, fmt.Errorf("dimension must be between 1 and 20")
	}

	if len(lines) < n+2 {
		return nil, fmt.Errorf("insufficient matrix data")
	}

	matrix := make([][]float64, n)
	vector := make([]float64, n)

	for i := 0; i < n; i++ {
		nums := strings.Fields(lines[i+1])
		if len(nums) != n+1 {
			return nil, fmt.Errorf("invalid number of coefficients in row %d", i+1)
		}

		row := make([]float64, n)
		for j := 0; j < n; j++ {
			val, err := strconv.ParseFloat(nums[j], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid coefficient at row %d, col %d: %v", i+1, j+1, err)
			}
			row[j] = val
		}

		bVal, err := strconv.ParseFloat(nums[n], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid right-hand side value at row %d: %v", i+1, err)
		}

		matrix[i] = row
		vector[i] = bVal
	}

	precision, err := strconv.ParseFloat(strings.TrimSpace(lines[n+1]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid precision value: %v", err)
	}

	return &inputData{
		matrix:    matrix,
		vector:    vector,
		precision: precision,
	}, nil
}

type solverMsg struct {
	result *solver.Result
	err    error
}

func processSolution(m model) tea.Cmd {
	return func() tea.Msg {
		result, err := solver.SolveSystem(m.matrix, m.vector, m.precision)
		return solverMsg{result: result, err: err}
	}
}

func (m model) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.inputMethod = "keyboard"
		m.state = stateDimension
		m.errorMsg = ""
	case "2":
		m.inputMethod = "file"
		m.state = stateFileInput
		m.errorMsg = ""
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleDimensionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		n, err := strconv.Atoi(strings.TrimSpace(m.inputBuffer))
		if err != nil {
			m.errorMsg = "Please enter a valid number"
		} else if n <= 0 || n > 20 {
			m.errorMsg = "Dimension must be between 1 and 20"
		} else {
			m.dimension = n
			m.matrix = make([][]float64, n)
			m.vector = make([]float64, n)
			m.currentRow = 0
			m.state = stateMatrixInput
			m.errorMsg = ""
		}
		m.inputBuffer = ""
	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyCtrlC:
		return m, tea.Quit
	default:
		if len(msg.String()) == 1 && strings.Contains("0123456789", msg.String()) {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

func (m model) handleMatrixInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		numbers := strings.Fields(m.inputBuffer)
		if len(numbers) != m.dimension+1 {
			m.errorMsg = fmt.Sprintf("Please enter %d numbers (matrix row + vector element)", m.dimension+1)
			m.inputBuffer = ""
			return m, nil
		}

		row := make([]float64, m.dimension)
		for i, num := range numbers[:m.dimension] {
			val, err := strconv.ParseFloat(num, 64)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Invalid number: %s", num)
				m.inputBuffer = ""
				return m, nil
			}
			row[i] = val
		}

		bVal, err := strconv.ParseFloat(numbers[m.dimension], 64)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Invalid vector element: %s", numbers[m.dimension])
			m.inputBuffer = ""
			return m, nil
		}

		m.matrix[m.currentRow] = row
		m.vector[m.currentRow] = bVal
		m.currentRow++
		m.inputBuffer = ""
		m.errorMsg = ""

		if m.currentRow >= m.dimension {
			m.state = statePrecision
		}

	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyCtrlC:
		return m, tea.Quit
	default:
		if len(msg.String()) == 1 && (strings.Contains("0123456789.-", msg.String()) || msg.String() == " ") {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

func (m model) handlePrecisionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		val, err := strconv.ParseFloat(strings.TrimSpace(m.inputBuffer), 64)
		if err != nil {
			m.errorMsg = "Please enter a valid precision value"
			m.inputBuffer = ""
			return m, nil
		}
		if val <= 0 {
			m.errorMsg = "Precision must be greater than 0"
			m.inputBuffer = ""
			return m, nil
		}
		m.precision = val
		m.state = stateProcessing
		return m, processSolution(m)

	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyCtrlC:
		return m, tea.Quit
	default:
		if len(msg.String()) == 1 && (strings.Contains("0123456789.", msg.String()) || msg.String() == "e" || msg.String() == "-") {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

func (m model) handleFileInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		filePath := strings.TrimSpace(m.inputBuffer)
		if filePath == "" {
			m.errorMsg = "Please enter a file path"
			return m, nil
		}

		info, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				m.errorMsg = "File does not exist"
			} else {
				m.errorMsg = fmt.Sprintf("Error accessing file: %v", err)
			}
			return m, nil
		}

		if info.IsDir() {
			m.errorMsg = "Path is a directory, not a file"
			return m, nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to read file: %v", err)
			return m, nil
		}

		data, err := parseInput(string(content))
		if err != nil {
			m.errorMsg = fmt.Sprintf("Invalid file format: %v", err)
			return m, nil
		}

		m.dimension = len(data.matrix)
		m.matrix = data.matrix
		m.vector = data.vector
		m.precision = data.precision
		m.errorMsg = ""
		m.state = stateProcessing
		return m, processSolution(m)

	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyEsc:
		m.state = stateMenu
		m.inputBuffer = ""
		m.errorMsg = ""
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	default:
		switch msg.Type {
		case tea.KeySpace:
			m.inputBuffer += " "
		case tea.KeyTab:
			m.inputBuffer += "    "
		default:
			if len(msg.String()) == 1 {
				m.inputBuffer += msg.String()
			}
		}
	}
	return m, nil
}
