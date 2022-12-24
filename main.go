package main

import (
	"github.com/jm199seo/dhg_bot/cmd"
	"github.com/jm199seo/dhg_bot/util/logger"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.Log.Panic(err)
	}
}
