package ppu

import (
	"github.com/theaaronruss/nes-emulator/internal/cartridge"
)

const (
	FrameWidth       int = 256
	FrameHeight      int = 240
	paletteMemSize   int = 32
	nametableMemSize int = 2048
)

// increment mode offsets
const (
	incrementHorizontal uint16 = 1
	incrementVertical   uint16 = 32
)

// control registers
const (
	ppuCtrl   uint16 = iota
	ppuMask   uint16 = iota
	ppuStatus uint16 = iota
	oamAddr   uint16 = iota
	oamData   uint16 = iota
	ppuScroll uint16 = iota
	ppuAddr   uint16 = iota
	ppuData   uint16 = iota
)

// ppuCtrl flag masks
const (
	flagBaseNameTable  uint8 = 0x03
	flagIncrementMode  uint8 = 0x04
	flagSpritePatTable uint8 = 0x08
	flagBackPatTable   uint8 = 0x10
	flagSpriteSize     uint8 = 0x20
	flagMasterSlave    uint8 = 0x40
	flagVblankEnable   uint8 = 0x80
)

// ppuStatus flag masks
const (
	flagSpriteOverflow uint8 = 0x20
	flagSpriteHit      uint8 = 0x40
	flagVblank         uint8 = 0x80
)

// internal registers
var (
	v uint16 // 15 bits
	t uint16 // 15 bits
	x uint8  // 3 bits
	w bool   // 1 bit

	readBuffer      uint8
	incrementOffset uint16 = incrementHorizontal
	scanline        int
	cycle           int
	vblank          bool
	HandleVblank    bool
)

var (
	paletteMem   = make([]uint8, paletteMemSize)
	nametableMem = make([]uint8, nametableMemSize)
)

var (
	FrameBuffer     []uint8 = make([]uint8, 4*FrameWidth*FrameHeight)
	IsFrameComplete bool
)

func Clock() {
	if cycle == 341 {
		scanline++
		cycle = 0
	}
	if scanline == 262 {
		scanline = 0
		IsFrameComplete = true
	}
	if scanline == 241 && cycle == 1 {
		vblank = true
		HandleVblank = true
	} else if scanline == 261 && cycle == 1 {
		vblank = false
	}
	if cycle >= FrameWidth || scanline >= FrameHeight {
		cycle++
		return
	}

	tileX := cycle / 8
	tileY := scanline / 8
	tileId, _ := getNameTableData(tileX, tileY)
	spriteX := cycle % 8
	spriteY := scanline % 8
	spriteDataAddr := int(tileId)*16 + spriteY
	spriteData1 := internalRead(uint16(spriteDataAddr))
	spriteData2 := internalRead(uint16(spriteDataAddr + 8))
	pixelMask := 1 << (7 - spriteX)
	spriteData1 &= uint8(pixelMask)
	spriteData2 &= uint8(pixelMask)
	spriteData1 >>= (7 - spriteX)
	spriteData2 >>= (7 - spriteX)
	paletteId := spriteData1 | spriteData2
	color := colors[paletteId]
	frameBufferIndex := (scanline*FrameWidth + cycle) * 4
	FrameBuffer[frameBufferIndex] = color.r
	FrameBuffer[frameBufferIndex+1] = color.g
	FrameBuffer[frameBufferIndex+2] = color.b
	FrameBuffer[frameBufferIndex+3] = 0xFF

	cycle++
}

func Read(address uint16) uint8 {
	switch address {
	case ppuStatus:
		w = false
		if vblank {
			return flagVblank
		} else {
			return 0x00
		}
	case oamData:
	case ppuData:
		value := readBuffer
		readBuffer = internalRead(v)
		v += incrementOffset
		return value
	}
	return 0x00
}

func Write(address uint16, data uint8) {
	switch address {
	case ppuCtrl:
	case ppuMask:
	case oamAddr:
	case oamData:
	case ppuScroll:
	case ppuAddr:
		if !w {
			t = (uint16(data) << 8) | (t & 0x00FF)
		} else {
			t |= uint16(data)
			v = t
		}
		w = !w
	case ppuData:
		internalWrite(v, data)
		v += incrementOffset
	}
}

func internalRead(address uint16) uint8 {
	if address <= 0x1FFF {
		return cartridge.ReadCharacterData(address)
	} else if address >= 0x2000 && address <= 0x2FFF {
		offset := (address - 0x2000) % uint16(nametableMemSize)
		return nametableMem[offset]
	} else if address >= 0x3F00 && address <= 0x3FFF {
		offset := (address - 0x3F00) % uint16(paletteMemSize)
		return paletteMem[offset]
	}
	return 0x00
}

func internalWrite(address uint16, data uint8) {
	if address >= 0x2000 && address <= 0x2FFF {
		offset := (address - 0x2000) % uint16(nametableMemSize)
		nametableMem[offset] = data
	} else if address >= 0x3F00 && address <= 0x3FFF {
		offset := address % uint16(paletteMemSize)
		paletteMem[offset] = data
	}
}

func getNameTableData(tileX int, tileY int) (uint8, uint8) {
	offset := tileY*32 + tileX
	return nametableMem[offset], nametableMem[960+offset/8]
}
