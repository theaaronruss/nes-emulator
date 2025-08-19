package cpu

const (
	initialProgCounter uint16 = 0xFFFC
	stackSize          uint8  = 0xFF
	stackBase          uint16 = 0x0100
	initialStatus      uint8  = 0b00000100
	memorySize         uint16 = 0xFFFF
	irqVector          uint16 = 0xFFFE
	negativeFlagMask   uint8  = 0b10000000
	overflowFlagMask   uint8  = 0b01000000
	unnamedFlagMask    uint8  = 0b00100000
	breakFlagMask      uint8  = 0b00010000
	decimalFlagMask    uint8  = 0b00001000
	interruptFlagMask  uint8  = 0b00000100
	zeroFlagMask       uint8  = 0b00000010
	carryFlagMask      uint8  = 0b00000001
)

type Cpu struct {
	accumulator  uint8
	xIndex       uint8
	yIndex       uint8
	progCounter  uint16
	stackPointer uint8
	status       uint8
	memory       [memorySize]uint8
}

func NewCpu() *Cpu {
	return &Cpu{
		accumulator:  0,
		xIndex:       0,
		yIndex:       0,
		progCounter:  initialProgCounter,
		stackPointer: stackSize,
		status:       initialStatus,
		memory:       [memorySize]uint8{},
	}
}

func (c *Cpu) Cycle() {
	opcode := c.memory[c.progCounter]
	switch opcode {
	case 0x00:
		c.forceBreak()
	case 0x01:
		arg := c.memory[c.progCounter+1]
		address := c.getIndirectXAddress(arg)
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x05:
		arg := c.memory[c.progCounter+1]
		c.bitwiseOr(arg)
	case 0x06:
		c.arithmeticShiftLeftZeroPage()
	}
}

func (c *Cpu) stackPush(data uint8) {
	if c.stackPointer == 0x00 {
		panic("stack overflow")
	}
	address := stackBase + uint16(c.stackPointer)
	c.memory[address] = data
	c.stackPointer--
}

func (c *Cpu) stackPop() uint8 {
	if c.stackPointer == 0xFF {
		panic("stack underflow")
	}
	c.stackPointer++
	address := stackBase + uint16(c.stackPointer)
	return c.memory[address]
}

func (c *Cpu) getIndirectXAddress(arg uint8) uint16 {
	zeroPageOffset := uint16(arg + c.xIndex)
	zeroPageAddr := uint8(zeroPageOffset % 256)
	lowByte := c.memory[zeroPageAddr]
	highByte := uint16(c.memory[zeroPageAddr+1])
	address := highByte<<8 | uint16(lowByte)
	return address
}

func (c *Cpu) forceBreak() {
	c.progCounter += 2
	pcByte1 := uint8(c.progCounter & 0xFF00 >> 8)
	pcByte2 := uint8(c.progCounter & 0x00FF)
	c.stackPush(pcByte1)
	c.stackPush(pcByte2)
	c.stackPush(c.status | unnamedFlagMask | breakFlagMask)
	c.status |= interruptFlagMask
	c.progCounter = irqVector
}

func (c *Cpu) bitwiseOr(value uint8) {
	c.accumulator |= value
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	}
}

func (c *Cpu) arithmeticShiftLeftZeroPage() {
	address := c.memory[c.progCounter+1]
	value := c.memory[address]
	if value&0b10000000 > 0 {
		c.status |= carryFlagMask
	}
	value <<= 1
	c.memory[address] = value
}
