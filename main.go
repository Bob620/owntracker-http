package main

import (
	"fmt"
	"io"
	"net/http"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	Data     []byte
	Username string
	DeviceId string
}

func main() {
	messagePasser := make(chan Message)

	tdr := MQTT.ClientOptions{
		AutoReconnect: true,
		Username:      "",
		Password:      "",
		ClientID:      "HttpBridge",
		KeepAlive:     3600,
	}
	tdr.AddBroker(":8883")

	tdr.OnConnect = func(client MQTT.Client) {
		go func() {
			for client.IsConnected() {
				message := <-messagePasser
				messageString := fmt.Sprintf("%s", message.Data)

				_ = client.Publish(fmt.Sprintf("%s/%s", message.Username, message.DeviceId), 1, true, messageString)
			}
		}()
	}

	tdr.OnReconnecting = func(client MQTT.Client, options *MQTT.ClientOptions) {
		println("Retrying connecting to tdr.moe MQTT")
	}

	tdr.OnConnectionLost = func(client MQTT.Client, err error) {
		println("Lost connection to tdr.moe MQTT")
	}

	tdrClient := MQTT.NewClient(&tdr)
	if token := tdrClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		println("Connected to tdr.moe MQTT")
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.ContentLength > 0 {
			data := make([]byte, request.ContentLength)
			if _, err := io.ReadFull(request.Body, data); err != nil {
				fmt.Println("error:", err)
			}

			messagePasser <- Message{data, request.Header.Get("X-Limit-U"), request.Header.Get("X-Limit-D")}
		}
	})

	println("Listening on port 8884")
	_ = http.ListenAndServe(":8884", nil)
}
