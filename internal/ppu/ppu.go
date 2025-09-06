package ppu

import "math/rand"

const (
	FrameWidth     int = 256
	FrameHeight    int = 240
	paletteMemSize int = 32
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
)

var (
	paletteMem = make([]uint8, paletteMemSize)
)

var (
	FrameBuffer     []uint8 = make([]uint8, 4*FrameWidth*FrameHeight)
	IsFrameComplete bool
	pixelOffset     int // TODO: remove
)

func Clock() {
	// TODO: remove
	var color uint8
	if rand.Intn(2) == 1 {
		color = 0xFF
	} else {
		color = 0x00
	}
	FrameBuffer[pixelOffset*4] = color
	FrameBuffer[pixelOffset*4+1] = color
	FrameBuffer[pixelOffset*4+2] = color
	FrameBuffer[pixelOffset*4+3] = 0xFF
	pixelOffset++
	if pixelOffset >= FrameWidth*FrameHeight {
		pixelOffset = 0
		IsFrameComplete = true
	}
}

func Read(address uint16) uint8 {
	switch address {
	case ppuStatus:
		w = false
		// TODO: this is temporarily hardcoded for testing
		return flagVblank
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
	if address >= 0x3F00 && address <= 0x3FFF {
		offset := address % uint16(paletteMemSize)
		return paletteMem[offset]
	}
	return 0x00
}

func internalWrite(address uint16, data uint8) {
	if address >= 0x3F00 && address <= 0x3FFF {
		offset := address % uint16(paletteMemSize)
		paletteMem[offset] = data
	}
}
