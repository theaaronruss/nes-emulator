package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240
)

// bit masks
const (
	vblankNmiEnableBitMask   uint8  = 0x80
	bgPatternAddrBitMask     uint8  = 0x10
	baseNameTableAddrBitMask uint8  = 0x03
	incrementAmountBitMask   uint8  = 0x04
	vblankBitMask            uint8  = 0x80
	fgEnabledBitMask         uint8  = 0x10
	bgEnabledBitMask         uint8  = 0x08
	nameTableBitMask         uint16 = 0x0C00
	nameTableXBitMask        uint16 = 0x0400
	nameTableYBitMask        uint16 = 0x0800
	coarseXBitMask           uint16 = 0x001F
	coarseYBitMask           uint16 = 0x03E0
	fineYBitMask             uint16 = 0x7000
)

type ppu struct {
	sys                 *System
	frameBuffer         []uint8
	frameComplete       bool
	paletteMem          [32]uint8
	nameTableMem        [2048]uint8
	oamMem              [256]uint8
	cycle               int
	scanLine            int
	oddFrame            bool
	tempAddr            uint16
	vramAddr            uint16
	oamAddr             uint8
	fineX               uint8
	writeToggle         bool
	vblankNmiEnable     bool
	bgPatternAddr       uint16
	incrementAmount     uint16
	vblank              bool
	dataBuffer          uint8
	bgEnabled           bool
	fgEnabled           bool
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
	return &ppu{
		sys:             sys,
		frameBuffer:     make([]uint8, int(FrameWidth)*int(FrameHeight)*4),
		incrementAmount: 1,
	}
}

func (ppu *ppu) Clock() {
	if ppu.cycle == 0 && ppu.scanLine == 261 && ppu.bgEnabled && ppu.oddFrame {
		ppu.cycle++
	}

	if ppu.cycle >= 256 && ppu.cycle <= 320 {
		ppu.writeOamAddr(0)
	}

	if ((ppu.cycle >= 2 && ppu.cycle <= 257) || (ppu.cycle >= 321 && ppu.cycle < 338)) &&
		(ppu.scanLine <= 239 || ppu.scanLine == 261) {
		ppu.shiftShifters()
		switch (ppu.cycle - 1) % 8 {
		case 0:
			ppu.loadIntoShifters()
			ppu.fetchTileId()
		case 2:
			ppu.fetchAttribute()
		case 4:
			ppu.fetchBackgroundLow()
		case 6:
			ppu.fetchBackgroundHigh()
		case 7:
			ppu.incrementCoarseX()
		}
	}

	if ppu.scanLine < 240 || ppu.scanLine == 261 {
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

	if ppu.cycle >= 280 && ppu.cycle <= 304 && ppu.scanLine == 261 {
		ppu.loadYIntoVram()
	}

	if ppu.cycle == 1 && ppu.scanLine == 241 {
		ppu.vblank = true
		if ppu.vblankNmiEnable {
			ppu.sys.cpu.Nmi()
		}
	} else if ppu.cycle == 1 && ppu.scanLine == 261 {
		ppu.vblank = false
	}

	if ppu.cycle < int(FrameWidth) && ppu.scanLine < int(FrameHeight) {
		var color *color
		if ppu.bgEnabled {
			bitMask := uint16(0x8000 >> ppu.fineX)

			var colorIndex int
			if ppu.bgPatternLsbShifter&bitMask > 0 {
				colorIndex |= 0x01
			}
			if ppu.bgPatternMsbShifter&bitMask > 0 {
				colorIndex |= 0x02
			}

			var paletteIndex int
			if ppu.bgAttrLsbShifter&bitMask > 0 {
				paletteIndex |= 0x01
			}
			if ppu.bgAttrMsbShifter&bitMask > 0 {
				paletteIndex |= 0x02
			}

			color = ppu.getColorFromPalette(paletteIndex, colorIndex)
		} else {
			color = &colors[0]
		}
		frameBufferIndex := ppu.scanLine*int(FrameWidth) + ppu.cycle
		frameBufferIndex *= 4
		ppu.frameBuffer[frameBufferIndex] = color.r
		ppu.frameBuffer[frameBufferIndex+1] = color.g
		ppu.frameBuffer[frameBufferIndex+2] = color.b
		ppu.frameBuffer[frameBufferIndex+3] = 0xFF
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

func (ppu *ppu) fetchAttribute() {
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
	if data&vblankNmiEnableBitMask > 0 {
		ppu.vblankNmiEnable = true
	} else {
		ppu.vblankNmiEnable = false
	}

	if data&bgPatternAddrBitMask > 0 {
		ppu.bgPatternAddr = 0x1000
	} else {
		ppu.bgPatternAddr = 0x0000
	}

	if data&incrementAmountBitMask > 0 {
		ppu.incrementAmount = 32
	} else {
		ppu.incrementAmount = 1
	}

	baseNameTableAddr := data & baseNameTableAddrBitMask
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
	case addr <= 0x1FFF:
		return ppu.sys.cartridge.ReadCharacterData(addr)
	case addr >= 0x2000 && addr <= 0x2FFF:
		var nameTableAddr uint16
		if ppu.sys.cartridge.HasHorizontalNameTableMirroring() {
			nameTableAddr = ppu.translateHorizontalNameTableAddr(addr - 0x2000)
		} else {
			nameTableAddr = ppu.translateVerticalNameTableAddr(addr - 0x2000)
		}
		return ppu.nameTableMem[nameTableAddr]
	case addr >= 0x3F00 && addr <= 0x3FFF:
		paletteAddr := (addr - 0x3F00) % 32
		return ppu.paletteMem[paletteAddr]
	}
	return 0
}

func (ppu *ppu) internalWrite(addr uint16, data uint8) {
	switch {
	case addr >= 0x2000 && addr <= 0x2FFF:
		var nameTableAddr uint16
		if ppu.sys.cartridge.HasHorizontalNameTableMirroring() {
			nameTableAddr = ppu.translateHorizontalNameTableAddr(addr - 0x2000)
		} else {
			nameTableAddr = ppu.translateVerticalNameTableAddr(addr - 0x2000)
		}
		ppu.nameTableMem[nameTableAddr] = data
	case addr >= 0x3F00 && addr <= 0x3FFF:
		paletteAddr := (addr - 0x3F00) % 32
		ppu.paletteMem[paletteAddr] = data
	}
}

func (ppu *ppu) translateHorizontalNameTableAddr(addr uint16) uint16 {
	nameTableNum := addr / 0x0400
	offset := addr % 0x0400
	if nameTableNum == 0 || nameTableNum == 2 {
		return offset
	}
	return 0x0400 + offset
}

func (ppu *ppu) translateVerticalNameTableAddr(addr uint16) uint16 {
	nameTableNum := addr / 0x0400
	offset := addr % 0x0400
	if nameTableNum == 0 || nameTableNum == 1 {
		return offset
	}
	return 0x0400 + offset
}
