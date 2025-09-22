package nes

// memory mapping
const (
	cpuRamStartAddr    uint16 = 0x0000
	cpuRamEndAddr      uint16 = 0x07FF
	cpuRamSize         uint16 = cpuRamEndAddr - cpuRamStartAddr + 1
	cartridgeStartAddr uint16 = 0x8000
	cartridgeEndAddr   uint16 = 0xFFFF
)

// ppu registers
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

type System struct {
	cpu       *cpu
	ppu       *ppu
	cpuRam    [cpuRamSize]uint8
	cartridge *Cartridge

	ppuClocks int
}

func NewSystem(cartridge *Cartridge) *System {
	sys := &System{
		cartridge: cartridge,
	}
	sys.ppu = NewPpu(sys)
	sys.cpu = NewCpu(sys)
	return sys
}

func (sys *System) FrameBuffer() []uint8 {
	return sys.ppu.frameBuffer
}

func (sys *System) ClockFrame() {
	for range clocksPerFrame {
		sys.ppu.Clock()
		sys.ppuClocks++
		if sys.ppuClocks >= 3 {
			sys.ppuClocks = 0
			sys.cpu.Clock()
		}
	}
}

func (sys *System) read(addr uint16) uint8 {
	switch {
	case addr <= 0x07FF:
		return sys.cpuRam[addr]
	case addr == ppuStatus:
		return sys.ppu.readPpuStatus()
	case addr == oamData:
		return sys.ppu.readOamData()
	case addr == ppuData:
		return sys.ppu.readPpuData()
	case sys.cartridge != nil && addr >= 0x8000:
		return sys.cartridge.ReadProgramData(addr)
	default:
		return 0
	}
}

func (sys *System) write(addr uint16, data uint8) {
	switch {
	case addr >= cpuRamStartAddr && addr <= cpuRamEndAddr:
		sys.cpuRam[addr] = data
	case addr == ppuCtrl:
		sys.ppu.writePpuCtrl(data)
	case addr == ppuMask:
		sys.ppu.writePpuMask(data)
	case addr == oamAddr:
		sys.ppu.writeOamAddr(data)
	case addr == oamData:
		sys.ppu.writeOamData(data)
	case addr == ppuScroll:
		sys.ppu.writePpuScroll(data)
	case addr == ppuAddr:
		sys.ppu.writePpuAddr(data)
	case addr == ppuData:
		sys.ppu.writePpuData(data)
	}
}
