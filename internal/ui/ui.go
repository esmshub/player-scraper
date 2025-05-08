package ui

import (
	"errors"
	"fmt"
	"os"
	"player-scraper/internal/core"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScrapeMode int

const (
	ScrapeOnly ScrapeMode = iota
	ScrapeAndDownload
	ScrapeOnlyLocal
)

var (
	pageStyle = lipgloss.NewStyle().
			Margin(2, 2)

	listStyle = lipgloss.NewStyle().
			Padding(2, 0)

	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			MarginBottom(1)
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var choices = []list.Item{
	item{title: "Scrape", desc: "Scrape latest rosters from the web"},
	item{title: "Scrape and download", desc: "Scrape latest rosters the web and download a copy"},
	item{title: "Scrape local", desc: "Scrape rosters from a local directory (include INFO data if present)"},
}

type model struct {
	cancelled bool
	step      int
	mode      ScrapeMode
	list      list.Model
	form      FormModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	curStep := m.step

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.step == 0 {
				m.mode = ScrapeMode(m.list.Index())
				m.form.inputs = getInputsForMode(m.mode)
				m.step++
			} else if m.step == 1 && m.form.IsComplete() {
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		// h, v := docStyle.GetFrameSize()
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		// formStyle = formStyle.Height(msg.Height - v).Width(msg.Width - h)
	}

	var cmd tea.Cmd
	if curStep == 0 {
		m.list, cmd = m.list.Update(msg)
		if m.step == 1 {
			focusCmd := tea.Batch(cmd, m.form.inputs[0].Focus())
			m.form, cmd = m.form.Update(msg)
			cmd = tea.Batch(focusCmd, cmd)
		}
	} else {
		m.form, cmd = m.form.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.step == 0 {
		return listStyle.Render(m.list.View())
	} else {
		return pageStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render(m.list.Title),
				fmt.Sprintf("%s: %s", filledStyle.Render("Scrape type"), filledValueStyle.Render(m.list.SelectedItem().(item).Title())),
				m.form.View(),
				// formStyle.Render(m.form.View()),
				// pageStyle.Render(m.list.Help.View(m.list)),
			),
		)
	}
}

func createInputModel(prompt, placeholder string, limit int, validator textinput.ValidateFunc) textinput.Model {
	// Create a new text input model.
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = prompt
	ti.CharLimit = limit
	ti.Width = 255
	ti.Validate = validator
	return ti
}

func validatePathDir(v string) error {
	stat, err := os.Stat(v)
	if err != nil || !stat.IsDir() {
		return errors.New("not a valid directory")
	}

	return nil
}

func validateBool(v string) error {
	if strings.ToLower(v) != "y" && strings.ToLower(v) != "n" {
		return errors.New("must be 'y' or 'n'")
	}

	return nil
}

func getInputsForMode(mode ScrapeMode) []FormInputModel {
	var cwd, err = os.Getwd()
	if err != nil {
		cwd = "."
	}

	inputs := []FormInputModel{
		{id: "outputDir", field: createInputModel("Report output dir: ", cwd, 255, validatePathDir)},
		{id: "excelExport", field: createInputModel("Use Excel formulas: ", "y", 1, validateBool)},
	}

	if mode != ScrapeOnly {
		inputs = append([]FormInputModel{
			{id: "rosterDir", field: createInputModel("Roster dir: ", cwd, 255, validatePathDir)},
		}, inputs...)
	}

	return inputs
}

func StyleTitle(appName string) string {
	return titleStyle.Render(appName)
}

func Run(appName string) (core.ScraperOptions, bool) {
	m := model{
		list: list.New(choices, list.NewDefaultDelegate(), 0, 0),
		form: FormModel{
			focusIndex: -1,
		},
	}

	m.list.Title = appName
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowPagination(false)

	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run returns the model as a tea.Model.
	output, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	mo := output.(model)
	getFormValue := func(id string) string {
		var input *FormInputModel = nil
		for _, in := range mo.form.inputs {
			if in.id == id {
				input = &in
				break
			}
		}
		if input == nil {
			return ""
		}

		value := input.Value()
		if value == "" {
			value = input.field.Placeholder
		}

		return value
	}

	opts := core.ScraperOptions{
		LocalOnly:     mo.mode == ScrapeOnlyLocal,
		DownloadFiles: mo.mode == ScrapeAndDownload,
		RosterDir:     "",
		OutputDir:     getFormValue("outputDir"),
		ExcelExport:   strings.ToLower(getFormValue("excelExport")) == "y",
	}

	if mo.mode != ScrapeOnly {
		opts.RosterDir = getFormValue("rosterDir")
	}

	return opts, mo.cancelled
}
