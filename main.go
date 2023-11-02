package main

import (
	"camloc-go/util"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ClientConfig struct {
    x       float32
    y       float32
    rot     float32
    lastX   float32
}

var ClientList = make(map[string]ClientConfig)
var TopicMatchers = make(map[string]regexp.Regexp)

func f32FromBytes(bytes []byte) float32 {
    bits := binary.BigEndian.Uint32(bytes)
    return math.Float32frombits(bits)
}

// replaces + sign in topics to regexes for matching client ids
func FillMatchers(topics ...string) {
    var plus = regexp.MustCompile(`\+`)
    for _, v := range topics {
        re := regexp.MustCompile(string(plus.ReplaceAll([]byte(v), []byte("([a-zA-Z\\d]+)"))))
        TopicMatchers[v] = *re
    }
}

// gets the client id from topic wildcard
func getClientId(wildcardTopic string, topic string) *string {
    if v, ok := TopicMatchers[wildcardTopic]; ok {
        if match := v.FindStringSubmatch(topic); len(match) > 1 {
            util.D("%v", match)
            return &match[1]
        }
    }
    return nil
}

var defaultPubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {    
    util.Msg(msg.Topic(), msg.Payload())
}

var locateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {    
    util.I("got position %f", f32FromBytes(msg.Payload()))
}

// remove client from map when last will is published
var disconnectHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {    
    if id := getClientId(util.Disconnect, msg.Topic()); id != nil {
        delete(ClientList, *id)
        util.I("%s disconnected", *id)
    }
}

// update configuration
var configHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {    
    topic, payload := msg.Topic(), msg.Payload()
    id := getClientId(util.Disconnect, topic)
    if id == nil {
        return
    }

    // update or create
    x, y, rot := f32FromBytes(payload[0:4]), f32FromBytes(payload[4:8]), f32FromBytes(payload[8:12])
    if v, exists := ClientList[*id]; exists {
        ClientList[*id] = ClientConfig{ x: x, y: y, rot: rot, lastX: v.lastX }
    } else {
        ClientList[*id] = ClientConfig{ x: x, y: y, rot: rot, lastX: -1 }
    }

    util.D("new config for %s : %v", *id, ClientList[*id])
}

// self connected
var onConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    util.I("connected to broker")
    // ask for config
    pub(client, util.AskForConfig, []byte{})
}

var onLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    util.E("connection lost: %v", err)
}

func main() {
    // args
    broker := flag.String("broker", "127.0.0.1", "the broker ip address")
    port := flag.Int("port", 1883, "the broker port")
    flag.Parse()

    // ctrlc handler 
    sigs := make(chan os.Signal, 1)
    end := make(chan bool, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", *broker, *port))
    opts.SetClientID("go_mqtt_client")
    opts.SetDefaultPublishHandler(defaultPubHandler)

    opts.OnConnect = onConnectHandler
    opts.OnConnectionLost = onLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.WaitTimeout(time.Duration(time.Duration.Seconds(5))) && token.Error() != nil {
        panic(token.Error())
    }
    
    go func() {
        FillMatchers(util.Disconnect, util.GetConfig, util.GetLocation, util.SetConfig)
    
        sub(client, util.GetConfig, configHandler)
        sub(client, util.GetLocation, locateHandler)
        sub(client, util.Disconnect, disconnectHandler)
    }()

    
    // cleanup
    go func() {
        <-sigs
        util.I("shutdown")
        client.Disconnect(200)
        end <- true
    }()
    <-end 
}

func pub(client mqtt.Client, topic string, message any) {
    token := client.Publish(topic, 0, false, message)
    token.Wait()
    util.D("published %s: %s", topic, message)
}

func sub(client mqtt.Client, topic string, handler mqtt.MessageHandler) {
    token := client.Subscribe(topic, 0, handler)
    token.Wait()
    util.I("subscribed to topic: %s", topic)
}
 
