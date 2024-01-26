package main

import (
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

func main() {

	app := tview.NewApplication()

	powerAll(true, 1)

	table_col_names := []string{"Lane", "Laps", "Fast Lap", "Last Lap"}
	table := tview.NewTable().
		SetBorders(true)
	for col, name := range table_col_names {
		table.SetCell(0, col,
			tview.NewTableCell(paddedString(name)).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false))
	}
	for lane, pins := range lane_layout {
		go lapCounter(lane, pins[1], app, table)
	}
	go updateLapTable(app, table)

	app.SetRoot(table, true).SetFocus(table).Run()

}
