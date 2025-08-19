package cpu

const (
	initialProgCounter uint16 = 0x0800
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
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x05:
		arg := c.memory[c.progCounter+1]
		c.bitwiseOr(arg)
	case 0x06:
		arg := c.memory[c.progCounter+1]
		c.arithmeticShiftLeftMemory(uint16(arg))
	case 0x08:
		c.stackPush(c.status | unnamedFlagMask | breakFlagMask)
	case 0x09:
		arg := c.memory[c.progCounter+1]
		c.bitwiseOr(arg)
	case 0x0A:
		c.arithmeticShiftLeftAccumulator()
	case 0x0D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x0E:
		address := c.getAbsoluteAddress()
		c.arithmeticShiftLeftMemory(address)
	case 0x10:
		c.branchIfPlus()
	case 0x11:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x15:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x16:
		address := c.getZeroPageXAddress()
		c.arithmeticShiftLeftMemory(uint16(address))
	case 0x18:
		c.status &= ^carryFlagMask
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

func (c *Cpu) getIndirectXAddress() uint16 {
	arg := c.memory[c.progCounter+1]
	zeroPageOffset := uint16(arg + c.xIndex)
	zeroPageAddr := uint8(zeroPageOffset % 256)
	lowByte := c.memory[zeroPageAddr]
	highByte := uint16(c.memory[zeroPageAddr+1])
	address := highByte<<8 | uint16(lowByte)
	return address
}

func (c *Cpu) getIndirectYAddress() uint16 {
	arg := c.memory[c.progCounter+1]
	addrLow := c.memory[arg]
	addrHigh := uint16(c.memory[arg+1]) << 8
	address := addrHigh | uint16(addrLow)
	address += uint16(c.yIndex)
	return address
}

func (c *Cpu) getAbsoluteAddress() uint16 {
	addrLow := c.memory[c.progCounter+1]
	addrHigh := c.memory[c.progCounter+2]
	return uint16(addrHigh)<<8 | uint16(addrLow)
}

func (c *Cpu) getZeroPageXAddress() uint8 {
	arg := c.memory[c.progCounter+1]
	return arg + c.xIndex
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

func (c *Cpu) arithmeticShiftLeftAccumulator() {
	if c.accumulator&0b10000000 > 0 {
		c.status |= carryFlagMask
	}
	c.accumulator <<= 1
}

func (c *Cpu) arithmeticShiftLeftMemory(address uint16) {
	value := c.memory[address]
	if value&0b10000000 > 0 {
		c.status |= carryFlagMask
	}
	value <<= 1
	c.memory[address] = value
}

func (c *Cpu) branchIfPlus() {
	if c.status&negativeFlagMask > 0 {
		return
	}
	arg := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + int8(arg))
}
