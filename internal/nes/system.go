package nes

const (
	cpuRamSize int = 2048
)

// memory mapping
const (
	cpuRamAddr    uint16 = 0x0000
	cpuRamMemSize uint16 = 2048
	cartridgeAddr uint16 = 0x8000
)

type System struct {
	cpuRam    [cpuRamSize]uint8
	cartridge *Cartridge
	cpu       *Cpu
	ppu       *Ppu
}

func NewSystem() *System {
	sys := &System{}
	sys.ppu = NewPpu(sys)
	sys.cpu = NewCpu(sys)
	return sys
}

func (sys *System) InsertCartridge(cartridge *Cartridge) {
	sys.cartridge = cartridge
}

func (sys *System) FrameBuffer() []uint8 {
	return sys.ppu.FrameBuffer
}

func (sys *System) Read(address uint16) uint8 {
	switch {
	case address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize:
		return sys.cpuRam[address]
	case sys.ppu != nil && address == PpuStatus:
		return sys.ppu.ReadPpuStatus()
	case sys.ppu != nil && address == OamData:
		return sys.ppu.ReadOamData()
	case sys.ppu != nil && address == PpuData:
		return sys.ppu.ReadPpuData()
	case address >= cartridgeAddr && sys.cartridge != nil:
		return sys.cartridge.MustReadProgramData(address)
	}
	return 0x00
}

func (sys *System) Write(address uint16, data uint8) {
	switch {
	case address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize:
		sys.cpuRam[address] = data
	case sys.ppu != nil && address == PpuCtrl:
		sys.ppu.WritePpuCtrl(data)
	case sys.ppu != nil && address == PpuMask:
		sys.ppu.WritePpuMask(data)
	case sys.ppu != nil && address == OamAddr:
		sys.ppu.WriteOamAddr(data)
	case sys.ppu != nil && address == OamData:
		sys.ppu.WriteOamData(data)
	case sys.ppu != nil && address == PpuScroll:
		sys.ppu.WritePpuScroll(data)
	case sys.ppu != nil && address == PpuAddr:
		sys.ppu.WritePpuAddr(data)
	case sys.ppu != nil && address == PpuData:
		sys.ppu.WritePpuData(data)
	}
}
