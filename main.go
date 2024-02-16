package main

import (
	"C"
)
import (
	"os"

	"github.com/joho/godotenv"
	"github.com/oklookat/synchro/commander"
)

func main() {
	// for testing
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	commander.SetOnLogger(&onLogger{})
	if err := commander.Boot(os.Getenv("DATA_DIR")); err != nil {
		panic(err)
	}
}

type onLogger struct {
}

func (e onLogger) OnLog(level int, msg string) {
	println(msg)
}

type onUrler struct {
}

func (e onUrler) OnURL(url string) {
	println(url)
}
