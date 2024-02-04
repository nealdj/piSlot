package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	last_lap_unix  = make(map[string]int64)
	lap_count_time = make(map[string][]float64)
)

type lapStats struct {
	laps    string
	fastLap string
	lastLap string
}

// Colors the given string with the lane's color
func laneColor(lane string, text string) string {
	lane = strings.ToLower(lane)
	var colorText string
	switch lane {
	case "black":
		colorText = "[#b4b4b4]" + text + "[#b4b4b4]"
	default:
		colorText = "[" + lane + "]" + text + "[" + lane + "]"
	}

	return colorText
}

// Registers a lap against the lap counter if it's above the min lap time.
// Will trigger an update to the lap table.
func lap(lane string, unix_now int64, app *tview.Application, table *tview.Table) bool {
	var lap_time_s float64
	if last_lap_unix[lane] == 0 {
		//log.Println("First lap on", lane)
		last_lap_unix[lane] = unix_now
		return true
	}

	lane_lap_mu.Lock()
	defer lane_lap_mu.Unlock()

	lap_time_ms := unix_now - last_lap_unix[lane]
	lap_time_s = float64(lap_time_ms) / 1000.00
	if lap_time_s > min_lap_time {
		//log.Println("Counted lap on", lane, lap_time_s)
		last_lap_unix[lane] = unix_now
		lap_count_time[lane] = append(lap_count_time[lane], lap_time_s)
		updateLapTable(app, table)

	} else {
		//log.Println("Fast lap NOT counted on", lane, lap_time_s)
		return false
	}
	//fmt.Println(lap_count_time)
	return true
}

// Sets GPIO pins as input, and then watches for laps.
func lapCounter(lane string, gpio_pin int, app *tview.Application, table *tview.Table) {
	pin := rpio.Pin(gpio_pin)
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge)
	//log.Println("Starting Lap Counter on", lane, "Pin:", gpio_pin)
	for {
		if pin.EdgeDetected() {
			lap(lane, time.Now().UnixMilli(), app, table)
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// Returns the fastest lap from the lap slice.
func getFastLap(numbers []float64) float64 {
	if len(numbers) == 0 {
		return math.NaN()
	}

	minValue := numbers[0]
	for _, num := range numbers {
		if num < minValue {
			minValue = num
		}
	}
	return minValue
}

// Returns the lap data as a map, to be used in the lap table.
func readLapStats() map[string][]string {
	lapStats := make(map[string][]string)
	for _, lane := range lane_order {
		var fastLap string
		var lastLap string
		laps := strconv.Itoa(len(lap_count_time[lane]))
		if laps == "0" {
			fastLap = "."
			lastLap = "."
		} else {
			fastLap = strconv.FormatFloat(getFastLap(lap_count_time[lane]), 'f', -1, 64)
			lastLap = strconv.FormatFloat(lap_count_time[lane][len(lap_count_time[lane])-1], 'f', -1, 64)
		}
		lapStats[lane] = []string{laps, fastLap, lastLap}
	}
	return lapStats
}

// Pads string to make the tview.Table more readable when using borders.
func paddedString(text string) string {
	padding := "  "
	return padding + text + padding
}

// Updates the lap table when called after a lap is registered.
func updateLapTable(app *tview.Application, table *tview.Table) {
	var laneRow int = 1
	lapStats := readLapStats()
	for _, lane := range lane_order {
		laneText := paddedString(laneColor(lane, lane))
		table.SetCell(laneRow, 0,
			tview.NewTableCell(laneText).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold).
				SetSelectable(false).
				SetExpansion(0))
		for col, stat := range lapStats[lane] {
			colorStat := paddedString(laneColor(lane, stat))
			table.SetCell(laneRow, col+1,
				tview.NewTableCell(colorStat).
					SetAlign(tview.AlignCenter).
					SetAttributes(tcell.AttrBold).
					SetSelectable(false).
					SetExpansion(0))
			col++
		}
		laneRow++
	}
	app.QueueUpdateDraw(func() {})
}
