package sysbus

import "github.com/theaaronruss/nes-emulator/internal/cartridge"

const (
	cpuRamSize uint16 = 2048

	cpuRamAddr  uint16 = 0x0000
	ppuRegsAddr uint16 = 0x2000
)

var cpuRam [cpuRamSize]uint8
var GamePak *cartridge.Cartridge

func Read(address uint16) uint8 {
	if address >= cpuRamAddr && address < ppuRegsAddr {
		return cpuRam[address%cpuRamSize]
	}
	return 0x00
}

func Write(address uint16, data uint8) {
	if address >= cpuRamAddr && address < ppuRegsAddr {
		cpuRam[address%cpuRamSize] = data
	}
}
