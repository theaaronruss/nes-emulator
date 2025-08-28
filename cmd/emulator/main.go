package main

import (
	"fmt"

	"github.com/theaaronruss/nes-emulator/internal/bus"
	"github.com/theaaronruss/nes-emulator/internal/cpu"
)

func main() {
	mainBus := bus.NewBus()
	err := mainBus.LoadRom("test.nes")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cpu := cpu.NewCpu(mainBus)
	cpu.ClockCycle()
}
