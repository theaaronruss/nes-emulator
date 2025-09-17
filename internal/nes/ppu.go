package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240
	cycles      int     = 341
	scanLines   int     = 262
)

const (
	nametableDataSize int = 8192
	paletteDataSize   int = 32

	nametableAddr    uint16 = 0x2000
	nametableMemSize uint16 = 4096
	paletteAddr      uint16 = 0x3F00
	paletteMemSize   uint16 = 256
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

	currScanLine  int
	currCycle     int
	dataBuffer    uint8
	nametableData [nametableDataSize]uint8
	paletteData   [paletteDataSize]uint8
	v             uint16 // current vram address
	t             uint16 // temp address
	x             uint8  // fine x scroll
	w             bool   // write latch
}

func NewPpu() *Ppu {
	return &Ppu{
		FrameBuffer:   make([]uint8, 4*int(FrameWidth)*int(FrameHeight)),
		FrameComplete: false,
		currScanLine:  0,
		currCycle:     0,
		dataBuffer:    0,
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
	if address >= nametableAddr && address < nametableAddr+nametableMemSize {
		nametableAddress := address - nametableAddr
		return ppu.nametableData[nametableAddress]
	} else if address >= paletteAddr && address < paletteAddr+paletteMemSize {
		paletteAddress := (address - paletteAddr) % uint16(paletteDataSize)
		return ppu.paletteData[paletteAddress]
	}
	return 0x00
}

func (ppu *Ppu) internalWrite(address uint16, data uint8) {
	if address >= 0x2000 && address < 0x3000 {
		ppu.nametableData[address-0x2000] = data
	} else if address >= paletteAddr && address < paletteAddr+paletteMemSize {
		paletteAddress := (address - paletteAddr) % uint16(paletteDataSize)
		ppu.paletteData[paletteAddress] = data
	}
}

func (ppu *Ppu) Read(address uint16) uint8 {
	switch address {
	case ppuStatus:
		// TODO: implement ppustatus
	case oamData:
		// TODO: implement oamdata
	case ppuData:
		data := ppu.dataBuffer
		ppu.dataBuffer = ppu.internalRead(ppu.v)
		ppu.v++ // TODO: increment by 1 or 32 based on ppuCtrl
		return data
	}
	return 0x0000
}

func (ppu *Ppu) Write(address uint16, data uint8) {
	switch address {
	case ppuCtrl:
		// TODO: implement ppuctrl
	case ppuMask:
		// TODO: implement ppumask
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
		ppu.internalWrite(ppu.v, data)
		ppu.v++ // TODO: increment by 1 or 32 based on ppuCtrl
	}
}
