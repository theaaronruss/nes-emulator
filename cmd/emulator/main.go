package main

import (
	"github.com/theaaronruss/nes-emulator/internal/bus"
	"github.com/theaaronruss/nes-emulator/internal/cpu"
)

func main() {
	mainBus := bus.NewBus()
	cpu := cpu.NewCpu(mainBus)
	cpu.ClockCycle()
}
