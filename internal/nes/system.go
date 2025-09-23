package nes

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

const (
	cpuRamStartAddr uint16 = 0x0000
	cpuRamEndAddr   uint16 = 0x07FF
	cpuRamSize      uint16 = cpuRamEndAddr - cpuRamStartAddr + 1
)

// ppu registers
const (
	ppuCtrl   uint16 = 0x2000
	ppuMask   uint16 = 0x2001
	ppuStatus uint16 = 0x2002
	oamAddr   uint16 = 0x2003
	oamData   uint16 = 0x2004
	ppuScroll uint16 = 0x2005
	ppuAddr   uint16 = 0x2006
	ppuData   uint16 = 0x2007
)

// controller buttons
const (
	btnA      = 1 << iota
	btnB      = 1 << iota
	btnSelect = 1 << iota
	btnStart  = 1 << iota
	btnUp     = 1 << iota
	btnDown   = 1 << iota
	btnLeft   = 1 << iota
	btnRight  = 1 << iota
)

type System struct {
	cpu            *cpu
	ppu            *ppu
	cpuRam         [cpuRamSize]uint8
	controllerData uint8
	win            *opengl.Window
	cartridge      *Cartridge

	ppuClocks int
}

func NewSystem(win *opengl.Window, cartridge *Cartridge) *System {
	sys := &System{
		win:       win,
		cartridge: cartridge,
	}
	sys.ppu = NewPpu(sys)
	sys.cpu = NewCpu(sys)
	return sys
}

func (sys *System) FrameBuffer() []uint8 {
	return sys.ppu.frameBuffer
}

func (sys *System) ClockFrame() {
	for range clocksPerFrame {
		sys.ppu.Clock()
		sys.ppuClocks++
		if sys.ppuClocks >= 3 {
			sys.ppuClocks = 0
			sys.cpu.Clock()
		}
	}
}

func (sys *System) read(addr uint16) uint8 {
	switch {
	case addr <= 0x07FF:
		return sys.cpuRam[addr]
	case addr == ppuStatus:
		return sys.ppu.readPpuStatus()
	case addr == oamData:
		return sys.ppu.readOamData()
	case addr == ppuData:
		return sys.ppu.readPpuData()
	case addr == 0x4016:
		data := sys.controllerData & 0x01
		sys.controllerData >>= 1
		sys.controllerData |= 0x80
		return data
	case sys.cartridge != nil && addr >= 0x8000:
		return sys.cartridge.ReadProgramData(addr)
	default:
		return 0
	}
}

func (sys *System) write(addr uint16, data uint8) {
	switch {
	case addr >= cpuRamStartAddr && addr <= cpuRamEndAddr:
		sys.cpuRam[addr] = data
	case addr == ppuCtrl:
		sys.ppu.writePpuCtrl(data)
	case addr == ppuMask:
		sys.ppu.writePpuMask(data)
	case addr == oamAddr:
		sys.ppu.writeOamAddr(data)
	case addr == oamData:
		sys.ppu.writeOamData(data)
	case addr == ppuScroll:
		sys.ppu.writePpuScroll(data)
	case addr == ppuAddr:
		sys.ppu.writePpuAddr(data)
	case addr == ppuData:
		sys.ppu.writePpuData(data)
	case addr == 0x4016:
		if data&0x01 > 0 {
			sys.updateControllerInput()
		}
	}
}

func (sys *System) updateControllerInput() {
	sys.controllerData = 0

	if sys.win.Pressed(pixel.KeyPeriod) {
		sys.controllerData |= btnA
	}

	if sys.win.Pressed(pixel.KeyComma) {
		sys.controllerData |= btnB
	}

	if sys.win.Pressed(pixel.KeyK) {
		sys.controllerData |= btnSelect
	}

	if sys.win.Pressed(pixel.KeyL) {
		sys.controllerData |= btnStart
	}

	if sys.win.Pressed(pixel.KeyW) {
		sys.controllerData |= btnUp
	}

	if sys.win.Pressed(pixel.KeyS) {
		sys.controllerData |= btnDown
	}

	if sys.win.Pressed(pixel.KeyA) {
		sys.controllerData |= btnLeft
	}

	if sys.win.Pressed(pixel.KeyD) {
		sys.controllerData |= btnRight
	}
}
