package bubble

import (
	"strings"

	"github.com/WhiCu/TCommit/internal/core/template"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// keyMap defines the key bindings for the application
type keyMap struct {
	Forward     key.Binding
	Back        key.Binding
	Enter       key.Binding
	Help        key.Binding
	Quit        key.Binding
	ChangeState key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.ChangeState}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help},
		{k.Quit},
		{k.Forward},
		{k.Back},
		{k.Enter},
		{k.ChangeState},
	}
}

var defaultKeys = keyMap{
	Help:        key.NewBinding(key.WithKeys("h", "?"), key.WithHelp("h", "toggle help")),
	Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Forward:     key.NewBinding(key.WithKeys("up", "right", "k", "w"), key.WithHelp("↑/➝/k/w", "move forward")),
	Back:        key.NewBinding(key.WithKeys("down", "left", "j", "s"), key.WithHelp("↓/🠔/j/s", "move back")),
	Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	ChangeState: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "change state")),
}

// UI constants
const (
	indentWidth   = 2
	indentHeight  = 3
	paddingWidth  = 3
	paddingHeight = 1
)

// UI styles
var (
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Margin(1, 0)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Margin(1, 0)
	}()
)

// model represents the application state
type model struct {
	// Data
	fileName string
	tmpl     *template.Template
	replace  map[string]string

	// UI dimensions
	width  int
	height int

	// UI styles
	windowStyle lipgloss.Style

	// State
	staticTexts       []string
	isInputFocused    bool
	inputFields       []textinput.Model
	currentInputIndex int

	// Controls
	keys keyMap

	// Help
	help help.Model
}

// initModel creates a new model with default values
func initModel(fileName string, tmpl *template.Template, replace map[string]string) tea.Model {
	h := help.New()

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(paddingHeight, paddingWidth)

	inputFields := make([]textinput.Model, 0)
	staticTexts := make([]string, 0)
	var input textinput.Model
	var text strings.Builder

	for _, node := range tmpl.Nodes {
		switch n := node.(type) {
		case *template.TextNode:
			text.WriteString(n.Text)
		case *template.VarNode:
			staticTexts = append(staticTexts, text.String())
			text.Reset()

			input = textinput.New()
			input.Prompt = ""
			input.Placeholder = n.Key
			input.Width = len(n.Key)
			if n.Default != "" {
				input.Width = len(n.Default) + 1
				input.SetValue(n.Default)
			}
			input.TextStyle = inputStyle
			input.Cursor.SetMode(cursor.CursorStatic)

			inputFields = append(inputFields, input)
		}
	}

	staticTexts = append(staticTexts, text.String())

	return model{
		fileName:          fileName,
		tmpl:              tmpl,
		replace:           replace,
		windowStyle:       windowStyle,
		staticTexts:       staticTexts,
		isInputFocused:    false,
		inputFields:       inputFields,
		currentInputIndex: 0,
		help:              h,
		keys:              defaultKeys,
	}
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.windowStyle = m.windowStyle.Width(m.width - indentWidth).Height(m.height - indentHeight)
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.ChangeState) {
			m.isInputFocused = !m.isInputFocused
			if m.isInputFocused {
				m.inputFields[m.currentInputIndex].Focus()
			} else {
				m.inputFields[m.currentInputIndex].Blur()
			}
		}

		if m.isInputFocused {
			m.inputFields[m.currentInputIndex], cmd = m.inputFields[m.currentInputIndex].Update(msg)
			cmds = append(cmds, cmd)

			// Update width based on current value
			valLen := len(m.inputFields[m.currentInputIndex].Value())
			placeholderLen := len(m.inputFields[m.currentInputIndex].Placeholder)
			// Set width to exactly match text length + cursor
			if valLen > 0 {
				m.inputFields[m.currentInputIndex].Width = valLen + 1
			} else {
				m.inputFields[m.currentInputIndex].Width = placeholderLen
			}

			break
		}

		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Forward):
			m.inputFields[m.currentInputIndex].Blur()
			m.currentInputIndex = (m.currentInputIndex + 1) % len(m.inputFields)
			m.inputFields[m.currentInputIndex].Focus()
		case key.Matches(msg, m.keys.Back):
			m.inputFields[m.currentInputIndex].Blur()
			m.currentInputIndex = (m.currentInputIndex - 1 + len(m.inputFields)) % len(m.inputFields)
			m.inputFields[m.currentInputIndex].Focus()
		case key.Matches(msg, m.keys.Enter):
			for _, inputField := range m.inputFields {
				if inputField.Value() != "" {
					m.replace[inputField.Placeholder] = inputField.Value()
				}
			}
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m model) View() string {

	text := m.buildText()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.windowStyle.Render(
			m.headerView(),
			text,
			m.footerView(),
		),
		m.help.View(m.keys),
	)
}

func (m model) buildText() string {
	var b strings.Builder

	for i, text := range m.staticTexts {
		b.WriteString(text)
		if i < len(m.inputFields) {
			b.WriteString(m.inputFields[i].View())
		}
	}

	return b.String()
}

func (m model) headerView() string {
	title := titleStyle.Render(m.fileName)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(title)-indentWidth-paddingWidth*2))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(m.inputFields[m.currentInputIndex].Placeholder)
	line := strings.Repeat("─", max(0, m.width-lipgloss.Width(info)-indentWidth-paddingWidth*2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func NewProgram(fileName string, tmpl *template.Template, replace map[string]string) *tea.Program {
	return tea.NewProgram(
		initModel(fileName, tmpl, replace),
		tea.WithAltScreen(),
	)
}
