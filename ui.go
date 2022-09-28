package letswatch

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type MovieItem struct {
	movie *Movie
}

func (i MovieItem) Title() string {
	return fmt.Sprintf("%v (%v)", i.movie.Title, i.movie.ReleaseYear)
}

func (i MovieItem) Description() string {
	/*
		text := fmt.Sprintf("language:%v\n", i.movie.Language)
		text = text + fmt.Sprintf("title:%v\n", i.movie.Title)
		return text
	*/
	return i.movie.IMDBLink
}

func (i MovieItem) FilterValue() string {
	text := fmt.Sprintf("language:%v\n", i.movie.Language)
	text = text + fmt.Sprintf("title:%v\n", i.movie.Title)
	return text
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func NewUI(movies []*Movie) error {
	mi := []list.Item{}
	for _, m := range movies {
		mi = append(mi, MovieItem{movie: m})
	}
	m := model{list: list.New(mi, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Movies!!"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}
