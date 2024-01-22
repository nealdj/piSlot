package main

import (
	"fmt"
	"os"
	"sync"
	"time"

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

	lap_count_unix map[string][]int64
	lap_count_time map[string][]float64

	min_lap_time float64 = 2.1
)

func lap(lane string, unix_now int64) {
	//s := a[len(a)-1]
	lane_laps_unix := lap_count_unix[lane]
	if len(lane_laps_unix) > 0 {
		last_lap_unix := lane_laps_unix[:len(lane_laps_unix)-1]
		fmt.Println("\n", lane, "last lap:", last_lap_unix)
	} else {
		fmt.Println("\n", lane, "First lap:", unix_now)
	}

	lane_laps_unix = append(lane_laps_unix, unix_now)
	fmt.Println(lane, "Laps:", lane_laps_unix)
	fmt.Println(lap_count_unix)

	//if lap_time > min_lap_time {
	//	lap_count[lane] = append(lap_count[lane], unix_now)
	//	fmt.Println(lane, "lap:", unix_now)
	//	fmt.Println(lap_count)
	//} else {
	//	fmt.Println(lane, "lap (not counted):", unix_now)
	//}
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
	fmt.Println("Starting Lap Counter on", lane, "Pin:", gpio_pin)
	for {
		if pin.EdgeDetected() {
			lap(lane, time.Now().UnixMilli())

		}
		time.Sleep(2 * time.Millisecond)
	}
}

func power_lane(lane string, gpio_pin int, power bool) {
	pin := rpio.Pin(gpio_pin)
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	pin.Output()

	if power {
		pin.High()
		fmt.Println(lane, "Powered on, pin:", gpio_pin)
	} else {
		pin.Low()
		fmt.Println(lane, "Powered off, pin:", gpio_pin)
	}
}

func main() {
	for lane, gpio_pin := range lane_power_layout {
		power_lane(lane, gpio_pin, true)
		time.Sleep(200 * time.Millisecond)
		power_lane(lane, gpio_pin, false)

	}
	var wg = &sync.WaitGroup{}
	for lane, gpio_pin := range lane_lap_layout {
		go lapcounter(lane, gpio_pin)
		wg.Add(1)
	}
	wg.Wait()
}
