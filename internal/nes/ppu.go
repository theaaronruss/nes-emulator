package nes

const (
	FrameWidth       float64 = 256
	FrameHeight      float64 = 240
	incrementHor     uint16  = 1
	incrementVer     uint16  = 32
	paletteMemSize   int     = 32
	nameTableMemSize int     = 2048
	oamMemSize       int     = 256
	nameTableSize    uint16  = 0x0400
)

// ppuctrl bit masks
const (
	nmiEnableBitMask         uint8 = 0x80
	tallSpritesBitMask       uint8 = 0x20
	bgPatternAddrBitMask     uint8 = 0x10
	vramIncrementBitMask     uint8 = 0x04
	baseNameTableAddrBitMask uint8 = 0x03
)

// ppumask bit masks
const (
	fgEnabledBitMask uint8 = 0x10
	bgEnabledBitMask uint8 = 0x08
)

// ppustatus bit mask
const (
	vblankBitMask         uint8 = 0x80
	spriteOverflowBitMask uint8 = 0x20
)

// vram bit masks
const (
	nameTableXBitMask uint16 = 0x0400
	nameTableYBitMask uint16 = 0x0800
	coarseXBitMask    uint16 = 0x001F
	coarseYBitMask    uint16 = 0x03E0
	fineYBitMask      uint16 = 0x7000
)

// addresses
const (
	cartridgeAddrEnd   uint16 = 0x1FFF
	nameTableAddrStart uint16 = 0x2000
	nameTableAddrEnd   uint16 = 0x2FFF
	paletteAddrStart   uint16 = 0x3F00
	paletteAddrEnd     uint16 = 0x3FFF
)

type ppu struct {
	sys           *System
	frameBuffer   []uint8
	frameComplete bool

	paletteMem   [paletteMemSize]uint8
	nameTableMem [nameTableMemSize]uint8
	oamMem       [oamMemSize]uint8
	secondOamMem [32]uint8
	spriteCount  int
	dataBuffer   uint8

	cycle          int
	scanLine       int
	oddFrame       bool
	vblank         bool
	spriteOverflow bool
	tempAddr       uint16
	vramAddr       uint16
	oamAddr        uint8
	fineX          uint8
	writeToggle    bool

	vblankNmiEnable bool
	spriteHeight    int
	bgPatternAddr   uint16
	incrementAmount uint16
	bgEnabled       bool
	fgEnabled       bool

	bgTileId            uint8
	bgTileAttr          uint8
	bgTileLsb           uint8
	bgTileMsb           uint8
	bgPatternLsbShifter uint16
	bgPatternMsbShifter uint16
	bgAttrLsbShifter    uint16
	bgAttrMsbShifter    uint16
}

func NewPpu(sys *System) *ppu {
	ppu := &ppu{
		sys:             sys,
		frameBuffer:     make([]uint8, int(FrameWidth)*int(FrameHeight)*4),
		incrementAmount: 1,
	}
	ppu.clearSecondOamMem()
	return ppu
}

func (ppu *ppu) clearSecondOamMem() {
	ppu.spriteCount = 0
	for i := range len(ppu.secondOamMem) {
		ppu.secondOamMem[i] = 0xFF
	}
}

