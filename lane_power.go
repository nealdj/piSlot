package main

import (
	"fmt"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// Starts or stops power on all lanes for x number of minutes given.
func powerAll(power bool, minutes int) {
	for lane, pins := range lane_layout {
		go powerLane(lane, pins[0], power, minutes)
	}
}

// Starts power on lane for x number of minutes or stops power immediately.
func powerLane(lane string, gpio_pin int, power bool, minutes int) {
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
