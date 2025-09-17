package nes

const (
	cpuRamSize   int = 2048
	ppuRegisters int = 8
)

// memory mapping
const (
	cpuRamAddr    uint16 = 0x0000
	cpuRamMemSize uint16 = 2048
	ppuAddr       uint16 = 0x2000
	ppuMemSize    uint16 = 8192
	cartridgeAddr uint16 = 0x8000
)

type BusReadWriter interface {
	Read(uint16) uint8
	Write(uint16, uint8)
}

type SysBus struct {
	cpuRam    [cpuRamSize]uint8
	cartridge *Cartridge
	ppu       *Ppu
}

func NewSysBus() *SysBus {
	return &SysBus{}
}

func (bus *SysBus) InsertCartridge(cartridge *Cartridge) {
	bus.cartridge = cartridge
}

func (bus *SysBus) AddPpu(ppu *Ppu) {
	bus.ppu = ppu
}

func (bus *SysBus) Read(address uint16) uint8 {
	if address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize {
		return bus.cpuRam[address]
	} else if address >= ppuAddr && address < ppuAddr+ppuMemSize &&
		bus.ppu != nil {
		ppuAddress := (address - 0x2000) % uint16(ppuRegisters)
		return bus.ppu.Read(ppuAddress)
	} else if address >= cartridgeAddr && bus.cartridge != nil {
		return bus.cartridge.MustReadProgramData(address)
	}
	return 0x00
}

func (bus *SysBus) Write(address uint16, data uint8) {
	if address > cpuRamAddr && address < cpuRamAddr+cpuRamMemSize {
		bus.cpuRam[address] = data
	} else if address >= ppuAddr && address < ppuAddr+ppuMemSize &&
		bus.ppu != nil {
		ppuAddress := (address - 0x2000) % uint16(ppuRegisters)
		bus.ppu.Write(ppuAddress, data)
	}
}
