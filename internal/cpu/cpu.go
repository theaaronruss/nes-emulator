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
	case 0x5D:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
	case 0x5E:
		address := c.getAbsoluteXAddress()
		c.logicalShiftRightMemory(address)
	case 0x60:
		addrLow := c.stackPop()
		addrHigh := uint16(c.stackPop())
		address := addrHigh<<8 | uint16(addrLow)
		c.progCounter = address
	case 0x61:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x65:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x66:
		address := c.memory[c.progCounter+1]
		c.rotateRightMemory(uint16(address))
	case 0x68:
		c.pullA()
	case 0x69:
		value := c.memory[c.progCounter+1]
		c.addWithCarry(value)
	case 0x6A:
		c.rotateRightAccumulator()
	case 0x6C:
		address := c.getIndirectAddress()
		c.progCounter = address
	case 0x6D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x6E:
		address := c.getAbsoluteAddress()
		c.rotateRightMemory(address)
	case 0x70:
		c.branchIfOverflowSet()
	case 0x71:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x75:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x76:
		address := c.getZeroPageXAddress()
		c.rotateRightMemory(uint16(address))
	case 0x78:
		c.status |= interruptFlagMask
	case 0x79:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x7D:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
	case 0x7E:
		address := c.getAbsoluteXAddress()
		c.rotateRightMemory(address)
	case 0x81:
		address := c.getIndirectXAddress()
		c.memory[address] = c.accumulator
	case 0x84:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.yIndex
	case 0x85:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.accumulator
	case 0x86:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.xIndex
	case 0x88:
		c.decrementY()
	case 0x8A:
		c.transferXToA()
	case 0x8C:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.yIndex
	case 0x8D:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.accumulator
	case 0x8E:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.xIndex
	case 0x90:
		c.branchIfCarryClear()
	case 0x91:
		address := c.getIndirectYAddress()
		c.memory[address] = c.accumulator
	case 0x94:
		address := c.getZeroPageXAddress()
		c.memory[address] = c.yIndex
	case 0x95:
		address := c.getZeroPageXAddress()
		c.memory[address] = c.accumulator
	case 0x96:
		address := c.getZeroPageYAddress()
		c.memory[address] = c.xIndex
	case 0x98:
		c.transferYToA()
	case 0x99:
		address := c.getAbsoluteYAddress()
		c.memory[address] = c.accumulator
	case 0x9A:
		c.stackPointer = c.xIndex
	case 0x9D:
		address := c.getAbsoluteXAddress()
		c.memory[address] = c.accumulator
	case 0xA0:
		value := c.memory[c.progCounter+1]
		c.loadY(value)
	case 0xA1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xA2:
		value := c.memory[c.progCounter+1]
		c.loadX(value)
	case 0xA4:
		address := c.memory[c.progCounter+1]
		c.yIndex = c.memory[address]
	case 0xA5:
		address := c.memory[c.progCounter+1]
		c.accumulator = c.memory[address]
	case 0xA6:
		address := c.memory[c.progCounter+1]
		c.xIndex = c.memory[address]
	case 0xA8:
		c.transferAToY()
	case 0xA9:
		value := c.memory[c.progCounter+1]
		c.loadA(value)
	case 0xAA:
		c.transferAToX()
	case 0xAC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadY(value)
	case 0xAD:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xAE:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadX(value)
	case 0xB0:
		c.branchIfCarrySet()
	case 0xB1:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xB4:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.loadY(value)
	case 0xB5:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xB6:
		address := c.getZeroPageYAddress()
		value := c.memory[address]
		c.loadX(value)
	case 0xB8:
		c.status &= ^overflowFlagMask
	case 0xB9:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xBA:
		c.transferStackPointerToX()
	case 0xBC:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.loadY(value)
	case 0xBD:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.loadA(value)
	case 0xBE:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.loadX(value)
	case 0xC0:
		value := c.memory[c.progCounter+1]
		c.compareY(value)
	case 0xC1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xC4:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareY(value)
	case 0xC5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareA(value)
	case 0xC6:
		address := c.memory[c.progCounter+1]
		c.decrementMemory(uint16(address))
	case 0xC8:
		c.incrementY()
	case 0xC9:
		value := c.memory[c.progCounter+1]
		c.compareA(value)
	case 0xCA:
		c.decrementX()
	case 0xCC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareY(value)
	case 0xCD:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xCE:
		address := c.getAbsoluteAddress()
		c.decrementMemory(address)
	case 0xD0:
		c.branchIfNotEqual()
	case 0xD1:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xD5:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xD6:
		address := c.getZeroPageXAddress()
		c.decrementMemory(uint16(address))
	case 0xD8:
		c.status &= ^decimalFlagMask
	case 0xD9:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xDD:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.compareA(value)
	case 0xDE:
		address := c.getAbsoluteXAddress()
		c.decrementMemory(address)
	case 0xE0:
		value := c.memory[c.progCounter+1]
		c.compareX(value)
	case 0xE1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xE4:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareX(value)
	case 0xE5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xE6:
		address := c.memory[c.progCounter+1]
		c.incrementMemory(uint16(address))
	case 0xE8:
		c.incrementX()
	case 0xE9:
		value := c.memory[c.progCounter+1]
		c.subtractWithCarry(value)
	case 0xEA:
		// no operation
	case 0xEC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareX(value)
	case 0xED:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xEE:
		address := c.getAbsoluteAddress()
		c.incrementMemory(address)
	case 0xF0:
		c.branchIfEqual()
	case 0xF1:
		address := c.getIndirectYAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xF5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xF6:
		address := c.getZeroPageXAddress()
		c.incrementMemory(uint16(address))
	case 0xF8:
		c.status |= decimalFlagMask
	case 0xF9:
		address := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xFD:
		address := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
	case 0xFE:
		address := c.getAbsoluteXAddress()
		c.incrementMemory(address)
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

