package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240
	cycles      int     = 341
	scanLines   int     = 262
)

// public registers
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

type Ppu struct {
	FrameBuffer   []uint8
	FrameComplete bool

	currScanLine int
	currCycle    int
	v            uint16
	t            uint16
	x            uint8
	w            bool
}

func NewPpu() *Ppu {
	return &Ppu{
		FrameBuffer:   make([]uint8, 4*int(FrameWidth)*int(FrameHeight)),
		FrameComplete: false,
		currScanLine:  0,
		currCycle:     0,
		v:             0, t: 0, x: 0, w: false,
	}
}

func (ppu *Ppu) Clock() {
	// TODO: implement clock cycle
	ppu.currCycle++
	if ppu.currCycle == cycles {
		ppu.currScanLine++
		ppu.currCycle = 0
		if ppu.currScanLine == scanLines {
			ppu.currScanLine = 0
			ppu.currCycle = 0
		}
	}
}

func (ppu *Ppu) internalRead(address uint16) uint8 {
	return 0x00
}

func (ppu *Ppu) internalWrite(address uint16, data uint8) {
}

func (ppu *Ppu) Read(address uint16) uint8 {
	switch address {
	case ppuCtrl:
		// TODO: implement ppuctrl
	case ppuMask:
		// TODO: implement ppumask
	case ppuStatus:
		// TODO: implement ppustatus
	case oamAddr:
		// TODO: implement oamaddr
	case oamData:
		// TODO: implement oamdata
	case ppuScroll:
		// TODO: implement ppuscroll
	case ppuAddr:
		// TODO: implement ppuaddr
	case ppuData:
		// TODO: implement ppudata
	}
	return 0x0000
}

func (ppu *Ppu) Write(address uint16, data uint8) {
	switch address {
	case ppuCtrl:
		// TODO: implement ppuctrl
	case ppuMask:
		// TODO: implement ppumask
	case ppuStatus:
		// TODO: implement ppustatus
	case oamAddr:
		// TODO: implement oamaddr
	case oamData:
		// TODO: implement oamdata
	case ppuScroll:
		// TODO: implement ppuscroll
	case ppuAddr:
		if !ppu.w {
			ppu.t = uint16(data) << 8
			ppu.t &= 0xBFFF // clear bit 14
		} else {
			ppu.t |= uint16(data)
			ppu.v = ppu.t
		}
		ppu.w = !ppu.w
	case ppuData:
		// TODO: implement ppudata
	}
}
