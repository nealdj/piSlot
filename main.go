package main

import (
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	lane_order = []string{
		"red",
		"white",
		"green",
		"orange",
		"blue",
		"yellow",
		"purple",
		"black"}

	lane_layout = map[string][]int{
		//Lane   Power, Lap Counter
		"red":    {18, 4},
		"white":  {23, 17},
		"green":  {24, 27},
		"orange": {25, 22},
		"blue":   {12, 5},
		"yellow": {16, 6},
		"purple": {20, 13},
		"black":  {21, 19},
	}

	lane_power_mu sync.Mutex
	lane_lap_mu   sync.Mutex
	min_lap_time  float64 = 1.0
)

func quitPiSlot(app *tview.Application, flexScreen *tview.Flex, trackControlButtons *tview.List) {
	form := tview.NewForm()
	form.AddTextView("", "Are you sure you want to quit piSlot?", 70, 2, true, false)
	form.AddButton("Back", func() {
		app.SetRoot(flexScreen, true).SetFocus(trackControlButtons)

	})
	form.AddButton("Quit", func() {
		app.Stop()
		os.Exit(0)
	})
	if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
		panic(err)
	}

}

func main() {

	app := tview.NewApplication()

	lapTableColNames := []string{"Lane", "Laps", "Fast Lap", "Last Lap"}
	lapTable := tview.NewTable().
		SetBorders(true).
		SetBordersColor(tcell.ColorSlateGrey)
	for col, name := range lapTableColNames {
		lapTable.SetCell(0, col,
			tview.NewTableCell(paddedString(name)).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false).
				SetExpansion(10))
	}
	for lane, pins := range lane_layout {
		go lapCounter(lane, pins[1], app, lapTable)
	}
	go updateLapTable(app, lapTable)

	trackControlButtons := tview.NewList()
	trackControls := tview.NewFrame(trackControlButtons).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Track Controls", true, tview.AlignCenter, tcell.ColorWhite)

	flexScreen := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().
				SetBorder(true).
				SetTitle("piSlot"), 0, 1, false).
			AddItem(lapTable, 0, 3, false).
			AddItem(tview.NewBox().
				SetBorder(true).
				SetTitle("Stats"), 5, 1, false), 0, 2, false).
		AddItem(trackControls, 50, 1, false)

	trackControlButtons.AddItem("Set Lane Time", "Turns on specific lanes for X minutes", 'a', func() {
		powerLaneForm(app, flexScreen, trackControlButtons)
	}).
		AddItem("Set Track Time", "Turns on all lanes for X minutes", 'b', nil).
		AddItem("Start Race", "Setup, and then start a race", 'c', nil).
		AddItem("Reset Lap Counts & Time", "Sets lap counts to 0, clears fast/last lap", 'd', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			quitPiSlot(app, flexScreen, trackControlButtons)
		})

	if err := app.SetRoot(flexScreen, true).SetFocus(trackControlButtons).Run(); err != nil {
		panic(err)
	}

}
