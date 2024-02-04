package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
	laneTime map[string]int
)

// Starts or stops power on all lanes for x number of minutes given.
func powerAll(power bool, minutes int) {
	for _, lane := range lane_order {
		go powerLane(lane, power, minutes)
	}
}

// Starts power on lane for x number of minutes or stops power immediately.
func powerLane(lane string, power bool, minutes int) {
	gpio_pin := lane_layout[lane][0]
	pin := rpio.Pin(gpio_pin)
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pin.Output()

	if power {
		pin.High()
		//log.Println("Powered ON", lane, "lane, pin", gpio_pin)
		if minutes > 0 {
			time.Sleep(time.Duration(minutes) * time.Minute)
			//log.Println("Power OFF, timeout", lane, "lane, pin", gpio_pin)
			pin.Low()

		}
	} else {
		pin.Low()
		//log.Println("Powered OFF", lane, "lane, pin", gpio_pin)
	}
}

// Focus on dynamically generated tview Form, asking user for minutes and checkbox to power lanes
func powerLaneForm(app *tview.Application, flexScreen *tview.Flex, trackControlButtons *tview.List) {
	laneCheck := make(map[string]bool)
	laneTime := make(map[string]int)
	minutes := 0
	form := tview.NewForm()
	form.AddTextView("", "Set the number of minutes, use TAB to move down the lanes, and space to select a lane.", 70, 2, true, false)
	form.AddInputField("Minutes:", "", 6, tview.InputFieldInteger, func(text string) {
		if i, err := strconv.Atoi(text); err == nil {
			form.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			minutes = i
		} else {
			form.SetFieldBackgroundColor(tcell.ColorRed)
		}
	})
	form.AddButton("Start", func() {
		for lane, power := range laneCheck {
			if power {
				laneTime[lane] = minutes
				go powerLane(lane, power, minutes)
			}
		}
		app.SetRoot(flexScreen, true).SetFocus(trackControlButtons)

	})
	form.AddButton("Cancel", func() {
		app.SetRoot(flexScreen, true).SetFocus(trackControlButtons)

	})

	for _, lane := range lane_order {
		lane := lane
		form.AddCheckbox(laneColor(lane, lane), laneCheck[lane], func(checked bool) {
			laneCheck[lane] = checked
		})
	}

	form.SetBorder(true).SetTitle("Start Lane Time").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
		panic(err)
	}
}
