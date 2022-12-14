package main

import (
	_ "github.com/jkandasa/autoeasy/cmd/init"

	rootCmd "github.com/jkandasa/autoeasy/cmd/root"
)

func main() {
	rootCmd.Execute()
}
