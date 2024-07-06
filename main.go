package main

import (
	"log/slog"
	"os"

	"github.com/oklookat/synchro/commander"
)

func main() {
	if err := commander.Boot(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
