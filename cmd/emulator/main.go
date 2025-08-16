package main

import (
	"github.com/theaaronruss/nes-emulator/internal/cpu"
)

func main() {
	cpu := cpu.NewCpu()
	cpu.Cycle()
}
