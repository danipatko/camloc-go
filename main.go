package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const TOPIC_LOCATE = "camloc/locate"

func f32FromBytes(bytes []byte) float32 {
    bits := binary.BigEndian.Uint32(bytes)
    return math.Float32frombits(bits)
}

var onPublishHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    switch msg.Topic() {
        case TOPIC_LOCATE:
            fmt.Printf("got position %f\n", f32FromBytes(msg.Payload()))
        default:
            fmt.Printf("Received message from %s: %s\n", msg.Topic(), msg.Payload())
    }
}

var onConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    fmt.Println("Connected")
}

var onLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    fmt.Printf("Connect lost: %v", err)
}

func main() {
    sigs := make(chan os.Signal, 1)
    end := make(chan bool, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    
    // TODO! cli args
    var broker = "127.0.0.1"
    var port = 1883

    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetClientID("go_mqtt_client")
    opts.SetDefaultPublishHandler(onPublishHandler)

    opts.OnConnect = onConnectHandler
    opts.OnConnectionLost = onLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    sub(client, "topic/test")
    sub(client, "camloc/locate")
    sub(client, "camloc/loc")
    publish(client)

    go func() {
        <-sigs
        fmt.Println("received shutdown request")
        client.Disconnect(200)
        end <- true
    }()
    <-end 
}

// test
func publish(client mqtt.Client) {
    num := 10
    for i := 0; i < num; i++ {
        text := fmt.Sprintf("Message %d", i)
        token := client.Publish("topic/test", 0, false, text)
        token.Wait()
        time.Sleep(time.Second)
    }
}

func sub(client mqtt.Client, topic string) {
    token := client.Subscribe(topic, 1, nil)
    token.Wait()
    fmt.Printf("Subscribed to topic: %s", topic)
}


 
