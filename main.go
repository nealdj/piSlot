package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stianeikeland/go-rpio/v4"
)

var (
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

	min_lap_time float64 = 2.1

	view *tview.Box
	app  *tview.Application
)

const refreshInterval = 500 * time.Millisecond

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
	fmt.Println(lap_count_time)
	return true
}

func lapcounter(lane string, gpio_pin int) {
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
			lap(lane, time.Now().UnixMilli())
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
		log.Println("Powered ON", lane, "lane, pin", gpio_pin)
		if minutes > 0 {
			time.Sleep(time.Duration(minutes) * time.Minute)
			log.Println("Power OFF, timeout", lane, "lane, pin", gpio_pin)
			pin.Low()

		}
	} else {
		pin.Low()
		//log.Println("Powered OFF", lane, "lane, pin", gpio_pin)
	}
}

func lapcounts(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	var lane_stats string
	for lane := range lane_lap_layout {
		lap_count := len(lap_count_time[lane])
		lane_stat := lane_stats + lane + strconv.Itoa(lap_count) + "\n"
		tview.Print(screen, lane_stat, x, height/2, width, tview.AlignCenter, tcell.ColorLime)
	}

	return 0, 0, 0, 0
}

func refresh() {
	tick := time.NewTicker(refreshInterval)
	for {
		select {
		case <-tick.C:
			app.Draw()
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	//power_all(false, 0)
	power_all(true, 1)

	for lane, gpio_pin := range lane_lap_layout {
		go lapcounter(lane, gpio_pin)
	}

	//app = tview.NewApplication()
	//view = tview.NewBox().SetBorder(true).SetTitle("Hello, world!").SetDrawFunc(lapcounts)

	//go refresh()
	//if err := app.SetRoot(view, true).Run(); err != nil {
	//	panic(err)
	//}

	var wg = &sync.WaitGroup{}
	for lane, gpio_pin := range lane_lap_layout {
		go lapcounter(lane, gpio_pin)
		wg.Add(1)
	}
	wg.Wait()
}
