package main

import (
	"os"
	"github.com/fabledruns/blockchain/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()
}