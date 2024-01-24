package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rivo/tview"
	"github.com/stianeikeland/go-rpio/v4"
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
	lane_lap_layout = map[string]int{
		"red":    4,
		"white":  17,
		"green":  27,
		"orange": 22,
		"blue":   5,
		"yellow": 6,
		"purple": 13,
		"black":  19,
	}

	lane_power_layout = map[string]int{
		"red":    18,
		"white":  23,
		"green":  24,
		"orange": 25,
		"blue":   12,
		"yellow": 16,
		"purple": 20,
		"black":  21,
	}

	lap_count_unix = make(map[string][]int64)
	lap_count_time = make(map[string][]float64)
	lane_power_mu  sync.Mutex
	mu             sync.Mutex

	min_lap_time float64 = 0.4

	view *tview.Box
	app  *tview.Application
)

func power_all(power bool, minutes int) {
	for lane, gpio_pin := range lane_power_layout {
		if power {
			go power_lane(lane, gpio_pin, power, minutes)
		} else {
			power_lane(lane, gpio_pin, power, minutes)
		}

	}
}

func lap(lane string, unix_now int64) bool {
	//s := a[len(a)-1]
	var last_lap_unix int64
	var lap_time_s float64
	if len(lap_count_unix[lane]) == 0 {
		//log.Println("First lap on", lane)
		lap_count_unix[lane] = append(lap_count_unix[lane], unix_now)
		return true
	}

	last_lap_unix = lap_count_unix[lane][len(lap_count_unix[lane])-1]
	mu.Lock()
	defer mu.Unlock()

	lap_time_ms := unix_now - last_lap_unix
	lap_time_s = float64(lap_time_ms) / 1000.00
	if lap_time_s > min_lap_time {
		//log.Println("Counted lap on", lane, lap_time_s)
		// TODO: prevent fatal error: concurrent map read and map write
		lap_count_unix[lane] = append(lap_count_unix[lane], unix_now)
		lap_count_time[lane] = append(lap_count_time[lane], lap_time_s)

	} else {
		//log.Println("Fast lap NOT counted on", lane, lap_time_s)
		return false
	}
	//fmt.Println(lap_count_time)
	return true
}

func lapcounter(lane string, gpio_pin int, textView *tview.TextView) {
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
			if lap(lane, time.Now().UnixMilli()) {
				textView.SetText(read_lap_stats())
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func power_lane(lane string, gpio_pin int, power bool, minutes int) {
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

func get_fast_lap(numbers []float64) float64 {
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

func read_lap_stats() string {
	var stats bytes.Buffer
	var fast_lap string
	var lastLapSeconds string
	for _, lane := range lane_order {
		laps := strconv.Itoa(len(lap_count_time[lane]))
		if laps == "0" {
			fast_lap = ""
			lastLapSeconds = ""
		} else {
			fast_lap = strconv.FormatFloat(get_fast_lap(lap_count_time[lane]), 'f', -1, 64)
			lastLapSeconds = strconv.FormatFloat(lap_count_time[lane][len(lap_count_time[lane])-1], 'f', -1, 64)
		}
		stats.WriteString(lane + " laps:" + laps + " last lap:" + lastLapSeconds + " Fast Lap:" + fast_lap + "\n")
	}
	return stats.String()
}

func main() {
	power_all(true, 1)
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	for lane, gpio_pin := range lane_lap_layout {
		go lapcounter(lane, gpio_pin, textView)
	}
	fmt.Fprintf(textView, read_lap_stats())
	textView.SetBorder(true)
	if err := app.SetRoot(textView, true).SetFocus(textView).Run(); err != nil {
		panic(err)
	}
}
