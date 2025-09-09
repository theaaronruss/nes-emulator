package sysbus

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/theaaronruss/nes-emulator/internal/cartridge"
	"github.com/theaaronruss/nes-emulator/internal/ppu"
)

const (
	cpuRamSize uint16 = 2048

	cpuRamAddr  uint16 = 0x0000
	ppuRegsAddr uint16 = 0x2000
	cartRomAddr uint16 = 0x8000
)

// controller button masks
const (
	btnA      uint8 = 1 << iota
	btnB      uint8 = 1 << iota
	btnSelect uint8 = 1 << iota
	btnStart  uint8 = 1 << iota
	btnUp     uint8 = 1 << iota
	btnDown   uint8 = 1 << iota
	btnLeft   uint8 = 1 << iota
	btnRight  uint8 = 1 << iota
)

var (
	Window         *opengl.Window
	cpuRam         [cpuRamSize]uint8
	controllerData uint8
)

func Read(address uint16) uint8 {
	if address >= cpuRamAddr && address < ppuRegsAddr {
		return cpuRam[address%cpuRamSize]
	} else if address >= 0x2000 && address <= 0x2007 {
		return ppu.Read(address - 0x2000)
	} else if address == 0x4016 {
		data := controllerData
		controllerData >>= 1
		controllerData |= 0x80
		return data
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
	} else if address == 0x4016 {
		controllerData = encodeControllerButtons()
	}
}

func encodeControllerButtons() uint8 {
	if Window == nil {
		return 0
	}
	var buttons uint8 = 0
	if Window.Pressed(pixel.KeyPeriod) {
		buttons |= btnA
	}
	if Window.Pressed(pixel.KeyComma) {
		buttons |= btnB
	}
	if Window.Pressed(pixel.KeyK) {
		buttons |= btnSelect
	}
	if Window.Pressed(pixel.KeyL) {
		buttons |= btnStart
	}
	if Window.Pressed(pixel.KeyW) {
		buttons |= btnUp
	}
	if Window.Pressed(pixel.KeyS) {
		buttons |= btnDown
	}
	if Window.Pressed(pixel.KeyA) {
		buttons |= btnLeft
	}
	if Window.Pressed(pixel.KeyD) {
		buttons |= btnRight
	}
	return buttons
}
