package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tarm/serial"
)

func main() {
	// MQTT broker configuration
	mqttBroker := "tcp://localhost:1883"
	topic := "topic/temperature"

	// MQTT connection setup
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID("go_mqtt_publisher")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar ao servidor MQTT: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Arduino Serial Port
	serialPort := "/dev/ttyACM0" // maybe /dev/ttyACM0 or /dev/ttyUSB0
	if len(os.Args) > 1 {
		serialPort = os.Args[1]
	}

	// Serial connection setup
	config := &serial.Config{Name: serialPort, Baud: 9600}
	s, err := serial.OpenPort(config)
	if err != nil {
		log.Fatalf("Serial port fail %s: %v", serialPort, err)
	}
	defer s.Close()

	reader := bufio.NewReader(s)

	for {
		// Read a line from serial port
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Serial port read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Remove whitespaces and newline characters
		line = strings.TrimSpace(line)

		if line == "" {
			// Empty line? Log it and continue
			log.Println("Empty line received from serial port")
			continue
		}

		// Parse ADC value
		var adcValue uint32
		_, err = fmt.Sscanf(line, "%d", &adcValue)
		if err != nil {
			log.Printf("Error parsing ADC value '%s': %v", line, err)
			continue
		}

		// Calculate temperature
		tempCelsius := computeTemperature(adcValue)
		tempFahrenheit := tempCelsius*9/5 + 32

		// Publish temperature to MQTT topic
		text := fmt.Sprintf("ADC: %d, Temp: %.2f°C / %.2f°F", adcValue, tempCelsius, tempFahrenheit)
		token := client.Publish(topic, 0, false, text)
		token.Wait()

		fmt.Printf("Published to MQTT: %s\n", text)
	}
}

// Celsius temperature converter
func computeTemperature(adcValue uint32) float64 {
	// Constants for the Steinhart-Hart equation
	const (
		VCC    = 5.0           // Tension (V)
		R      = 1000.0        // Fixed resistor (Ohms)
		RT0    = 1000.0        // Thermistor resistir T0 (Ohms)
		T0     = 25.0 + 273.15 // Reference temperature (Kelvin)
		B      = 3977.0        // Beta coefficient (K)
		ADCMax = 65535.0       // Max ADC value (16 bits TinyGo)
	)

	// Tension on thermistor
	VRT := VCC * float64(adcValue) / ADCMax

	// Avoid division by zero
	if VRT == 0 {
		return math.NaN() // Return "Not a Number" if division by zero
	}

	// Calculate thermistor resistance
	RT := VRT * R / (VCC - VRT)

	// Calculate Kelvin temperature using Steinhart-Hart equation
	ln := math.Log(RT / RT0)
	TX := 1 / ((ln / B) + (1 / T0))

	// Convert to Celsius
	tempCelsius := TX - 273.15

	return tempCelsius
}
