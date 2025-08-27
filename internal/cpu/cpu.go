package cpu

import (
	"github.com/theaaronruss/nes-emulator/internal/bus"
)

type Cpu struct {
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8

	mainBus *bus.Bus
}

func NewCpu(mainBus *bus.Bus) *Cpu {
	return &Cpu{
		mainBus: mainBus,
	}
}

func (c *Cpu) ClockCycle() {
	// TODO: read instruction at pc
	// TODO: increment program counter
	// TODO: execute instruction
}
