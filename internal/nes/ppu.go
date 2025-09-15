package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240
)

type Ppu struct {
	FrameBuffer   []uint8
	FrameComplete bool
}

func NewPpu() *Ppu {
	return &Ppu{
		FrameBuffer:   make([]uint8, 4*int(FrameWidth)*int(FrameHeight)),
		FrameComplete: false,
	}
}

func (ppu *Ppu) internalRead(address uint16) uint8 {
	return 0x0000
}

func (ppu *Ppu) internalWrite(address uint16, data uint8) {
}

func (ppu *Ppu) Read(address uint16) uint8 {
	switch address {
	case 0x2000:
		// TODO: implement ppuctrl
	case 0x2001:
		// TODO: implement ppumask
	case 0x2002:
		// TODO: implement ppustatus
	case 0x2003:
		// TODO: implement oamaddr
	case 0x2004:
		// TODO: implement oamdata
	case 0x2005:
		// TODO: implement ppuscroll
	case 0x2006:
		// TODO: implement ppuaddr
	case 0x2007:
		// TODO: implement ppudata
	}
	return 0x0000
}

func (ppu *Ppu) Write(address uint16, data uint8) {
	switch address {
	case 0x2000:
		// TODO: implement ppuctrl
	case 0x2001:
		// TODO: implement ppumask
	case 0x2002:
		// TODO: implement ppustatus
	case 0x2003:
		// TODO: implement oamaddr
	case 0x2004:
		// TODO: implement oamdata
	case 0x2005:
		// TODO: implement ppuscroll
	case 0x2006:
		// TODO: implement ppuaddr
	case 0x2007:
		// TODO: implement ppudata
	}
}
