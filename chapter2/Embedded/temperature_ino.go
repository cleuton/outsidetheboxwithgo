// tinygo flash -target=arduino -port=/dev/ttyACM0 temperature_ino.go
// tinygo monitor -baudrate=9600
package main

import (
	"machine"
	"time"
)

const (
	adcReadings = 10  // Number of readings to average
	delayMs     = 500 // Delay between readings in milliseconds
)

func main() {

	// Initialize ADC for thermistor
	machine.InitADC()
	thermistor := machine.ADC{Pin: machine.ADC0}
	thermistor.Configure(machine.ADCConfig{})
	var sumADC uint32 = 0
	var count uint8 = 0

	for {
		// Read ADC value from thermistor
		adcValue := thermistor.Get()
		sumADC += uint32(adcValue)
		count++

		if count == adcReadings {
			// Compute average ADC value
			avgADC := sumADC / uint32(count)

			// Send average ADC value over UART
			println(avgADC)

			// Reset sum and count for the next cycle
			sumADC = 0
			count = 0
		}

		// Wait before nexting
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}
