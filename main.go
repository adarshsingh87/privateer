package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var searchTerm string
var result JackettResponseReq

func getResults() tea.Cmd {
	return func () tea.Msg {
		var url string = "http://localhost:9117/api/v2.0/indexers/test:passed/results?apikey=m2td3d4z7xykqyb3end2gvvvydsyi1b4&Query=" + strings.ToLower(strings.Join(strings.Split(searchTerm, " "), "+")) + "&Tracker%5B%5D=1337x&Tracker%5B%5D=bitsearch&Tracker%5B%5D=eztv&Tracker%5B%5D=gamestorrents&Tracker%5B%5D=internetarchive&Tracker%5B%5D=itorrent&Tracker%5B%5D=kickasstorrents-to&Tracker%5B%5D=kickasstorrents-ws&Tracker%5B%5D=limetorrents&Tracker%5B%5D=moviesdvdr&Tracker%5B%5D=nyaasi&Tracker%5B%5D=pctorrent&Tracker%5B%5D=rutor&Tracker%5B%5D=rutracker-ru&Tracker%5B%5D=solidtorrents&Tracker%5B%5D=torrentfunk&Tracker%5B%5D=yts"
	
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}
	
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Can not unmarshal JSON", err)
		}
		return tea.Quit
	}
}

func main() {
	searchTerm = strings.Join(os.Args[1:], "+")

	if searchTerm == "" || searchTerm == "null" {
		inputStage := tea.NewProgram(InputModel())
		if _, err := inputStage.Run(); err != nil {
			log.Fatal(err)
		}
	}

	spinnerStage := tea.NewProgram(spennerInitialModel())

	if _, err := spinnerStage.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(PrettyPrint(result))

}

type JackettResponseReq struct {
	Results []struct {
		Tracker      string    `json:"Tracker"`
		CategoryDesc string    `json:"CategoryDesc"`
		Title        string    `json:"Title"`
		Link         string    `json:"Link"`
		PublishDate  time.Time `json:"PublishDate"`
		Size         int       `json:"Size"`
		Seeders      int       `json:"Seeders"`
		Peers        int       `json:"Peers"`
		MagnetURI    string    `json:"MagnetUri,omitempty"`
	} `json:"Results"`
}

// JSON pretty print
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// text input config
type (
	errMsg error
)

type textInputModel struct {
	textInput textinput.Model
	err       error
}

func InputModel() textInputModel {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return textInputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			searchTerm = m.textInput.Value()
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	tea.ClearScreen()
	return fmt.Sprintf(
		"What's your favorite Pokémon?\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

// spinner Config

type spinnerModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
}

func spennerInitialModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return spinnerModel{spinner: s}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(getResults(), m.spinner.Tick)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if(len(result.Results) > 0) {
		m.quitting = true
	}
	if(m.quitting) {
		return m, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s Searching... \n\n", m.spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
}
