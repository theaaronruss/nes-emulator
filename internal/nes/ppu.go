package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240

	cycles            int = 341
	scanLines         int = 262
	nameTableDataSize int = 8192
	paletteDataSize   int = 32
)

// memory mapping
const (
	patternTableAddr    uint16 = 0x0000
	patternTableMemSize uint16 = 8192
	nameTableAddr       uint16 = 0x2000
	nameTableMemSize    uint16 = 4096
	paletteAddr         uint16 = 0x3F00
	paletteMemSize      uint16 = 256
)

// public register addresses
const (
	PpuCtrl   uint16 = 0x2000
	PpuMask   uint16 = 0x2001
	PpuStatus uint16 = 0x2002
	OamAddr   uint16 = 0x2003
	OamData   uint16 = 0x2004
	PpuScroll uint16 = 0x2005
	PpuAddr   uint16 = 0x2006
	PpuData   uint16 = 0x2007
)

// ppuctrl masks
const (
	vblankNmiEnableMask    uint8 = 0x80
	backPatternTableMask   uint8 = 0x10
	spritePatternTableMask uint8 = 0x08
	vramIncrementMask      uint8 = 0x04
	baseNameTableAddrMask  uint8 = 0x03
)

// ppustatus masks
const (
	vblankMask uint8 = 0x80
)

type Ppu struct {
	FrameBuffer   []uint8
	FrameComplete bool

	currScanLine int
	currCycle    int

	bus           *SysBus
	dataBuffer    uint8
	nametableData [nameTableDataSize]uint8
	paletteData   [paletteDataSize]uint8

	vblankFlag         bool
	vblankNmiEnable    bool
	backPatternTable   bool
	spritePatternTable bool
	vramIncrement      bool
	baseNameTableAddr  uint8

	v uint16
	t uint16
	x uint8
	w bool
}

func NewPpu(bus *SysBus) *Ppu {
	return &Ppu{
		FrameBuffer: make([]uint8, 4*int(FrameWidth)*int(FrameHeight)),
		bus:         bus,
	}
}

func (ppu *Ppu) Clock() {
	ppu.currCycle++

	if ppu.currScanLine == 241 && ppu.currCycle == 1 {
		ppu.vblankFlag = true
	} else if ppu.currScanLine == 261 && ppu.currCycle == 1 {
		ppu.vblankFlag = false
	}

	if ppu.currCycle == cycles {
		ppu.currScanLine++
		ppu.currCycle = 0
		if ppu.currScanLine == scanLines {
			ppu.currScanLine = 0
			ppu.currCycle = 0
		}
	}
}

func (ppu *Ppu) Read(address uint16) uint8 {
	switch address {
	case PpuStatus:
		var status uint8
		if ppu.vblankFlag {
			status |= vblankMask
		}
		return status
	case OamData:
		// TODO: implement oamdata
	case PpuData:
		data := ppu.dataBuffer
		ppu.dataBuffer = ppu.internalRead(ppu.v)
		ppu.v++ // TODO: increment by 1 or 32 based on ppuctrl
		return data
	}
	return 0x0000
}

func (ppu *Ppu) Write(address uint16, data uint8) {
	switch address {
	case PpuCtrl:
		ppu.control(data)
	case PpuMask:
		// TODO: implement ppumask
	case OamAddr:
		// TODO: implement oamaddr
	case OamData:
		// TODO: implement oamdata
	case PpuScroll:
		// TODO: implement ppuscroll
	case PpuAddr:
		if !ppu.w {
			ppu.t = uint16(data) << 8
			ppu.t &= 0xBFFF // clear bit 14
		} else {
			ppu.t |= uint16(data)
			ppu.v = ppu.t
		}
		ppu.w = !ppu.w
	case PpuData:
		ppu.internalWrite(ppu.v, data)
		if !ppu.vramIncrement {
			ppu.v++
		} else {
			ppu.v += 32
		}
	}
}

func (ppu *Ppu) internalRead(address uint16) uint8 {
	if address >= patternTableAddr && address < patternTableAddr+patternTableMemSize {
		cartridgeAddress := address - patternTableAddr
		return ppu.bus.cartridge.MustReadProgramData(cartridgeAddress)
	} else if address >= nameTableAddr && address < nameTableAddr+nameTableMemSize {
		nametableAddress := address - nameTableAddr
		return ppu.nametableData[nametableAddress]
	} else if address >= paletteAddr && address < paletteAddr+paletteMemSize {
		paletteAddress := (address - paletteAddr) % uint16(paletteDataSize)
		return ppu.paletteData[paletteAddress]
	}
	return 0x00
}

func (ppu *Ppu) internalWrite(address uint16, data uint8) {
	if address >= nameTableAddr && address < nameTableAddr+nameTableMemSize {
		nametableAddress := address - nameTableAddr
		ppu.nametableData[nametableAddress] = data
	} else if address >= paletteAddr && address < paletteAddr+paletteMemSize {
		paletteAddress := (address - paletteAddr) % uint16(paletteDataSize)
		ppu.paletteData[paletteAddress] = data
	}
}

func (ppu *Ppu) control(data uint8) {
	if data&vblankNmiEnableMask > 0 {
		ppu.vblankNmiEnable = true
	} else {
		ppu.vblankNmiEnable = false
	}

	if data&backPatternTableMask > 0 {
		ppu.backPatternTable = true
	} else {
		ppu.backPatternTable = false
	}

	if data&spritePatternTableMask > 0 {
		ppu.spritePatternTable = true
	} else {
		ppu.spritePatternTable = false
	}

	if data&vramIncrementMask > 0 {
		ppu.vramIncrement = true
	} else {
		ppu.vramIncrement = false
	}

	ppu.baseNameTableAddr = data & baseNameTableAddrMask
}
