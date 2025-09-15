package nes

const (
	cpuRamSize int = 2048
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
	if address < 0x0800 {
		return bus.cpuRam[address]
	} else if address >= 0x2000 && address < 0x2008 && bus.ppu != nil {
		return bus.ppu.Read(address - 0x2000)
	} else if address >= 0x8000 && bus.cartridge != nil {
		return bus.cartridge.MustRead(address)
	}
	return 0x00
}

func (bus *SysBus) Write(address uint16, data uint8) {
	if address >= 0x2000 && address < 0x2008 && bus.ppu != nil {
		bus.ppu.Write(address-0x2000, data)
	} else if address < 0x0800 {
		bus.cpuRam[address] = data
	}
}
