package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kenilt/go-wordwise-creator/filepicker"
)

type PickerModel struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
}

func (m PickerModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path

		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m PickerModel) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString("  Dir: " + m.filepicker.CurrentDirectory)
	s.WriteString("\n" + m.filepicker.View() + "\n")
	s.WriteString("  ↑↓ to browsing, ← to parent dir, ↵ Enter to choose\n")
	return s.String()
}

// clone from here https://github.com/charmbracelet/bubbletea/blob/master/examples/file-picker/main.go
func main1() {
	fp := filepicker.New()
	fp.Path, _ = os.UserHomeDir()
	fp.CurrentDirectory, _ = os.Getwd()

	m := PickerModel{
		filepicker: fp,
	}
	tm, _ := tea.NewProgram(&m, tea.WithOutput(os.Stderr)).StartReturningModel()
	mm := tm.(PickerModel)
	fmt.Println("\n  You selected: " + m.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")
}
