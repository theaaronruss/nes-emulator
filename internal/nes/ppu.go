package nes

const (
	FrameWidth     float64 = 256
	FrameHeight    float64 = 240
	clocksPerFrame int     = 341 * 262
)

type ppu struct {
	sys         *System
	frameBuffer []uint8
	cycle       int
	scanLine    int

	v          uint16
	t          uint16
	w          bool
	vblankFlag bool

	vblankCtrl    bool
	vramIncrement uint16

	readBuffer   uint8
	paletteRam   [32]uint8
	nameTableRam [2048]uint8
}

func NewPpu(sys *System) *ppu {
	return &ppu{
		sys:           sys,
		frameBuffer:   make([]uint8, int(FrameWidth*FrameHeight*4)),
		vblankCtrl:    true,
		vramIncrement: 1,
	}
}

func (ppu *ppu) Clock() {
	ppu.renderBackground()

	if ppu.cycle == 1 && ppu.scanLine == 241 && ppu.vblankCtrl {
		ppu.vblankFlag = true
		ppu.sys.cpu.Nmi()
	} else if ppu.cycle == 1 && ppu.scanLine == 261 {
		ppu.vblankFlag = false
	}

	ppu.cycle++
	if ppu.cycle == 341 {
		ppu.cycle = 0
		ppu.scanLine++
	}
	if ppu.scanLine == 262 {
		ppu.scanLine = 0
	}
}

func (ppu *ppu) renderBackground() {
	if ppu.cycle >= int(FrameWidth) || ppu.scanLine >= int(FrameHeight) {
		return
	}

	tileX := ppu.cycle / 8
	tileY := ppu.scanLine / 8
	tileId := ppu.getTileId(tileX, tileY)
	paletteId := ppu.getPaletteId(tileX, tileY)
	color := ppu.getColorFromPalette(tileId, paletteId)

	frameBufferIndex := (ppu.scanLine*int(FrameWidth) + ppu.cycle) * 4
	ppu.frameBuffer[frameBufferIndex] = color.r
	ppu.frameBuffer[frameBufferIndex+1] = color.g
	ppu.frameBuffer[frameBufferIndex+2] = color.b
	ppu.frameBuffer[frameBufferIndex+3] = 0xFF
}

func (ppu *ppu) getTileId(tileX int, tileY int) uint8 {
	if tileX < 0 || tileX >= 32 ||
		tileY < 0 || tileY >= 30 {
		return 0
	}

	index := tileY*32 + tileX
	return ppu.nameTableRam[index]
}

func (ppu *ppu) getPaletteId(tileX int, tileY int) uint8 {
	if tileX < 0 || tileX >= 32 ||
		tileY < 0 || tileY >= 30 {
		return 0
	}

	attrX := tileX / 4
	attrY := tileY / 4
	index := attrY*8 + attrX
	attr := ppu.nameTableRam[960+index]

	quadX := attrX % 2
	quadY := attrY % 2
	switch {
	case quadX == 0 && quadY == 0:
		return attr & 0x03
	case quadX == 1 && quadY == 0:
		return (attr >> 2) & 0x03
	case quadX == 0 && quadY == 1:
		return (attr >> 4) & 0x03
	case quadX == 1 && quadY == 1:
		return (attr >> 6) & 0x03
	}
	return 0
}

func (ppu *ppu) getColorFromPalette(tileIndex uint8, attr uint8) *color {
	tileX := ppu.cycle % 8
	tileY := ppu.scanLine % 8

	patternAddr := (uint16(tileIndex) * 16) + uint16(tileY)
	low := ppu.sys.cartridge.ReadCharacterData(patternAddr)
	high := ppu.sys.cartridge.ReadCharacterData(patternAddr + 8)

	shift := 7 - tileX
	lowBit := (low >> shift) & 0x01
	highBit := (high >> shift) & 0x01
	pixel := uint16((highBit << 1) | lowBit)

	paletteNum := attr & 0x03
	paletteAddr := 0x3F00 + uint16(paletteNum)*4 + pixel
	colorIndex := ppu.internalRead(paletteAddr) & 0x3F
	return &colors[colorIndex]
}

func (ppu *ppu) writePpuCtrl(data uint8) {
	if data&0x80 > 0 {
		if !ppu.vblankCtrl && ppu.vblankFlag {
			ppu.sys.cpu.Nmi()
		}
		ppu.vblankCtrl = true
	}

	if data&0x04 > 0 {
		ppu.vramIncrement = 32
	} else {
		ppu.vramIncrement = 1
	}
}

func (ppu *ppu) writePpuMask(data uint8) {
}

func (ppu *ppu) readPpuStatus() uint8 {
	ppu.w = false
	if ppu.vblankFlag {
		return 0x80
	} else {
		return 0x00
	}
}

func (ppu *ppu) writeOamAddr(data uint8) {
}

func (ppu *ppu) readOamData() uint8 {
	return 0
}

func (ppu *ppu) writeOamData(data uint8) {
}

func (ppu *ppu) writePpuScroll(data uint8) {
}

func (ppu *ppu) writePpuAddr(data uint8) {
	if !ppu.w {
		ppu.t = uint16(data) << 8
	} else {
		ppu.t |= uint16(data)
		ppu.v = ppu.t
	}
	ppu.w = !ppu.w
}

func (ppu *ppu) readPpuData() uint8 {
	data := ppu.readBuffer
	ppu.readBuffer = ppu.internalRead(ppu.v)
	ppu.v += ppu.vramIncrement
	return data
}

func (ppu *ppu) writePpuData(data uint8) {
	ppu.internalWrite(ppu.v, data)
	ppu.v += ppu.vramIncrement
}

func (ppu *ppu) internalRead(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF:
		// character data
		return ppu.sys.cartridge.ReadCharacterData(addr)
	case addr >= 0x2000 && addr <= 0x2FFF:
		// name table ram
		nameTableAddr := (addr - 0x2000) % 2048
		return ppu.nameTableRam[nameTableAddr]
	case addr >= 0x3F00 && addr <= 0x3FFF:
		// palette ram
		paletteAddr := (addr - 0x3F00) % 32
		return ppu.paletteRam[paletteAddr]
	default:
		return 0
	}
}

func (ppu *ppu) internalWrite(addr uint16, data uint8) {
	switch {
	case addr >= 0x2000 && addr <= 0x2FFF:
		// name table ram
		nameTableAddr := (addr - 0x2000) % 2048
		ppu.nameTableRam[nameTableAddr] = data
	case addr >= 0x3F00 && addr <= 0x3FFF:
		// palette ram
		paletteAddr := (addr - 0x3F00) % 32
		ppu.paletteRam[paletteAddr] = data
	}
}
