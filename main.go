package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	g "github.com/AllenDang/giu"
)

var aircraftChecklists = make(map[string][]ChecklistCategory)
var currentChecklistIndex int32 = 0
var WIDTH float32 = 200 // Width for the combo box
var (
	errorMessage   string
	showErrorPopup bool
)

type ChecklistItem struct {
	Text    string
	Checked bool
}

type ChecklistCategory struct {
	Title string
	Items []ChecklistItem
}

var (
	aircrafts = []string{
		"Cessna 172", // Default aircraft (0 index)
		"Piper PA-28",
		"Beechcraft Bonanza",
		"Cirrus SR22",
	}
	currentAircraft int32 = 0 // Default to Cessna 172 (index 0)
)

func init() {
	// Load the checklist for all aircraft
	for _, aircraft := range aircrafts {
		if err := loadChecklist(aircraft); err != nil {
			errorMessage = fmt.Sprintf("Failed to load checklists for %s: %v", aircraft, err)
			showErrorPopup = true
			return
		}
	}
}

func DrawUI() {
	currentChecklist := getAircaftChecklist()
	autoCycleCompleteChecklist()

	// Open error popup if needed
	if showErrorPopup {
		g.OpenPopup("Error")
	}

	g.SingleWindow().Layout(
		g.Label("Aircraft Checklist Program"),
		g.Combo("Select Aircraft", aircrafts[currentAircraft], aircrafts, &currentAircraft).Size(150),
		g.Separator(),
		showError(),

		// Only show checklist UI if we have checklists loaded
		g.Condition(len(currentChecklist) > 0,
			g.Layout{
				g.Combo("Select Checklist", currentChecklist[currentChecklistIndex].Title, getChecklistTitles(), &currentChecklistIndex).Size(150),
				g.Separator(),

				g.Label(fmt.Sprintf("- %s -", currentChecklist[currentChecklistIndex].Title)),
				// Put checklist in its own container with a border
				g.Layout{
					g.Row(
						g.Button("<- Previous Checklist").OnClick(prevChecklist),
						g.Button("Next Checklist ->").OnClick(nextChecklist),
					),
				},

				g.Child().Size(500, 500).Border(true).Layout(
					buildChecklistLayout(),
				),
				g.Label(getProgressText()),
				g.Button("Reset Checklist").OnClick(resetCurrentChecklist),
			},
			g.Label("No checklist available for this aircraft"),
		),
	)
}

func getAircaftChecklist() []ChecklistCategory {
	aircraft := aircrafts[currentAircraft]
	if checklist, exists := aircraftChecklists[aircraft]; exists {
		return checklist
	}
	return nil
}

func loadChecklist(aircraft string) error {
	filename := fmt.Sprintf("Resources/%s-Checklist.json", aircraft)
	filepath := filepath.Clean(filename)

	//read data from file
	data, err := os.ReadFile(filepath)
	if err != nil {
		errorMessage = fmt.Sprintf("Failed to read checklist file for %s: %v", aircraft, err)
		showErrorPopup = true
		return errors.New(errorMessage)
	}

	//parse JSON data
	var categories []ChecklistCategory
	if err := json.Unmarshal(data, &categories); err != nil {
		errorMessage = fmt.Sprintf("Failed to parse checklist file for %s: %v", aircraft, err)
		showErrorPopup = true
		return errors.New(errorMessage)
	}

	//store in map
	aircraftChecklists[aircraft] = categories
	return nil
}

func getChecklistTitles() []string {
	currentChecklist := getAircaftChecklist()
	titles := make([]string, len(currentChecklist))
	for i, cl := range currentChecklist {
		titles[i] = cl.Title
	}
	return titles
}

func nextChecklist() {
	currentChecklist := getAircaftChecklist()

	currentChecklistIndex++

	if currentChecklistIndex >= int32(len(currentChecklist)) {
		currentChecklistIndex = 0 // Reset to the first checklist if all are completed
	}
}

func prevChecklist() {
	currentChecklist := getAircaftChecklist()

	currentChecklistIndex--

	if currentChecklistIndex >= int32(len(currentChecklist)) {
		currentChecklistIndex = int32(len(currentChecklist)) - 1 // Reset to the last checklist if all are completed
	}
}

func buildChecklistLayout() g.Layout {
	var layout g.Layout
	currentAircraft := getAircaftChecklist()
	if len(currentAircraft) == 0 {
		return layout
	}

	for i := range currentAircraft[currentChecklistIndex].Items {
		item := &currentAircraft[currentChecklistIndex].Items[i]
		layout = append(layout, g.Checkbox(item.Text, &item.Checked))
	}
	return layout
}

func getProgressText() string {
	currentChecklist := getAircaftChecklist()
	if len(currentChecklist) == 0 {
		return "No checklist available."
	}
	items := currentChecklist[currentChecklistIndex].Items
	complete := 0
	for _, items := range items {
		if items.Checked {
			complete++
		}
	}
	return fmt.Sprintf("Progress: %d/%d", complete, len(items))
}

func autoCycleCompleteChecklist() {
	currentChecklist := getAircaftChecklist()
	if len(currentChecklist) == 0 {
		return
	}

	items := currentChecklist[currentChecklistIndex].Items
	complete := 0
	for _, item := range items {
		if item.Checked {
			complete++
		}
	}

	if complete == len(items) {
		currentChecklistIndex++
		if currentChecklistIndex >= int32(len(currentChecklist)) {
			currentChecklistIndex = 0 // Reset to the first checklist if all are completed
		}
	}
}

func resetCurrentChecklist() {
	currentChecklist := getAircaftChecklist()
	if len(currentChecklist) == 0 {
		return
	}

	for i := range currentChecklist[currentChecklistIndex].Items {
		currentChecklist[currentChecklistIndex].Items[i].Checked = false
	}
}

func showError() g.Layout {
	if !showErrorPopup {
		return nil
	}

	return g.Layout{
		g.PopupModal("Error").Layout(
			g.Label(errorMessage),
			g.Separator(),
			g.Button("OK").OnClick(func() {
				showErrorPopup = false
			}),
		).IsOpen(&showErrorPopup),
	}
}

func main() {
	g.NewMasterWindow("Aircraft Checklists", 800, 600, g.MasterWindowFlagsMaximized).Run(DrawUI)
}
