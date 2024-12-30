package chat

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func RenderView(model tea.Model) {
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

type View struct {
	username    string
	viewport    viewport.Model
	messages    []string
	textarea    textinput.Model
	onMessage   func(string)
	senderStyle lipgloss.Style
	err         error
}

func InitView(username string, onMessage func(string)) *View {
	ta := textinput.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room! Type a message and press Enter to send.`)
	return &View{
		textarea:    ta,
		viewport:    vp,
		username:    username,
		onMessage:   onMessage,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	}
}

func (m *View) Init() tea.Cmd {
	return textarea.Blink
}

func (m *View) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case "enter":
			v := m.textarea.Value()
			if v == "" {
				return m, nil
			}
			if m.onMessage != nil {
				m.onMessage(v)
			}
			m.messages = append(m.messages, m.senderStyle.Render(m.username+": ")+v)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			return m, nil
		default:
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m *View) View() string {
	return fmt.Sprintf("%s\n\n%s", m.viewport.View(), m.textarea.View()) + "\n"
}
