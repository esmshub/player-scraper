package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})
	blurredStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	filledStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	filledValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle       = lipgloss.NewStyle().MarginTop(2).Foreground(lipgloss.Color("9"))
	noStyle          = lipgloss.NewStyle()
)

type FormInputModel struct {
	id       string
	optional bool
	field    textinput.Model
}

func (f FormInputModel) Update(msg tea.Msg) (FormInputModel, tea.Cmd) {
	var cmd tea.Cmd
	f.field, cmd = f.field.Update(msg)
	return f, cmd
}

func (f *FormInputModel) SetDefaultValue() {
	f.field.SetValue(f.field.Placeholder)
}

func (f *FormInputModel) Value() string {
	return f.field.Value()
}

func (f *FormInputModel) Error() error {
	return f.field.Err
}

func (f *FormInputModel) Focus() tea.Cmd {
	return f.field.Focus()
}

func (f *FormInputModel) Blur() {
	f.field.Blur()
}

type FormModel struct {
	focusIndex int
	inputs     []FormInputModel
}

func (f FormModel) Init() tea.Cmd {
	return nil
}

func (f FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	cmds := make([]tea.Cmd, len(f.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		// Set focus to next input
		case tea.KeyRight:
			if f.inputs[f.focusIndex].Value() == "" {
				f.inputs[f.focusIndex].SetDefaultValue()
			}
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			s := msg.String()

			nextKey := msg.Type == tea.KeyEnter || msg.Type == tea.KeyDown || msg.Type == tea.KeyTab
			if nextKey && f.focusIndex > -1 && f.inputs[f.focusIndex].Error() != nil {
				return f, nil
			}

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && f.focusIndex == len(f.inputs)-1 {
				return f, tea.Quit
			} else if s == "enter" && f.focusIndex < len(f.inputs)-1 {
				f.focusIndex++
			}

			// Cycle indexes
			if (s == "up" || s == "shift+tab") && f.focusIndex > 0 {
				f.focusIndex--
			} else if (s == "down" || s == "tab") && f.focusIndex < len(f.inputs)-1 {
				f.focusIndex++
			}

			// if f.focusIndex > len(f.inputs)-1 {
			// 	f.focusIndex = 0
			// } else if f.focusIndex < 0 {
			// 	f.focusIndex = len(f.inputs)
			// }

			cmds := make([]tea.Cmd, len(f.inputs))
			for i := 0; i <= len(f.inputs)-1; i++ {
				if i == f.focusIndex {
					// Set focused state
					cmds[i] = f.inputs[i].Focus()
					f.inputs[i].field.PromptStyle = focusedStyle
					f.inputs[i].field.TextStyle = noStyle
					f.inputs[i].field.PlaceholderStyle = placeholderStyle
					continue
				}
				// Remove focused state
				f.inputs[i].Blur()
				if i < f.focusIndex {
					f.inputs[i].field.PromptStyle = filledStyle
					f.inputs[i].field.TextStyle = filledValueStyle
					f.inputs[i].field.PlaceholderStyle = filledValueStyle
				} else {
					f.inputs[i].field.PromptStyle = noStyle
					f.inputs[i].field.TextStyle = blurredStyle
					f.inputs[i].field.PlaceholderStyle = blurredStyle
				}
			}

			return f, tea.Batch(cmds...)
		}
	}

	for i, in := range f.inputs {
		f.inputs[i], cmds[i] = in.Update(msg)
	}
	return f, tea.Batch(cmds...)
}

func (f FormModel) View() string {
	var b strings.Builder

	for i, in := range f.inputs {
		view := in.field.View()
		b.WriteString(view)
		if i < len(f.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	if f.inputs[f.focusIndex].Error() != nil {
		b.WriteRune('\n')
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", f.inputs[f.focusIndex].Error())))
	}

	return b.String()
}

func (f FormModel) IsComplete() bool {
	complete := true
	for _, in := range f.inputs {
		if in.Error() != nil || in.Value() == "" {
			complete = false
			break
		}
	}
	return complete
}
