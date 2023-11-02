package util

import (
	"fmt"
	"time"
)

const reset = "\x1b[0m"
const bright = "\x1b[1m"
const dim = "\x1b[2m"
const blue = "\x1b[34m"
const red = "\x1b[31m"
const yellow = "\x1b[33m"
const green = "\x1b[32m";
const cyan = "\x1b[36m";
const italic = "\x1b[3m";

func now() string {
	// time fmt table: https://stackoverflow.com/a/69338568
	return fmt.Sprintf("%s%s%s", dim, time.Now().Format("15:04:05"), reset)
}

func D(format string, a ...any) {
	fmt.Printf("%s %s[DEBUG]%s %s\n", now(), green + bright + italic, reset, fmt.Sprintf(format, a...))
}

func E(format string, a ...any) {
	fmt.Printf("%s %s[ERROR]%s %s\n", now(), red + bright + italic, reset, fmt.Sprintf(format, a...))
}

func I(format string, a ...any) {
	fmt.Printf("%s %s[INFO]%s %s\n", now(), blue + bright + italic, reset, fmt.Sprintf(format, a...))
}

func W(format string, a ...any) {
	fmt.Printf("%s %s[WARN]%s %s\n", now(), yellow + bright + italic, reset, fmt.Sprintf(format, a...))
}

func Msg(topic string, content []byte) {
	fmt.Printf("%s %s[MSG] %s%s %s\n", now(), cyan + bright + italic, reset + bright + cyan + dim, topic + reset, content)
}
