package cpu

const (
	initialProgCounter uint16 = 0x0800
	stackSize          uint8  = 0xFF
	stackBase          uint16 = 0x0100
	initialStatus      uint8  = 0b00000100
	memorySize         uint16 = 0xFFFF
	pageSize           int    = 256
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
	cycleDelay   int
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
		cycleDelay:   0,
	}
}

func (c *Cpu) Cycle() {
	if c.cycleDelay > 0 {
		c.cycleDelay--
		return
	}
	opcode := c.memory[c.progCounter]
	switch opcode {
	case 0x00:
		c.forceBreak()
		c.cycleDelay = 7
	case 0x01:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x05:
		arg := c.memory[c.progCounter+1]
		c.bitwiseOr(arg)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x06:
		arg := c.memory[c.progCounter+1]
		c.arithmeticShiftLeftMemory(uint16(arg))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0x08:
		c.stackPush(c.status | unnamedFlagMask | breakFlagMask)
		c.progCounter++
		c.cycleDelay = 3
	case 0x09:
		arg := c.memory[c.progCounter+1]
		c.bitwiseOr(arg)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0x0A:
		c.arithmeticShiftLeftAccumulator()
		c.progCounter++
		c.cycleDelay = 2
	case 0x0D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x0E:
		address := c.getAbsoluteAddress()
		c.arithmeticShiftLeftMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0x10:
		c.branchIfPlus()
	case 0x11:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0x15:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x16:
		address := c.getZeroPageXAddress()
		c.arithmeticShiftLeftMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x18:
		c.status &= ^carryFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0x19:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x1D:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x1E:
		address, _ := c.getAbsoluteXAddress()
		c.arithmeticShiftLeftMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
	case 0x20:
		c.jumpToSubroutine()
		c.progCounter += 3
		c.cycleDelay = 6
	case 0x21:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x24:
		address := c.memory[c.progCounter+1]
		c.bitTest(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x25:
		address := c.memory[c.progCounter+1]
		c.bitwiseAnd(address)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x26:
		address := c.memory[c.progCounter+1]
		c.rotateLeftMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0x28:
		c.popProcessorStatus()
		c.progCounter++
		c.cycleDelay = 4
	case 0x29:
		value := c.memory[c.progCounter+1]
		c.bitwiseAnd(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0x2A:
		c.rotateLeftAccumulator()
		c.progCounter++
		c.cycleDelay = 2
	case 0x2C:
		address := c.getAbsoluteAddress()
		c.bitTest(address)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x2D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x2E:
		address := c.getAbsoluteAddress()
		c.rotateLeftMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0x30:
		c.branchIfMinus()
	case 0x31:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0x35:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x36:
		address := c.getZeroPageXAddress()
		c.rotateLeftMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x38:
		c.status |= carryFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0x39:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x3D:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseAnd(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x3E:
		address, _ := c.getAbsoluteXAddress()
		c.rotateLeftMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
	case 0x40:
		c.returnFromInterrupt()
		c.cycleDelay = 6
	case 0x41:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x45:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x46:
		address := c.memory[c.progCounter+1]
		c.logicalShiftRightMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0x48:
		c.stackPush(c.accumulator)
		c.progCounter++
		c.cycleDelay = 3
	case 0x49:
		arg := c.memory[c.progCounter+1]
		c.bitwiseXor(arg)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0x4A:
		c.logicalShiftRightAccumulator()
		c.progCounter++
		c.cycleDelay = 2
	case 0x4C:
		address := c.getAbsoluteAddress()
		c.progCounter = address
		c.cycleDelay = 3
	case 0x4D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.bitwiseOr(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x4E:
		address := c.getAbsoluteAddress()
		c.logicalShiftRightMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0x50:
		c.branchIfOverflowClear()
	case 0x51:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0x55:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x56:
		address := c.getZeroPageXAddress()
		c.logicalShiftRightMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x58:
		c.status &= ^interruptFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0x59:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x5D:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.bitwiseXor(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x5E:
		address, _ := c.getAbsoluteXAddress()
		c.logicalShiftRightMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
	case 0x60:
		addrLow := c.stackPop()
		addrHigh := uint16(c.stackPop())
		address := addrHigh<<8 | uint16(addrLow)
		c.progCounter = address
		c.cycleDelay = 6
	case 0x61:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x65:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x66:
		address := c.memory[c.progCounter+1]
		c.rotateRightMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0x68:
		c.pullA()
		c.progCounter++
		c.cycleDelay = 4
	case 0x69:
		value := c.memory[c.progCounter+1]
		c.addWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0x6A:
		c.rotateRightAccumulator()
		c.progCounter++
		c.cycleDelay = 2
	case 0x6C:
		address := c.getIndirectAddress()
		c.progCounter = address
		c.cycleDelay = 5
	case 0x6D:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x6E:
		address := c.getAbsoluteAddress()
		c.rotateRightMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0x70:
		c.branchIfOverflowSet()
	case 0x71:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0x75:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x76:
		address := c.getZeroPageXAddress()
		c.rotateRightMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x78:
		c.status |= interruptFlagMask
		c.progCounter += 1
		c.cycleDelay = 2
	case 0x79:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x7D:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.addWithCarry(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0x7E:
		address, _ := c.getAbsoluteXAddress()
		c.rotateRightMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
	case 0x81:
		address := c.getIndirectXAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x84:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.yIndex
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x85:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.accumulator
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x86:
		address := c.memory[c.progCounter+1]
		c.memory[address] = c.xIndex
		c.progCounter += 2
		c.cycleDelay = 3
	case 0x88:
		c.decrementY()
		c.progCounter += 1
		c.cycleDelay = 2
	case 0x8A:
		c.transferXToA()
		c.progCounter += 1
		c.cycleDelay = 2
	case 0x8C:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.yIndex
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x8D:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x8E:
		address := c.getAbsoluteAddress()
		c.memory[address] = c.xIndex
		c.progCounter += 3
		c.cycleDelay = 4
	case 0x90:
		c.branchIfCarryClear()
	case 0x91:
		address, _ := c.getIndirectYAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 2
		c.cycleDelay = 6
	case 0x94:
		address := c.getZeroPageXAddress()
		c.memory[address] = c.yIndex
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x95:
		address := c.getZeroPageXAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x96:
		address := c.getZeroPageYAddress()
		c.memory[address] = c.xIndex
		c.progCounter += 2
		c.cycleDelay = 4
	case 0x98:
		c.transferYToA()
		c.progCounter += 1
		c.cycleDelay = 2
	case 0x99:
		address, _ := c.getAbsoluteYAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 3
		c.cycleDelay = 5
	case 0x9A:
		c.stackPointer = c.xIndex
		c.progCounter++
		c.cycleDelay = 2
	case 0x9D:
		address, _ := c.getAbsoluteXAddress()
		c.memory[address] = c.accumulator
		c.progCounter += 3
		c.cycleDelay = 5
	case 0xA0:
		value := c.memory[c.progCounter+1]
		c.loadY(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xA1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0xA2:
		value := c.memory[c.progCounter+1]
		c.loadX(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xA4:
		address := c.memory[c.progCounter+1]
		c.yIndex = c.memory[address]
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xA5:
		address := c.memory[c.progCounter+1]
		c.accumulator = c.memory[address]
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xA6:
		address := c.memory[c.progCounter+1]
		c.xIndex = c.memory[address]
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xA8:
		c.transferAToY()
		c.progCounter++
		c.cycleDelay = 2
	case 0xA9:
		value := c.memory[c.progCounter+1]
		c.loadA(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xAA:
		c.transferAToX()
		c.progCounter++
		c.cycleDelay = 2
	case 0xAC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadY(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xAD:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xAE:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.loadX(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xB0:
		c.branchIfCarrySet()
	case 0xB1:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0xB4:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.loadY(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0xB5:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0xB6:
		address := c.getZeroPageYAddress()
		value := c.memory[address]
		c.loadX(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0xB8:
		c.status &= ^overflowFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0xB9:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xBA:
		c.transferStackPointerToX()
		c.progCounter++
		c.cycleDelay = 2
	case 0xBC:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.loadY(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xBD:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.loadA(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xBE:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.loadX(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xC0:
		value := c.memory[c.progCounter+1]
		c.compareY(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xC1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0xC4:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareY(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xC5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xC6:
		address := c.memory[c.progCounter+1]
		c.decrementMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0xC8:
		c.incrementY()
		c.progCounter++
		c.cycleDelay = 2
	case 0xC9:
		value := c.memory[c.progCounter+1]
		c.compareA(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xCA:
		c.decrementX()
		c.progCounter++
		c.cycleDelay = 2
	case 0xCC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareY(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xCD:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xCE:
		address := c.getAbsoluteAddress()
		c.decrementMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0xD0:
		c.branchIfNotEqual()
	case 0xD1:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0xD5:
		address := c.getZeroPageXAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0xD6:
		address := c.getZeroPageXAddress()
		c.decrementMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0xD8:
		c.status &= ^decimalFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0xD9:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xDD:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.compareA(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xDE:
		address, _ := c.getAbsoluteXAddress()
		c.decrementMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
	case 0xE0:
		value := c.memory[c.progCounter+1]
		c.compareX(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xE1:
		address := c.getIndirectXAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 6
	case 0xE4:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.compareX(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xE5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 3
	case 0xE6:
		address := c.memory[c.progCounter+1]
		c.incrementMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 5
	case 0xE8:
		c.incrementX()
		c.progCounter++
		c.cycleDelay = 2
	case 0xE9:
		value := c.memory[c.progCounter+1]
		c.subtractWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 2
	case 0xEA:
		c.progCounter++
		c.cycleDelay = 2
	case 0xEC:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.compareX(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xED:
		address := c.getAbsoluteAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 3
		c.cycleDelay = 4
	case 0xEE:
		address := c.getAbsoluteAddress()
		c.incrementMemory(address)
		c.progCounter += 3
		c.cycleDelay = 6
	case 0xF0:
		c.branchIfEqual()
	case 0xF1:
		address, pageCrossed := c.getIndirectYAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 2
		if pageCrossed {
			c.cycleDelay = 6
		} else {
			c.cycleDelay = 5
		}
	case 0xF5:
		address := c.memory[c.progCounter+1]
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 2
		c.cycleDelay = 4
	case 0xF6:
		address := c.getZeroPageXAddress()
		c.incrementMemory(uint16(address))
		c.progCounter += 2
		c.cycleDelay = 6
	case 0xF8:
		c.status |= decimalFlagMask
		c.progCounter++
		c.cycleDelay = 2
	case 0xF9:
		address, pageCrossed := c.getAbsoluteYAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xFD:
		address, pageCrossed := c.getAbsoluteXAddress()
		value := c.memory[address]
		c.subtractWithCarry(value)
		c.progCounter += 3
		if pageCrossed {
			c.cycleDelay = 5
		} else {
			c.cycleDelay = 4
		}
	case 0xFE:
		address, _ := c.getAbsoluteXAddress()
		c.incrementMemory(address)
		c.progCounter += 3
		c.cycleDelay = 7
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

func (c *Cpu) getIndirectYAddress() (uint16, bool) {
	arg := c.memory[c.progCounter+1]
	addrLow := c.memory[arg]
	addrHigh := uint16(c.memory[arg+1]) << 8
	baseAddress := addrHigh | uint16(addrLow)
	address := baseAddress + uint16(c.yIndex)
	pageCrossed := baseAddress/uint16(pageSize) != address/uint16(pageSize)
	return address, pageCrossed
}

func (c *Cpu) getAbsoluteAddress() uint16 {
	addrLow := c.memory[c.progCounter+1]
	addrHigh := c.memory[c.progCounter+2]
	return uint16(addrHigh)<<8 | uint16(addrLow)
}

func (c *Cpu) getAbsoluteXAddress() (uint16, bool) {
	baseAddress := c.getAbsoluteAddress()
	address := baseAddress + uint16(c.xIndex)
	pageCrossed := baseAddress/uint16(pageSize) != address/uint16(pageSize)
	return address, pageCrossed
}

func (c *Cpu) getAbsoluteYAddress() (uint16, bool) {
	baseAddress := c.getAbsoluteAddress()
	address := baseAddress + uint16(c.yIndex)
	pageCrossed := baseAddress/uint16(pageSize) != address/uint16(pageSize)
	return address, pageCrossed
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
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfMinus() {
	if c.status&negativeFlagMask == 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfOverflowClear() {
	if c.status&overflowFlagMask != 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfOverflowSet() {
	if c.status&overflowFlagMask == 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfCarryClear() {
	if c.status&carryFlagMask != 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfCarrySet() {
	if c.status&carryFlagMask == 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfEqual() {
	if c.status&zeroFlagMask != 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
}

func (c *Cpu) branchIfNotEqual() {
	if c.status&zeroFlagMask == 0 {
		c.cycleDelay = 2
		return
	}
	offset := c.memory[c.progCounter+1]
	address := c.progCounter + uint16(2+offset)
	if address/uint16(pageSize) == c.progCounter/uint16(pageSize) {
		c.cycleDelay = 3
	} else {
		c.cycleDelay = 4
	}
	c.progCounter = address
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
	oldAccumulator := c.accumulator
	diff := int16(c.accumulator) - int16(value)
	if c.status&carryFlagMask == 0 {
		diff--
	}
	if diff < 0x00 {
		c.status &= ^carryFlagMask
	} else {
		c.status |= carryFlagMask
	}
	c.accumulator = uint8(diff)
	if c.accumulator == 0 {
		c.status |= zeroFlagMask
	} else {
		c.status &= ^zeroFlagMask
	}
	if (c.accumulator^oldAccumulator)&(c.accumulator^^value)&0b10000000 > 0 {
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
