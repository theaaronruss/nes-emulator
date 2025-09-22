package nes

const (
	FrameWidth       float64 = 256
	FrameHeight      float64 = 240
	clocksPerFrame   int     = 341 * 262
	paletteRamSize   int     = 32
	nameTableRamSize int     = 2048
)

type ppu struct {
	sys          *System
	frameBuffer  []uint8
	currCycle    int
	currScanLine int

	v          uint16
	t          uint16
	w          bool
	vblankFlag bool

	readBuffer   uint8
	paletteRam   [paletteRamSize]uint8
	nameTableRam [nameTableRamSize]uint8
}

func NewPpu(sys *System) *ppu {
	return &ppu{
		sys:         sys,
		frameBuffer: make([]uint8, int(FrameWidth*FrameHeight*4)),
	}
}

func (ppu *ppu) Clock() {
	ppu.renderBackground()

	if ppu.currCycle == 1 && ppu.currScanLine == 241 {
		ppu.vblankFlag = true
		ppu.sys.cpu.Nmi()
	} else if ppu.currCycle == 1 && ppu.currScanLine == 261 {
		ppu.vblankFlag = false
	}

	ppu.currCycle++
	if ppu.currCycle == 341 {
		ppu.currCycle = 0
		ppu.currScanLine++
	}
	if ppu.currScanLine == 262 {
		ppu.currScanLine = 0
	}
}

func (ppu *ppu) renderBackground() {
	if ppu.currCycle >= 256 || ppu.currScanLine >= 240 {
		return
	}
	tileX := ppu.currCycle / 8
	tileY := ppu.currScanLine / 8
	nameTableIndex := tileY*32 + tileX
	tile := ppu.nameTableRam[nameTableIndex]

	charX := ppu.currCycle % 8
	charY := ppu.currScanLine % 8
	value := ppu.characterPixel(int(tile), charX, charY)

	color := ppu.characterColor(0, int(value))
	frameBufferIndex := (ppu.currScanLine*int(FrameWidth) + ppu.currCycle) * 4
	ppu.frameBuffer[frameBufferIndex] = color.r
	ppu.frameBuffer[frameBufferIndex+1] = color.g
	ppu.frameBuffer[frameBufferIndex+2] = color.b
	ppu.frameBuffer[frameBufferIndex+3] = 0xFF
}

func (ppu *ppu) characterPixel(char int, charX int, charY int) uint8 {
	index := uint16(char*16 + charY)
	plane1 := ppu.internalRead(index)
	plane2 := ppu.internalRead(index + 8)
	plane1 >>= 7 - charX
	plane2 >>= 7 - charX
	plane1 &= 0x01
	plane2 &= 0x01
	value := (plane2 << 1) | plane1
	return value
}

func (ppu *ppu) characterColor(palette int, index int) *color {
	paletteIndex := palette*4 + index
	colorIndex := ppu.paletteRam[paletteIndex]
	return &colors[colorIndex]
}

func (ppu *ppu) writePpuCtrl(data uint8) {
}

func (ppu *ppu) writePpuMask(data uint8) {
}

func (ppu *ppu) readPpuStatus() uint8 {
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
	ppu.v++
	return data
}

func (ppu *ppu) writePpuData(data uint8) {
	ppu.internalWrite(ppu.v, data)
	ppu.v++
}

func (ppu *ppu) internalRead(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF:
		// character data
		return ppu.sys.cartridge.ReadCharacterData(addr)
	case addr >= 0x2000 && addr <= 0x2FFF:
		// name table ram
		nameTableAddr := (addr - 0x2000) % uint16(nameTableRamSize)
		return ppu.nameTableRam[nameTableAddr]
	case addr >= 0x3F00 && addr <= 0x3FFF:
		// palette ram
		paletteAddr := (addr - 0x3F00) % uint16(paletteRamSize)
		return ppu.paletteRam[paletteAddr]
	default:
		return 0
	}
}

func (ppu *ppu) internalWrite(addr uint16, data uint8) {
	switch {
	case addr >= 0x2000 && addr <= 0x2FFF:
		// name table ram
		nameTableAddr := (addr - 0x2000) % uint16(nameTableRamSize)
		ppu.nameTableRam[nameTableAddr] = data
	case addr >= 0x3F00 && addr <= 0x3FFF:
		// palette ram
		paletteAddr := (addr - 0x3F00) % uint16(paletteRamSize)
		ppu.paletteRam[paletteAddr] = data
	}
}