func (ppu *ppu) Clock() {
	isRenderingEnabled := ppu.bgEnabled || ppu.fgEnabled

	// skip cycle on odd frames
	if ppu.cycle == 0 && ppu.scanLine == 261 &&
		isRenderingEnabled && ppu.oddFrame {
		ppu.cycle++
	}

	if ppu.bgEnabled &&
		((ppu.cycle >= 2 && ppu.cycle <= 257) || (ppu.cycle >= 321 && ppu.cycle < 338)) &&
		(ppu.scanLine < 240 || ppu.scanLine == 261) {
		ppu.shiftShifters()
		switch (ppu.cycle - 1) % 8 {
		case 0:
			ppu.loadIntoShifters()
			ppu.fetchTileId()
		case 2:
			ppu.fetchTileAttribute()
		case 4:
			ppu.fetchBackgroundLow()
		case 6:
			ppu.fetchBackgroundHigh()
		case 7:
			ppu.incrementCoarseX()
		}
	}

	if ppu.bgEnabled && ppu.scanLine < 240 || ppu.scanLine == 261 {
		if ppu.cycle == 256 {
			ppu.incrementFineY()
		}

		if ppu.cycle == 257 {
			ppu.loadIntoShifters()
			ppu.loadXIntoVram()
		}

		if ppu.cycle == 338 || ppu.cycle == 340 {
			ppu.fetchTileId()
		}
	}

	if ppu.bgEnabled && ppu.scanLine == 261 && ppu.cycle >= 280 &&
		ppu.cycle <= 304 {
		ppu.loadYIntoVram()
	}

	if ppu.fgEnabled && ppu.scanLine < 240 && ppu.cycle == 0 {
		ppu.spriteEvaluation()
	}

	// draw visible pixels
	if ppu.cycle < int(FrameWidth) && ppu.scanLine < int(FrameHeight) {
		var bgPaletteIndex int
		var fgPaletteIndex int
		var bgColorIndex int
		var fgColorIndex int
		var priority bool
		if ppu.bgEnabled {
			bgPaletteIndex = ppu.getBackgroundPaletteIndex()
			bgColorIndex = ppu.getBackgroundColorIndex()
		}
		if ppu.fgEnabled {
			for spriteNum := ppu.spriteCount - 1; spriteNum >= 0; spriteNum-- {
				spriteX := int(ppu.secondOamMem[spriteNum*4+3])
				priority = ppu.secondOamMem[spriteNum*4+2]&0x20 == 0
				fgPaletteIndex = int(ppu.secondOamMem[spriteNum*4+2]&0x02) + 4
				if ppu.cycle >= spriteX && ppu.cycle < spriteX+8 {
					fgColorIndex = 1
					break
				}
			}
		}

		paletteIndex := bgPaletteIndex
		colorIndex := bgColorIndex
		if bgColorIndex > 0 && fgColorIndex > 0 {
			if priority {
				paletteIndex = fgPaletteIndex
				colorIndex = fgColorIndex
			} else {
				paletteIndex = bgPaletteIndex
				colorIndex = bgColorIndex
			}
		} else if fgColorIndex > 0 {
			paletteIndex = fgPaletteIndex
			colorIndex = fgColorIndex
		}
		color := ppu.getColorFromPalette(paletteIndex, colorIndex)
		dot := (ppu.scanLine*int(FrameWidth) + ppu.cycle) * 4
		ppu.frameBuffer[dot] = color.r
		ppu.frameBuffer[dot+1] = color.g
		ppu.frameBuffer[dot+2] = color.b
		ppu.frameBuffer[dot+3] = 0xFF
	}

	if ppu.scanLine == 241 && ppu.cycle == 1 {
		ppu.vblank = true
		if ppu.vblankNmiEnable {
			ppu.sys.cpu.Nmi()
		}
	} else if ppu.cycle == 1 && ppu.scanLine == 261 {
		ppu.vblank = false
		ppu.spriteOverflow = false
	}

	ppu.cycle++
	if ppu.cycle > 340 {
		ppu.cycle = 0
		ppu.scanLine++
		if ppu.scanLine > 261 {
			ppu.scanLine = 0
			ppu.frameComplete = true
		}
	}
}

func (ppu *ppu) spriteEvaluation() {
	ppu.clearSecondOamMem()
	for spriteIndex := range len(ppu.oamMem) / 4 {
		spriteY := int(ppu.oamMem[spriteIndex*4])
		if ppu.scanLine >= spriteY+1 && ppu.scanLine < spriteY+9 {
			ppu.copyToSecondOamMem(spriteIndex)
			ppu.spriteCount++
			if ppu.spriteCount >= 8 {
				break
			}
		}
	}
}

func (ppu *ppu) copyToSecondOamMem(spriteIndex int) {
	for i := range 4 {
		ppu.secondOamMem[ppu.spriteCount*4+i] = ppu.oamMem[spriteIndex*4+i]
	}
}

