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
		g.Label("ChecklistPro v1.0"),
		g.Row(
			g.Combo("Select Aircraft", aircrafts[currentAircraft], aircrafts, &currentAircraft).Size(150),
			g.Combo("Select Checklist", currentChecklist[currentChecklistIndex].Title, getChecklistTitles(), &currentChecklistIndex).Size(150),
		),
		g.Row(
			g.Dummy(0, 0),
			g.Separator(),
			g.Dummy(0, 0),
		),
		showError(),

		// Checklist UI (always shown)
		g.Layout{
			g.Column(
				g.Dummy(0, 10),
				g.Style().
					SetFontSize(24).To(
					g.Label(fmt.Sprintf("%s ", currentChecklist[currentChecklistIndex].Title)),
				),
			),

			g.Row(
				g.Child().Size(500, 500).Border(true).Layout(
					buildChecklistLayout(),
				),
			),

			g.Layout{
				g.Child().Size(70, 38).Border(true).Layout(
					g.Row(
						g.Row(
							prevChecklist(),
							g.Button("->").OnClick(nextChecklist),
						),
					),
				),

				g.Label(getProgressText()),
				g.Button("Reset Checklist").OnClick(resetCurrentChecklist),
			},
		},
	)
}
func getAircaftChecklist() []ChecklistCategory {
	aircraft := aircrafts[currentAircraft]
	if checklist, exists := aircraftChecklists[aircraft]; exists {
		return checklist
	}
	// Show error popup if checklist is not found
	errorMessage = fmt.Sprintf("No checklist found for %s", aircraft)
	showErrorPopup = true
	return nil
}

func loadChecklist(aircraft string) error {
	filename := fmt.Sprintf("Resources/checklists/%s-Checklist.json", aircraft)
	cleanedPath := filepath.Clean(filename)

	//read data from file
	data, err := os.ReadFile(cleanedPath)
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

func prevChecklist() g.Layout {
	return g.Layout{
		g.Row(
			g.Condition(currentChecklistIndex == 0,
				g.Layout{
					g.Button("<-").Disabled(true),
				},
				g.Layout{
					g.Button("<-").OnClick(func() {
						if currentChecklistIndex > 0 {
							currentChecklistIndex--
						}
					}),
				},
			),
		),
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
	for _, item := range items {
		if item.Checked {
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
				showErrorPopup = true
			}),
		).IsOpen(&showErrorPopup),
	}
}

func main() {
	g.NewMasterWindow("Aircraft Checklists", 800, 600, g.MasterWindowFlagsMaximized).Run(DrawUI)
}
