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
	case 0x19:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x1D:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x1E:
		address := c.getAbsoluteXAddress()
		c.arithmeticShiftLeftMemory(address)
	case 0x20:
		c.jumpToSubroutine()
	case 0x21:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x24:
		address := c.memory[c.progCounter+1]
		c.bitTest(uint16(address))
	case 0x25:
		address := c.memory[c.progCounter+1]
		c.bitwiseAnd(address)
	case 0x26:
		address := c.memory[c.progCounter+1]
		c.rotateLeftMemory(uint16(address))
	case 0x28:
		c.popProcessorStatus()
	case 0x29:
		value := c.memory[c.progCounter+1]
		c.bitwiseAnd(value)
	case 0x2A:
		c.rotateLeftAccumulator()
	case 0x2C:
		address := c.getAbsoluteAddress()
		c.bitTest(address)
	case 0x2D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x2E:
		address := c.getAbsoluteAddress()
		c.rotateLeftMemory(address)
	case 0x30:
		c.branchIfMinus()
	case 0x31:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x35:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x36:
		address := c.getZeroPageXAddress()
		c.rotateLeftMemory(uint16(address))
	case 0x38:
		c.status |= carryFlagMask
	case 0x39:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x3D:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
	case 0x3E:
		address := c.getAbsoluteXAddress()
		c.rotateLeftMemory(address)
	case 0x40:
		c.returnFromInterrupt()
	case 0x41:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
	case 0x45:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.bitwiseXor(value)
	case 0x46:
		address := c.memory[c.progCounter+1]
		c.logicalShiftRightMemory(uint16(address))
	case 0x48:
		c.stackPush(c.accumulator)
	case 0x49:
		arg := c.memory[c.progCounter+1]
		c.bitwiseXor(arg)
	case 0x4A:
		c.logicalShiftRightAccumulator()
	case 0x4C:
		address := c.getAbsoluteAddress()
		c.progCounter = address
	case 0x4D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
	case 0x4E:
		address := c.getAbsoluteAddress()
		c.logicalShiftRightMemory(address)
	case 0x50:
		c.branchIfOverflowClear()
	case 0x51:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
	case 0x55:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
	case 0x56:
		address := c.getZeroPageXAddress()
		c.logicalShiftRightMemory(uint16(address))
	case 0x58:
		c.status &= ^interruptFlagMask
	case 0x59:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
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

func (c *Cpu) getAbsoluteXAddress() uint16 {
	address := c.getAbsoluteAddress()
	address += uint16(c.xIndex)
	return address
}

func (c *Cpu) getAbsoluteYAddress() uint16 {
	address := c.getAbsoluteAddress()
	address += uint16(c.yIndex)
	return address
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
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) bitwiseAnd(value uint8) {
	c.accumulator &= value
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) bitwiseXor(value uint8) {
	c.accumulator ^= value
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) arithmeticShiftLeftAccumulator() {
	if c.accumulator&0b10000000 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	c.accumulator <<= 1
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) arithmeticShiftLeftMemory(address uint16) {
	value := c.memory[address]
	if value&0b10000000 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	value <<= 1
	if value == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if value&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
	c.memory[address] = value
}

func (c *Cpu) logicalShiftRightAccumulator() {
	if c.accumulator&0b00000001 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	c.accumulator >>= 1
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	c.status &= ^negativeFlagMask
}

func (c *Cpu) logicalShiftRightMemory(address uint16) {
	value := c.memory[address]
	if value&0b00000001 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	value >>= 1
	if value == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	c.status &= ^negativeFlagMask
	c.memory[address] = value
}

func (c *Cpu) branchIfPlus() {
	if c.status&negativeFlagMask > 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfMinus() {
	if c.status&negativeFlagMask == 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfOverflowClear() {
	if c.status&overflowFlagMask != 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) jumpToSubroutine() {
	addrLow := c.memory[c.progCounter+1]
	addrHigh := uint16(c.memory[c.progCounter+2]) << 8
	address := addrHigh | uint16(addrLow)
	c.progCounter += 2
	currProgCountLow := uint8(c.progCounter & 0x00FF)
	currProgCountHigh := uint8(c.progCounter & 0xFF00 >> 8)
	c.stackPush(currProgCountHigh)
	c.stackPush(currProgCountLow)
	c.progCounter = address
}

func (c *Cpu) bitTest(address uint16) {
	result := c.accumulator & c.memory[address]
	if result == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if result&0b01000000 > 0 {
		c.status |= overflowFlagMask
	} else {
		c.status &= ^overflowFlagMask
	}
	if result&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) rotateLeftMemory(address uint16) {
	value := c.memory[address]
	carryFlag := c.status&carryFlagMask > 0
	if value&0b10000000 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	value <<= 1
	if carryFlag {
		value |= 0b00000001
	}
	c.memory[address] = value
}

func (c *Cpu) rotateLeftAccumulator() {
	carryFlag := c.status&carryFlagMask > 0
	if c.accumulator&0b10000000 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	c.accumulator <<= 1
	if carryFlag {
		c.accumulator |= 0b00000001
	}
}

func (c *Cpu) popProcessorStatus() {
	flags := c.stackPop()
	flags &= negativeFlagMask | overflowFlagMask | decimalFlagMask |
		interruptFlagMask | zeroFlagMask | carryFlagMask
	c.status = flags
}

func (c *Cpu) returnFromInterrupt() {
	c.popProcessorStatus()
	addrLow := c.stackPop()
	addrHigh := c.stackPop()
	address := uint16(addrHigh)<<8 | uint16(addrLow)
	c.progCounter = address
}