func (ppu *ppu) getBackgroundPaletteIndex() int {
	if !ppu.bgEnabled {
		return 0
	}

	bitMask := uint16(0x8000 >> ppu.fineX)
	var paletteIndex int
	if ppu.bgAttrLsbShifter&bitMask > 0 {
		paletteIndex |= 0x01
	}
	if ppu.bgAttrMsbShifter&bitMask > 0 {
		paletteIndex |= 0x02
	}
	return paletteIndex
}

func (ppu *ppu) getBackgroundColorIndex() int {
	if !ppu.bgEnabled {
		return 0
	}

	bitMask := uint16(0x8000 >> ppu.fineX)
	var colorIndex int
	if ppu.bgPatternLsbShifter&bitMask > 0 {
		colorIndex |= 0x01
	}
	if ppu.bgPatternMsbShifter&bitMask > 0 {
		colorIndex |= 0x02
	}
	return colorIndex
}

func (ppu *ppu) loadIntoShifters() {
	ppu.bgPatternLsbShifter &= 0xFF00
	ppu.bgPatternLsbShifter |= uint16(ppu.bgTileLsb)
	ppu.bgPatternMsbShifter &= 0xFF00
	ppu.bgPatternMsbShifter |= uint16(ppu.bgTileMsb)

	ppu.bgAttrLsbShifter &= 0xFF00
	if ppu.bgTileAttr&0x01 > 0 {
		ppu.bgAttrLsbShifter |= 0x00FF
	}
	ppu.bgAttrMsbShifter &= 0xFF00
	if ppu.bgTileAttr&0x02 > 0 {
		ppu.bgAttrMsbShifter |= 0x00FF
	}
}

func (ppu *ppu) loadXIntoVram() {
	if !ppu.fgEnabled && !ppu.bgEnabled {
		return
	}

	coarseX := ppu.tempAddr & coarseXBitMask
	nameTableX := (ppu.tempAddr & nameTableXBitMask) >> 10
	ppu.vramAddr &= ^coarseXBitMask
	ppu.vramAddr &= ^nameTableXBitMask
	ppu.vramAddr |= coarseX
	ppu.vramAddr |= nameTableX << 10
}

func (ppu *ppu) loadYIntoVram() {
	if !ppu.fgEnabled && !ppu.bgEnabled {
		return
	}

	coarseY := (ppu.tempAddr & coarseYBitMask) >> 5
	nameTableY := (ppu.tempAddr & nameTableYBitMask) >> 11
	fineY := (ppu.tempAddr & fineYBitMask) >> 12
	ppu.vramAddr &= ^coarseYBitMask
	ppu.vramAddr &= ^nameTableYBitMask
	ppu.vramAddr &= ^fineYBitMask
	ppu.vramAddr |= coarseY << 5
	ppu.vramAddr |= nameTableY << 11
	ppu.vramAddr |= fineY << 12
}

func (ppu *ppu) shiftShifters() {
	if !ppu.bgEnabled {
		return
	}

	ppu.bgPatternLsbShifter <<= 1
	ppu.bgPatternMsbShifter <<= 1
	ppu.bgAttrLsbShifter <<= 1
	ppu.bgAttrMsbShifter <<= 1
}

func (ppu *ppu) incrementCoarseX() {
	if !ppu.fgEnabled && !ppu.bgEnabled {
		return
	}

	coarseX := ppu.vramAddr & coarseXBitMask
	if coarseX == 31 {
		coarseX = 0
		ppu.vramAddr ^= nameTableXBitMask
	} else {
		coarseX++
	}
	ppu.vramAddr &= ^coarseXBitMask
	ppu.vramAddr |= coarseX
}

func (ppu *ppu) incrementFineY() {
	if !ppu.fgEnabled && !ppu.bgEnabled {
		return
	}

	fineY := (ppu.vramAddr & fineYBitMask) >> 12
	if fineY < 7 {
		fineY++
	} else {
		fineY = 0
		coarseY := (ppu.vramAddr & coarseYBitMask) >> 5
		switch coarseY {
		case 29:
			coarseY = 0
			ppu.vramAddr ^= nameTableYBitMask
		case 31:
			coarseY = 0
		default:
			coarseY++
		}
		ppu.vramAddr &= ^coarseYBitMask
		ppu.vramAddr |= coarseY << 5
	}
	ppu.vramAddr &= ^fineYBitMask
	ppu.vramAddr |= fineY << 12
}

