package bus

type Bus struct {
	ram [65535]uint8
}

func NewBus() *Bus {
	return &Bus{
		ram: [65535]uint8{},
	}
}

func (b *Bus) Read(address uint16) uint8 {
	return b.ram[address]
}

func (b *Bus) Write(address uint16, data uint8) {
	b.ram[address] = data
}
