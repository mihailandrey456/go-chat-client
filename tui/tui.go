package tui

import (
	"fmt"
	"io"
	"net"
	"strings"

	"andrewka/chatclient/message"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

func NewProgram(conn net.Conn) *tea.Program {
	return tea.NewProgram(initialModel(conn))
}

type (
	OuterMsg message.Msg
	ErrMsg   struct {
		Err   error
		Fatal bool
	}
)

func (e ErrMsg) String() string {
	return fmt.Sprintf("%s", e.Err)
}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	errorStyle  lipgloss.Style
	conn        net.Conn
}

func initialModel(conn net.Conn) model {
	ta := textarea.New()
	ta.Placeholder = "Введите сообщение..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		errorStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		conn:        conn,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.textarea.Value()
			m.textarea.Reset()
			_, err := io.WriteString(m.conn, input+"\n")
			if err != nil {
				m.addMessage(m.errorStyle.Render("Chat Client: ") + "Сообщение не доставлено")
				m.textarea.SetValue(input)
			}
			m.viewport.GotoBottom()
		}

	case OuterMsg:
		m.addMessage(m.senderStyle.Render(msg.From+": ") + msg.Content)

	case ErrMsg:
		m.addMessage(m.errorStyle.Render("Chat Client: ") + msg.String())

		if msg.Fatal {
			return m, tea.Quit
		}
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func (m *model) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
	m.viewport.GotoBottom()
}
