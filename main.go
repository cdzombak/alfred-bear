package main

import (
	"fmt"
	"os"

	aw "github.com/deanishe/awgo"
	humanize "github.com/dustin/go-humanize"
)

var (
	IconWorkflow = &aw.Icon{
		Value: "icon.png",
		Type:  aw.IconTypeImage,
	}
	wf        *aw.Workflow
	bearToken string
	xcallPath string
)

func init() {
	wf = aw.New()

	bearToken = os.Getenv("BEAR_TOKEN")
	xcallPath = wf.Dir() + "/xcall.app/Contents/MacOS/xcall"
	if bearToken == "" {
		wf.NewWarningItem("BEAR_TOKEN missing in workflow settings", "In Bear, go to Help > API Token to find your token")
		wf.SendFeedback()
	}
}

func run() {
	query := os.Args[1]
	notes, err := search(query)
	if err != nil {
		wf.Fatalf("%s", err)
	}
	for _, note := range notes {
		wf.NewItem(note.Title).
			Subtitle(fmt.Sprintf("Last edited %s", humanize.Time(note.ModificationDate))).
			UID(note.Identifier).
			Arg(note.Identifier).
			Valid(true).
			Icon(IconWorkflow)
	}
	wf.WarnEmpty("No matching notes found", "Try another query")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