func (ppu *ppu) fetchTileId() {
	addr := 0x2000 | (ppu.vramAddr & 0x0FFF)
	ppu.bgTileId = ppu.internalRead(addr)
}

func (ppu *ppu) fetchTileAttribute() {
	addr := 0x23C0 | (ppu.vramAddr & 0x0C00) | ((ppu.vramAddr >> 4) & 0x38) |
		((ppu.vramAddr >> 2) & 0x07)
	attr := ppu.internalRead(addr)
	coarseX := ppu.vramAddr & coarseXBitMask
	coarseY := (ppu.vramAddr & coarseYBitMask) >> 5

	if coarseX&0x02 != 0 {
		attr >>= 2
	}
	if coarseY&0x02 != 0 {
		attr >>= 4
	}

	ppu.bgTileAttr = attr & 0x03
}

func (ppu *ppu) fetchBackgroundLow() {
	fineY := ppu.vramAddr & fineYBitMask >> 12
	addr := ppu.bgPatternAddr + (uint16(ppu.bgTileId) * 16) + fineY
	ppu.bgTileLsb = ppu.internalRead(addr)
}

func (ppu *ppu) fetchBackgroundHigh() {
	fineY := ppu.vramAddr & fineYBitMask >> 12
	addr := ppu.bgPatternAddr + (uint16(ppu.bgTileId) * 16) + fineY + 8
	ppu.bgTileMsb = ppu.internalRead(addr)
}

func (ppu *ppu) getColorFromPalette(paletteIndex int, colorIndex int) *color {
	paletteIndex &= 0x0007
	colorIndex &= 0x0003
	index := paletteIndex*4 + colorIndex
	colorCode := ppu.paletteMem[index]
	return &colors[colorCode]
}

func (ppu *ppu) readPpuStatus() uint8 {
	var status uint8

	if ppu.vblank {
		status |= vblankBitMask
	}

	if ppu.spriteOverflow {
		status |= spriteOverflowBitMask
	}

	ppu.writeToggle = false
	return status
}

func (ppu *ppu) readOamData() uint8 {
	return 0
}

func (ppu *ppu) readPpuData() uint8 {
	data := ppu.dataBuffer
	ppu.dataBuffer = ppu.internalRead(ppu.vramAddr)
	ppu.vramAddr += ppu.incrementAmount
	return data
}

func (ppu *ppu) writePpuCtrl(data uint8) {
	if data&nmiEnableBitMask > 0 {
		ppu.vblankNmiEnable = true
	} else {
		ppu.vblankNmiEnable = false
	}

	if data&tallSpritesBitMask > 0 {
		ppu.spriteHeight = 16
	} else {
		ppu.spriteHeight = 8
	}

	if data&bgPatternAddrBitMask > 0 {
		ppu.bgPatternAddr = 0x1000
	} else {
		ppu.bgPatternAddr = 0x0000
	}

	if data&vramIncrementBitMask > 0 {
		ppu.incrementAmount = uint16(incrementVer)
	} else {
		ppu.incrementAmount = uint16(incrementHor)
	}

	baseNameTableAddr := data & baseNameTableAddrBitMask
	nameTableBitMask := nameTableXBitMask | nameTableYBitMask
	ppu.tempAddr &= ^nameTableBitMask
	ppu.tempAddr |= uint16(baseNameTableAddr) << 10
}

func (ppu *ppu) writePpuMask(data uint8) {
	if data&fgEnabledBitMask > 0 {
		ppu.fgEnabled = true
	} else {
		ppu.fgEnabled = false
	}

	if data&bgEnabledBitMask > 0 {
		ppu.bgEnabled = true
	} else {
		ppu.bgEnabled = false
	}
}

