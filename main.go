package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
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

func (m model) View() string {
	return fmt.Sprintf(
		"What’s your favorite Pokémon?\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func main() {

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get("http://localhost:9117/api/v2.0/indexers/test:passed/results?apikey=m2td3d4z7xykqyb3end2gvvvydsyi1b4&Query=deamon+slayer+s03e05&Tracker%5B%5D=1337x&Tracker%5B%5D=bitsearch&Tracker%5B%5D=eztv&Tracker%5B%5D=gamestorrents&Tracker%5B%5D=internetarchive&Tracker%5B%5D=itorrent&Tracker%5B%5D=kickasstorrents-to&Tracker%5B%5D=kickasstorrents-ws&Tracker%5B%5D=limetorrents&Tracker%5B%5D=moviesdvdr&Tracker%5B%5D=nyaasi&Tracker%5B%5D=pctorrent&Tracker%5B%5D=rutor&Tracker%5B%5D=rutracker-ru&Tracker%5B%5D=solidtorrents&Tracker%5B%5D=torrentfunk&Tracker%5B%5D=yts")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result JackettResponseReq
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
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

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
