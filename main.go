package main

import (
	"camloc-go/calc"
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


var ClientList = make(map[string]calc.Camera)
var TopicMatchers = make(map[string]regexp.Regexp)

func f32FromBytes(bytes []byte) float32 {
    bits := binary.BigEndian.Uint32(bytes)
    return math.Float32frombits(bits)
}

var plus = regexp.MustCompile(`\+`)

// replaces + sign in topics to regexes for matching client ids
func FillMatchers(topics ...string) {
    for _, v := range topics {
        re := regexp.MustCompile(string(plus.ReplaceAll([]byte(v), []byte("([a-zA-Z\\d]+)"))))
        TopicMatchers[v] = *re
    }
}

// gets the client id from topic wildcard
func getClientId(wildcardTopic string, topic string) *string {
    if v, ok := TopicMatchers[wildcardTopic]; ok {
        if match := v.FindStringSubmatch(topic); len(match) > 1 {
            return &match[1]
        }
    }
    return nil
}

// replaces the plus with the client id
func replaceClientId(topic string, clientId string) string {
    return string(plus.ReplaceAll([]byte(topic), []byte(clientId)))
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {    
    util.Msg(msg.Topic(), msg.Payload())
}

var locateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    if id := getClientId(util.GetLocation, msg.Topic()); id != nil {
        x := f32FromBytes(msg.Payload())
        util.I("got position %f from %s", x, *id)

        if entry, ok := ClientList[*id]; ok {
            entry.LastX = float64(x)
            ClientList[*id] = entry
        }
    }
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
    id := getClientId(util.GetConfig, topic)
    if id == nil || len(payload) < 16 {
        util.W("did not update configuration (%s)", topic)
        return
    }

    // update or create
    x, y, rot, fov := f32FromBytes(payload[0:4]), f32FromBytes(payload[4:8]), f32FromBytes(payload[8:12]), f32FromBytes(payload[12:16])

    if entry, ok := ClientList[*id]; ok {
        entry.Position = calc.Position{ X: float64(x), Y: float64(y), Rotation: float64(rot) }
        entry.Fov = float64(fov)
        ClientList[*id] = entry
    } else {
        ClientList[*id] = calc.Camera{ 
            Position: calc.Position{ X: float64(x), Y: float64(y), Rotation: float64(rot) }, 
            Fov: float64(fov),
            LastX: 0.,
        }

        // TESTING
        go func ()  {
            time.Sleep(time.Duration(time.Second * 2))
            setConfig(client, *id, calc.Camera{ Position: calc.Position{ X: 3.5, Y: 3, Rotation: 69 } })
            
            // time.Sleep(time.Duration(time.Second * 2))
            // setAllState(client, true)

            // time.Sleep(time.Duration(time.Second * 4)) 
            // flash(client, *id)

            // time.Sleep(time.Duration(time.Second * 2)) 
            // setAllState(client, false)
        }()
    }

    util.D("new config for %s : %v", *id, ClientList[*id])

    
}

// flash lights on a client
func flash(client mqtt.Client, clientId string) {
    pub(client, replaceClientId(util.Flash, clientId), []byte{})
}

func askState(client mqtt.Client) {
    pub(client, util.AskForState, []byte{})
}

func setState(client mqtt.Client, clientId string, on bool) {
    if on {
        pub(client, replaceClientId(util.SetState, clientId), []byte{0x1})
    } else {
        pub(client, replaceClientId(util.SetState, clientId), []byte{0x0})
    }
}

// turn on/off every client camera
func setAllState(client mqtt.Client, on bool) {
    if on {
        pub(client, util.SetAllState, []byte{0x1})
    } else {
        pub(client, util.SetAllState, []byte{0x0})
    }
}

func setConfig(client mqtt.Client, clientId string, config calc.Camera) {
    buf := make([]byte, 3 * 4)
    binary.BigEndian.PutUint32(buf[:4], math.Float32bits(float32(config.X)))
    binary.BigEndian.PutUint32(buf[4:8], math.Float32bits(float32(config.Y)))
    binary.BigEndian.PutUint32(buf[8:], math.Float32bits(float32(config.Rotation)))
    pub(client, replaceClientId(util.SetConfig, clientId), buf)
}

// self connected
var onConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    util.I("connected to broker")

    sub(client, util.GetConfig, configHandler)
    sub(client, util.GetLocation, locateHandler)
    sub(client, util.Disconnect, disconnectHandler)
    sub(client, util.GetState, nil)

    // ask for config
    pub(client, util.AskForConfig, []byte{})
}

var onLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    util.E("connection lost: %v", err)
}

func main() {
    // TEST
    // calc.Calc(0, 3, 3, 0, 0.5, 0.5, 90, 90, 0, 0)

    // cam1 := calc.Camera {
    //     Position: calc.Position{ X: 0, Y: 3, Rotation: 0 },
    // }
    // cam2 := calc.Camera {
    //     Position: calc.Position{ X: 3, Y: 0, Rotation: 90 },
    // }

    // fmt.Printf("calc.CheckSetup(cam1, cam2): %v\n", calc.CheckSetup(cam1, cam2))

    // return

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
    opts.SetDefaultPublishHandler(defaultHandler)
    opts.SetWill(replaceClientId(util.Disconnect, opts.ClientID), "goodbye cruel world", 0, false)

    opts.OnConnect = onConnectHandler
    opts.OnConnectionLost = onLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.WaitTimeout(time.Duration(time.Duration.Seconds(5))) && token.Error() != nil {
        panic(token.Error())
    }
    
    FillMatchers(util.Disconnect, util.GetConfig, util.GetLocation, util.SetConfig)

    // cleanup
    go func() {
        s := <-sigs
        util.I("%s", s)
        client.Disconnect(500)
        end <- true
    }()
    <-end 
}

func pub(client mqtt.Client, topic string, message any) {
    token := client.Publish(topic, 0, false, message)
    r := token.Wait()
    util.D("%v published %s: %s", r, topic, message)
}

func sub(client mqtt.Client, topic string, handler mqtt.MessageHandler) {
    token := client.Subscribe(topic, 0, handler)
    token.Wait()
    util.I("subscribed to topic: %s", topic)
}
 
