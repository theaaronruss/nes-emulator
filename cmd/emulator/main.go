package main

import (
	"fmt"
	"os"

	"github.com/theaaronruss/nes-emulator/internal/nes"
)

func main() {
	cartridge, err := nes.NewCartridge("nestest.nes")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	bus := nes.NewSysBus()
	bus.InsertCartridge(cartridge)
	cpu := nes.NewCpu(bus)
	for {
		cpu.Clock()
	}
}
