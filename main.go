package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/nsf/termbox-go"
)

var searchTerm string
var result JackettResponseReq
var apiKey string

type Config struct {
	ApiKey string `json:"ApiKey"`
}

func getResults() tea.Cmd {
	return func() tea.Msg {
		var url string = "http://localhost:9117/api/v2.0/indexers/test:passed/results?apikey=" + apiKey + "&Query=" + strings.ToLower(strings.Join(strings.Split(searchTerm, " "), "+")) + "&Tracker%5B%5D=1337x&Tracker%5B%5D=bitsearch&Tracker%5B%5D=eztv&Tracker%5B%5D=gamestorrents&Tracker%5B%5D=internetarchive&Tracker%5B%5D=itorrent&Tracker%5B%5D=kickasstorrents-to&Tracker%5B%5D=kickasstorrents-ws&Tracker%5B%5D=limetorrents&Tracker%5B%5D=moviesdvdr&Tracker%5B%5D=nyaasi&Tracker%5B%5D=pctorrent&Tracker%5B%5D=rutor&Tracker%5B%5D=rutracker-ru&Tracker%5B%5D=solidtorrents&Tracker%5B%5D=torrentfunk&Tracker%5B%5D=yts"

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
	appSettings, err := os.Open("privateer.json")

	if err != nil {
		apiInput := tea.NewProgram(ApiInputModel())
		if _, err := apiInput.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		var settings Config

		byteValue, _ := ioutil.ReadAll(appSettings)
		json.Unmarshal(byteValue, &settings)
		apiKey = settings.ApiKey
	}

	if apiKey == "" {
		apiInput := tea.NewProgram(ApiInputModel())
		if _, err := apiInput.Run(); err != nil {
			log.Fatal(err)
		}
	}

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

	p := tea.NewProgram(NewTableModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	// fmt.Println(PrettyPrint(result))

}

type JackettResponseReq struct {
	Results []JackettResult `json:"Results"`
}

type JackettResult struct {
	Tracker      string `json:"Tracker"`
	CategoryDesc string `json:"CategoryDesc"`
	Title        string `json:"Title"`
	Link         string `json:"Link"`
	PublishDate  string `json:"PublishDate"`
	Size         int    `json:"Size"`
	Seeders      int    `json:"Seeders"`
	MagnetURI    string `json:"MagnetUri"`
}

// JSON pretty print
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// util to row

func (p JackettResult) ToRow() table.Row {
	publishDate, _ := time.Parse(time.RFC3339, p.PublishDate)
	return table.NewRow(table.RowData{
		"Title":    p.Title,
		"Tracker":  p.Tracker,
		"Category": p.CategoryDesc,
		"Date":     publishDate,
		"Size":     fmt.Sprintf("%.2f", float32(p.Size/1024)/1024) + "Mb",
		"Seeders":  p.Seeders,

		// This isn't a visible column, but we can add the data here anyway for later retrieval
		"MagnetURI": p.MagnetURI,
		"Link":      p.Link,
	})
}

//util open in app

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

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
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 100

	return textInputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.ClearScreen)
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			os.Exit(1)
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
	return fmt.Sprintf(
		"What to privateer today?\n\n%s\n\n%s",
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
	return tea.Batch(getResults(), m.spinner.Tick, tea.ClearScreen)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(result.Results) > 0 {
		m.quitting = true
	}
	if m.quitting {
		return m, tea.Batch(tea.Quit, tea.ClearScreen)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			os.Exit(1)
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

// table tui

type TableModel struct {
	pokeTable table.Model

	// currentPokemonData JackettResult
}

func NewTableModel() TableModel {

	rows := []table.Row{}

	for _, p := range result.Results {
		rows = append(rows, p.ToRow())
	}
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	w, h := termbox.Size()
	termbox.Close()
	return TableModel{
		pokeTable: table.New([]table.Column{
			table.NewColumn("Title", "Title", w-70),
			table.NewColumn("Tracker", "Tracker", 10),
			table.NewColumn("Category", "Category", 10),
			table.NewColumn("Date", "Date", 20),
			table.NewColumn("Size", "Size", 10).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
			table.NewColumn("Seeders", "Seeders", 10),
		}).WithRows(rows).
			BorderRounded().WithPageSize(h - 6).SortByDesc("Seeders").
			Focused(true).WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left)),
	}
}

func (m TableModel) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.pokeTable, cmd = m.pokeTable.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		case "enter":
			var selectedData map[string]interface{} = m.pokeTable.HighlightedRow().Data
			dbByte, _ := json.Marshal(selectedData)
			var myStruct JackettResult
			_ = json.Unmarshal(dbByte, &myStruct)
			if myStruct.MagnetURI != "" {
				openbrowser(myStruct.MagnetURI)
			} else {
				openbrowser(myStruct.Link)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m TableModel) View() string {
	return m.pokeTable.View()
}

// api key input

type apiInputModel struct {
	textInput textinput.Model
	err       error
}

func ApiInputModel() apiInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Jackett Api key..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 100

	return apiInputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m apiInputModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.ClearScreen)
}

func (m apiInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			os.Exit(1)
			return m, tea.Quit
		case tea.KeyEnter:
			data := Config{
				ApiKey: m.textInput.Value(),
			}
			file, _ := json.Marshal(data)

			ioutil.WriteFile("privateer.json", file, 0644)
			apiKey = m.textInput.Value()
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

func (m apiInputModel) View() string {
	return fmt.Sprintf(
		"Enter Jackett Api key?\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