func (c *Cpu) getIndirectAddress() uint16 {
	pointerLow := c.memory[c.progCounter+1]
	pointerHigh := uint16(c.memory[c.progCounter+2])
	pointer := pointerHigh<<8 | uint16(pointerLow)
	addrLow := c.memory[pointer]
	addrHigh := uint16(c.memory[pointer+1])
	return addrHigh<<8 | uint16(addrLow)
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

func (c *Cpu) getZeroPageYAddress() uint8 {
	arg := c.memory[c.progCounter+1]
	return arg + c.yIndex
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

func (c *Cpu) branchIfOverflowSet() {
	if c.status&overflowFlagMask == 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfCarryClear() {
	if c.status&carryFlagMask != 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfCarrySet() {
	if c.status&carryFlagMask == 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfEqual() {
	if c.status&zeroFlagMask != 0 {
		return
	}
	offset := c.memory[c.progCounter+1]
	c.progCounter += uint16(2 + offset)
}

func (c *Cpu) branchIfNotEqual() {
	if c.status&zeroFlagMask == 0 {
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

func (c *Cpu) rotateRightMemory(address uint16) {
	value := c.memory[address]
	carryFlag := c.status&carryFlagMask > 0
	if value&0b00000001 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	value >>= 1
	if carryFlag {
		value |= 0b10000000
	}
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

func (c *Cpu) rotateRightAccumulator() {
	carryFlag := c.status&carryFlagMask > 0
	if c.accumulator&0b00000001 > 0 {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	c.accumulator >>= 1
	if carryFlag {
		c.accumulator |= 0b10000001
	}
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

func (c *Cpu) addWithCarry(value uint8) {
	oldAccumulator := c.accumulator
	sum := uint16(c.accumulator) + uint16(value)
	if c.status&carryFlagMask > 0 {
		sum++
	}
	if sum > 0xFF {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	c.accumulator = uint8(sum)
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if (c.accumulator^oldAccumulator)&(c.accumulator^value)&0b10000000 > 0 {
		c.status |= overflowFlagMask
	} else {
		c.status &= ^overflowFlagMask
	}
	if c.accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) subtractWithCarry(value uint8) {
	// TODO: implement
}

func (c *Cpu) pullA() {
	accumulator := c.stackPop()
	if accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if accumulator&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) decrementX() {
	c.xIndex--
	if c.xIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.xIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) decrementY() {
	c.yIndex--
	if c.yIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.yIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) decrementMemory(address uint16) {
	value := c.memory[address]
	value--
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

func (c *Cpu) incrementX() {
	c.xIndex++
	if c.xIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.xIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) incrementY() {
	c.yIndex++
	if c.yIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.yIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) incrementMemory(address uint16) {
	value := c.memory[address]
	value++
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

func (c *Cpu) transferXToA() {
	c.accumulator = c.xIndex
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

func (c *Cpu) transferYToA() {
	c.accumulator = c.yIndex
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

func (c *Cpu) transferAToX() {
	c.xIndex = c.accumulator
	if c.xIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.xIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) transferAToY() {
	c.yIndex = c.accumulator
	if c.yIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.yIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) transferStackPointerToX() {
	c.xIndex = c.stackPointer
	if c.xIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.xIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) loadA(value uint8) {
	c.accumulator = value
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

func (c *Cpu) loadX(value uint8) {
	c.xIndex = value
	if c.xIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.xIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) loadY(value uint8) {
	c.yIndex = value
	if c.yIndex == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if c.yIndex&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) compareA(value uint8) {
	diff := c.accumulator - value
	if c.accumulator >= value {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	if diff == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if diff&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) compareX(value uint8) {
	diff := c.xIndex - value
	if c.xIndex >= value {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	if diff == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if diff&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}

func (c *Cpu) compareY(value uint8) {
	diff := c.yIndex - value
	if c.yIndex >= value {
		c.status |= carryFlagMask
	} else {
		c.status &= ^carryFlagMask
	}
	if diff == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if diff&0b10000000 > 0 {
		c.status |= negativeFlagMask
	} else {
		c.status &= ^negativeFlagMask
	}
}
