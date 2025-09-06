package sysbus

import (
	"github.com/theaaronruss/nes-emulator/internal/cartridge"
	"github.com/theaaronruss/nes-emulator/internal/ppu"
)

const (
	cpuRamSize uint16 = 2048

	cpuRamAddr  uint16 = 0x0000
	ppuRegsAddr uint16 = 0x2000
	cartRomAddr uint16 = 0x8000
)

var cpuRam [cpuRamSize]uint8

func Read(address uint16) uint8 {
	if address >= cpuRamAddr && address < ppuRegsAddr {
		return cpuRam[address%cpuRamSize]
	} else if address >= 0x2000 && address <= 0x2007 {
		return ppu.Read(address - 0x2000)
	} else if address >= cartRomAddr {
		return cartridge.ReadProgramData(address - cartRomAddr)
	}
	return 0x00
}

func Write(address uint16, data uint8) {
	if address >= cpuRamAddr && address < ppuRegsAddr {
		cpuRam[address%cpuRamSize] = data
	} else if address >= 0x2000 && address <= 0x2007 {
		ppu.Write(address-0x2000, data)
	}
}