func (ppu *ppu) writeOamAddr(data uint8) {
	ppu.oamAddr = data
}

func (ppu *ppu) writeOamData(data uint8) {
	ppu.oamMem[ppu.oamAddr] = data
}

func (ppu *ppu) writePpuScroll(data uint8) {
	if !ppu.writeToggle {
		coarseX := (data & 0xF8) >> 3
		ppu.tempAddr &= ^coarseXBitMask
		ppu.tempAddr |= uint16(coarseX)
		ppu.fineX = data & 0x07
	} else {
		coarseY := (data & 0xF8) >> 3
		fineY := data & 0x07
		ppu.tempAddr &= ^coarseYBitMask
		ppu.tempAddr |= uint16(coarseY) << 5
		ppu.tempAddr &= ^fineYBitMask
		ppu.tempAddr |= uint16(fineY) << 12
	}
	ppu.writeToggle = !ppu.writeToggle
}

func (ppu *ppu) writePpuAddr(data uint8) {
	if !ppu.writeToggle {
		data &= 0x3F
		ppu.tempAddr &= 0x00FF
		ppu.tempAddr |= uint16(data) << 8
	} else {
		ppu.tempAddr &= 0xFF00
		ppu.tempAddr |= uint16(data)
		ppu.vramAddr = ppu.tempAddr
	}
	ppu.writeToggle = !ppu.writeToggle
}

func (ppu *ppu) writePpuData(data uint8) {
	ppu.internalWrite(ppu.vramAddr, data)
	ppu.vramAddr += ppu.incrementAmount
}

func (ppu *ppu) writeOamDma(data uint8) {
	for i := range len(ppu.oamMem) {
		addr := uint16(data)<<8 + uint16(i)
		ppu.oamMem[i] = ppu.sys.read(addr)
	}
}

func (ppu *ppu) internalRead(addr uint16) uint8 {
	switch {
	case addr <= cartridgeAddrEnd:
		return ppu.sys.cartridge.ReadCharacterData(addr)
	case addr >= nameTableAddrStart && addr <= nameTableAddrEnd:
		var nameTableAddr uint16
		if ppu.sys.cartridge.HasHorizontalNameTableMirroring() {
			nameTableAddr = ppu.translateHorizontalNameTableAddr(addr - nameTableAddrStart)
		} else {
			nameTableAddr = ppu.translateVerticalNameTableAddr(addr - nameTableAddrStart)
		}
		return ppu.nameTableMem[nameTableAddr]
	case addr >= paletteAddrStart && addr <= paletteAddrEnd:
		paletteAddr := (addr - paletteAddrStart) % uint16(paletteMemSize)
		return ppu.paletteMem[paletteAddr]
	}
	return 0
}

func (ppu *ppu) internalWrite(addr uint16, data uint8) {
	switch {
	case addr >= nameTableAddrStart && addr <= nameTableAddrEnd:
		var nameTableAddr uint16
		if ppu.sys.cartridge.HasHorizontalNameTableMirroring() {
			nameTableAddr = ppu.translateHorizontalNameTableAddr(addr - nameTableAddrStart)
		} else {
			nameTableAddr = ppu.translateVerticalNameTableAddr(addr - nameTableAddrStart)
		}
		ppu.nameTableMem[nameTableAddr] = data
	case addr >= paletteAddrStart && addr <= paletteAddrEnd:
		paletteAddr := (addr - paletteAddrStart) % uint16(paletteMemSize)
		ppu.paletteMem[paletteAddr] = data
	}
}

func (ppu *ppu) translateHorizontalNameTableAddr(addr uint16) uint16 {
	nameTableIndex := addr / nameTableSize
	offset := addr % nameTableSize
	if nameTableIndex == 0 || nameTableIndex == 2 {
		return offset
	}
	return nameTableSize + offset
}

func (ppu *ppu) translateVerticalNameTableAddr(addr uint16) uint16 {
	nameTableIndex := addr / nameTableSize
	offset := addr % nameTableSize
	if nameTableIndex == 0 || nameTableIndex == 1 {
		return offset
	}
	return nameTableSize + offset
}
