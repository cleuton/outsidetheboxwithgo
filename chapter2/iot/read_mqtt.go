package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// MQTT broker configuration
	mqttBroker := "tcp://localhost:1883"
	topic := "topic/temperature"

	// MQTT connection setup
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID("go_mqtt_subscriber")

	// Callback to process received messages
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Message received %s: %s\n", msg.Topic(), string(msg.Payload()))
	}

	// Callback setup
	opts.SetDefaultPublishHandler(messageHandler)

	// MQTT broker connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Topic subscription
	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to the topic: %v", token.Error())
	}
	fmt.Printf("Subscribed to the topic: %s\n", topic)

	// Keep program running
	select {}
}
