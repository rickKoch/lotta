package gui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	config "github.com/rickKoch/lotta/pkg/runfileconfig"
)

const (
	STEP_NAME step = iota
	STEP_DESCRIPTION
	STEP_FLAGS
	STEP_EXEC
	STEP_END
)

const (
	STEP_FLAG_QUESTION flagStep = iota
	STEP_FLAG_NAME
	STEP_FLAG_VALUE
	STEP_FLAG_REQUIRED
)

var f = flag{}

type (
	errMsg   error
	step     int
	flagStep int
)

type flag struct {
	name     string
	value    string
	required bool
}

type model struct {
	nameInput           textinput.Model
	descriptionTextarea textarea.Model
	execInput           textinput.Model
	flagNameInput       textinput.Model
	flagValueInput      textinput.Model
	senderStyle         lipgloss.Style
	choice              int
	checked             bool
	commander           config.Commander
	currentStep         step
	currentFlagStep     flagStep
	flags               []flag
	err                 error
}

func Run(commander config.Commander) error {
	p := tea.NewProgram(initialModel(commander))
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func initialModel(commander config.Commander) model {
	m := model{
		nameInput:           generateTextinput("Command alias"),
		descriptionTextarea: generateTextarea("Command description"),
		execInput:           generateTextinput("Command execution"),
		flagNameInput:       generateTextinput("Flag name"),
		flagValueInput:      generateTextinput("Flag value"),
		senderStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:                 nil,
		commander:           commander,
		currentStep:         STEP_NAME,
		currentFlagStep:     STEP_FLAG_QUESTION,
		flags:               []flag{},
	}
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentStep {
	case STEP_NAME:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case STEP_DESCRIPTION:
		m.descriptionTextarea, cmd = m.descriptionTextarea.Update(msg)
	case STEP_FLAGS:
		return updateFlags(msg, m)
	case STEP_EXEC:
		m.execInput, cmd = m.execInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.nextStep()
			if m.currentStep == STEP_END {
				flags := []*config.Flag{}
				for _, flag := range m.flags {
					flags = append(flags, &config.Flag{
						Name:     flag.name,
						Value:    flag.value,
						Required: flag.required,
					})
				}
				err := m.commander.AddCommand(m.nameInput.Value(), config.Command{
					Description: m.descriptionTextarea.Value(),
					Exec:        m.execInput.Value(),
					Flags:       flags,
				})
				if err != nil {
					m.err = err
					return m, nil
				}
				return m, tea.Quit
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, cmd
}

func updateFlags(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	if m.currentFlagStep == STEP_FLAG_QUESTION || m.currentFlagStep == STEP_FLAG_REQUIRED {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
			switch msg.String() {
			case "j", "down":
				m.choice++
				if m.choice > 1 {
					m.choice = 1
				}
			case "k", "up":
				m.choice--
				if m.choice < 0 {
					m.choice = 0
				}
			case "enter":
				m.checked = true
				if m.currentFlagStep == STEP_FLAG_QUESTION {
					if m.choice == 1 {
						m.nextStep()
						return m, nil
					}
					m.nextFlagStep()
					return m, nil
				}
				m.flags = append(m.flags, flag{
					name:     m.flagNameInput.Value(),
					value:    m.flagValueInput.Value(),
					required: m.choice == 0,
				})
				m.checked = false
				m.choice = 0
				m.nextFlagStep()
				return m, nil
			}
		}
	} else {
		var cmd tea.Cmd
		switch m.currentFlagStep {
		case STEP_FLAG_NAME:
			m.flagNameInput, cmd = m.flagNameInput.Update(msg)
		case STEP_FLAG_VALUE:
			m.flagValueInput, cmd = m.flagValueInput.Update(msg)
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				m.nextFlagStep()
			}
		case errMsg:
			m.err = msg
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

func (m *model) nextStep() {
	m.currentStep++
}

func (m *model) nextFlagStep() {
	m.currentFlagStep++
	if m.currentFlagStep > STEP_FLAG_REQUIRED {
		m.currentFlagStep = STEP_FLAG_QUESTION
		m.flagNameInput.Reset()
		m.flagValueInput.Reset()
	}
}

func (m model) View() string {
	switch m.currentStep {
	case STEP_NAME:
		return fmt.Sprintf("%s", m.nameInput.View()) + "\n\n"
	case STEP_DESCRIPTION:
		return fmt.Sprintf("%s", m.descriptionTextarea.View()) + "\n\n"
	case STEP_FLAGS:
		switch m.currentFlagStep {
		case STEP_FLAG_QUESTION:
			return m.choiceView("Do you want to add flags?\n")
		case STEP_FLAG_NAME:
			return fmt.Sprintf("%s", m.flagNameInput.View()) + "\n\n"
		case STEP_FLAG_VALUE:
			return fmt.Sprintf("%s", m.flagValueInput.View()) + "\n\n"
		case STEP_FLAG_REQUIRED:
			return m.choiceView("Is this flag required?\n")
		}
	case STEP_EXEC:
		return fmt.Sprintf("%s", m.execInput.View()) + "\n\n"
	}

	return fmt.Sprintln()
}

func (m model) choiceView(question string) string {
	c := m.choice
	choices := fmt.Sprintf("\n%s\n%s\n", checkbox("Yes", c == 0), checkbox("No", c == 1))
	return fmt.Sprintln(question, choices)
}

func checkbox(label string, checked bool) string {
	if checked {
		return fmt.Sprintf("[x] %s", label)
	}
	return fmt.Sprintf("[ ] %s", label)
}

func generateTextinput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	return ti
}

func generateTextarea(placeholder string) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)
	return ta
}
