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

type SysBus struct {
	cpuRam    [cpuRamSize]uint8
	cartridge *Cartridge
	cpu       *Cpu
	ppu       *Ppu
}

func NewSysBus() *SysBus {
	return &SysBus{}
}

func (bus *SysBus) InsertCartridge(cartridge *Cartridge) {
	bus.cartridge = cartridge
}

func (bus *SysBus) SetCpu(cpu *Cpu) {
	bus.cpu = cpu
}

func (bus *SysBus) SetPpu(ppu *Ppu) {
	bus.ppu = ppu
}

func (bus *SysBus) Read(address uint16) uint8 {
	switch {
	case address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize:
		return bus.cpuRam[address]
	case bus.ppu != nil && address == PpuStatus:
		return bus.ppu.ReadPpuStatus()
	case bus.ppu != nil && address == OamData:
		return bus.ppu.ReadOamData()
	case bus.ppu != nil && address == PpuData:
		return bus.ppu.ReadPpuData()
	case address >= cartridgeAddr && bus.cartridge != nil:
		return bus.cartridge.MustReadProgramData(address)
	}
	return 0x00
}

func (bus *SysBus) Write(address uint16, data uint8) {
	switch {
	case address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize:
		bus.cpuRam[address] = data
	case bus.ppu != nil && address == PpuCtrl:
		bus.ppu.WritePpuCtrl(data)
	case bus.ppu != nil && address == PpuMask:
		bus.ppu.WritePpuMask(data)
	case bus.ppu != nil && address == OamAddr:
		bus.ppu.WriteOamAddr(data)
	case bus.ppu != nil && address == OamData:
		bus.ppu.WriteOamData(data)
	case bus.ppu != nil && address == PpuScroll:
		bus.ppu.WritePpuScroll(data)
	case bus.ppu != nil && address == PpuAddr:
		bus.ppu.WritePpuAddr(data)
	case bus.ppu != nil && address == PpuData:
		bus.ppu.WritePpuData(data)
	}
}
