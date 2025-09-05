package processor

import "github.com/theaaronruss/nes-emulator/internal/sysbus"

const (
	resetVector         = 0xFFFC
	irqVector           = 0xFFFE
	stackBase           = 0x0100
	initialStackPointer = 0xFD
)

const (
	flagCarry      uint8 = 1 << iota
	flagZero             = 1 << iota
	flagIntDisable       = 1 << iota
	flagDecimal          = 1 << iota
	flagBreak            = 1 << iota
	flagUnused           = 1 << iota
	flagOverflow         = 1 << iota
	flagNegative         = 1 << iota
)

type Cpu struct {
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8

	cycleDelay int
}

func NewCpu() *Cpu {
	cpu := &Cpu{}
	cpu.Reset()
	return cpu
}

func (c *Cpu) Reset() {
	c.sp = initialStackPointer
	pcLow := sysbus.Read(resetVector)
	pcHigh := sysbus.Read(resetVector + 1)
	c.pc = uint16(pcHigh)<<8 | uint16(pcLow)
	c.setFlag(flagIntDisable)
	c.cycleDelay = 0
}

func (c *Cpu) ClockCycle() {
	if c.cycleDelay <= 0 {
		opcode := sysbus.Read(c.pc)
		instruction := opcodes[opcode]
		instruction.fn(c, &instruction)
		c.cycleDelay += instruction.cycles
		return
	}
	c.cycleDelay--
}

func (c *Cpu) setFlag(flag uint8) {
	c.status |= flag
}

func (c *Cpu) clearFlag(flag uint8) {
	c.status &= ^flag
}

func (c *Cpu) testFlag(flag uint8) bool {
	return c.status&flag > 0
}

func (c *Cpu) stackPush(data uint8) {
	address := stackBase + uint16(c.sp)
	sysbus.Write(address, data)
	c.sp--
}

func (c *Cpu) stackPop() uint8 {
	c.sp++
	address := stackBase + uint16(c.sp)
	return sysbus.Read(address)
}

func (c *Cpu) getAddress(addrMode addressMode) (uint16, bool) {
	switch addrMode {
	case addrModeZeroPageX:
		address := sysbus.Read(c.pc + 1)
		return uint16(address + c.x), false
	case addrModeZeroPageY:
		address := sysbus.Read(c.pc + 1)
		return uint16(address + c.y), false
	case addrModeAbsoluteX:
		baseAddress, _ := c.getAddress(addrModeAbsolute)
		address := baseAddress + uint16(c.x)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeAbsoluteY:
		baseAddress, _ := c.getAddress(addrModeAbsolute)
		address := baseAddress + uint16(c.y)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeIndexIndirX:
		zeroPageAddr := sysbus.Read(c.pc + 1)
		zeroPageAddr += c.x
		low := sysbus.Read(uint16(zeroPageAddr))
		high := sysbus.Read(uint16(zeroPageAddr) + 1)
		return uint16(high)<<8 | uint16(low), false
	case addrModeIndirIndexY:
		zeroPageAddr := sysbus.Read(c.pc + 1)
		low := sysbus.Read(uint16(zeroPageAddr))
		high := sysbus.Read(uint16(zeroPageAddr) + 1)
		baseAddress := uint16(high)<<8 | uint16(low)
		address := baseAddress + uint16(c.y)
		if baseAddress&0xFF00 == address&0xFF00 {
			return address, false
		} else {
			return address, true
		}
	case addrModeZeroPage:
		return uint16(sysbus.Read(c.pc + 1)), false
	case addrModeAbsolute:
		low := sysbus.Read(c.pc + 1)
		high := sysbus.Read(c.pc + 2)
		return uint16(high)<<8 | uint16(low), false
	case addrModeRelative:
		offset := int8(sysbus.Read(c.pc))
		return c.pc + uint16(offset), false
	case addrModeIndirect:
		address, _ := c.getAddress(addrModeAbsolute)
		low := sysbus.Read(address)
		high := sysbus.Read(address + 1)
		return uint16(high)<<8 | uint16(low), false
	}
	return 0x0000, false
}

func (c *Cpu) forceBreak(instr *instruction) {
	c.pc += 2
	oldPcLow := uint8(c.pc & 0x00FF)
	oldPcHigh := uint8(c.pc & 0xFF00 >> 8)
	c.stackPush(oldPcHigh)
	c.stackPush(oldPcLow)
	c.stackPush(c.status | flagUnused | flagBreak)
	newPcLow := sysbus.Read(irqVector)
	newPcHigh := sysbus.Read(irqVector + 1)
	newPc := uint16(newPcHigh)<<8 | uint16(newPcLow)
	c.pc = newPc
}

func (c *Cpu) returnFromInterrupt(instr *instruction) {
	flags := c.stackPop()
	low := c.stackPop()
	high := c.stackPop()
	c.status = flags & 0b11001111
	c.pc = uint16(high)<<8 | uint16(low)
}

func (c *Cpu) bitTest(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	value := sysbus.Read(address)
	result := c.a & value
	if result == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if value&0x40 > 0 {
		c.setFlag(flagOverflow)
	} else {
		c.clearFlag(flagOverflow)
	}
	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
}

func (c *Cpu) bitwiseOr(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.a |= value
	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) bitwiseXor(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.a ^= value
	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) bitwiseAnd(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.a &= value
	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) arithmeticShiftLeft(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = c.a
	} else {
		address, _ = c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}
	if value&0x80 > 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}
	value <<= 1
	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	if instr.addrMode == addrModeAccumulator {
		c.a = value
	} else {
		sysbus.Write(address, value)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) logicalShiftRight(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = c.a
	} else {
		address, _ = c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}
	if value&0x01 > 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}
	value >>= 1
	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	c.clearFlag(flagNegative)
	if instr.addrMode == addrModeAccumulator {
		c.a = value
	} else {
		sysbus.Write(address, value)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) rotateLeft(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = c.a
	} else {
		address, _ = c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if value&0x80 > 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	value <<= 1

	if c.testFlag(flagCarry) {
		value |= 0x01
	}

	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	if instr.addrMode == addrModeAccumulator {
		c.a = value
	} else {
		sysbus.Write(address, value)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) rotateRight(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = c.a
	} else {
		address, _ = c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if value&0x01 > 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	value >>= 1

	if c.testFlag(flagCarry) {
		value |= 0x80
	}

	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	if instr.addrMode == addrModeAccumulator {
		c.a = value
	} else {
		sysbus.Write(address, value)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) pushProcessorStatus(instr *instruction) {
	c.stackPush(c.status | flagUnused | flagBreak)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) pullProcessorStatus(instr *instruction) {
	flags := c.stackPop()
	c.status = flags & 0b11001111
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) pushA(instr *instruction) {
	c.stackPush(c.a)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) pullA(instr *instruction) {
	c.a = c.stackPop()

	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) storeA(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	sysbus.Write(address, c.a)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) loadA(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.a = value

	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) storeX(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	sysbus.Write(address, c.x)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) loadX(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.x = value

	if c.x == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.x&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) storeY(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	sysbus.Write(address, c.y)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) loadY(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}
	c.y = value

	if c.y == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.y&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferAToX(instr *instruction) {
	c.x = c.a

	if c.x == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.x&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferAToY(instr *instruction) {
	c.y = c.a

	if c.y == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.y&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferXToA(instr *instruction) {
	c.a = c.x

	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferYToA(instr *instruction) {
	c.a = c.x

	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.a&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferXToStackPointer(instr *instruction) {
	c.sp = c.x
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) transferStackPointerToX(instr *instruction) {
	c.x = c.sp
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) compareA(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}

	if c.a >= value {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	if c.a == value {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	result := c.a - value

	if result&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) compareX(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, _ := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if c.x >= value {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	if c.x == value {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	result := c.x - value

	if result&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) compareY(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, _ := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
	}

	if c.y >= value {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	if c.y == value {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	result := c.y - value

	if result&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) incrementY(instr *instruction) {
	c.y++

	if c.y == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.y&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) decrementY(instr *instruction) {
	c.y--

	if c.y == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.y&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) incrementX(instr *instruction) {
	c.x++

	if c.x == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.x&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) decrementX(instr *instruction) {
	c.x--

	if c.x == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if c.x&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) incrementMemory(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	value := sysbus.Read(address)
	value++
	sysbus.Write(address, value)

	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) decrementMemory(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	value := sysbus.Read(address)
	value--
	sysbus.Write(address, value)

	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if value&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.pc += uint16(instr.bytes)
}

func (c *Cpu) jump(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	c.pc = address
}

func (c *Cpu) jumpToSubroutine(instr *instruction) {
	address, _ := c.getAddress(instr.addrMode)
	c.pc += 2
	low := uint8(c.pc & 0x00FF)
	high := uint8((c.pc & 0xFF00) >> 8)
	c.stackPush(high)
	c.stackPush(low)
	c.pc = address
}

func (c *Cpu) returnFromSubroutine(instr *instruction) {
	low := c.stackPop()
	high := c.stackPop()
	address := uint16(high)<<8 | uint16(low)
	c.pc = address + 1
}

func (c *Cpu) branchIfPlus(instr *instruction) {
	if c.testFlag(flagNegative) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfMinus(instr *instruction) {
	if !c.testFlag(flagNegative) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfEqual(instr *instruction) {
	if !c.testFlag(flagZero) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfNotEqual(instr *instruction) {
	if c.testFlag(flagZero) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfCarrySet(instr *instruction) {
	if !c.testFlag(flagCarry) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfCarryClear(instr *instruction) {
	if c.testFlag(flagCarry) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfOverflowSet(instr *instruction) {
	if !c.testFlag(flagOverflow) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) branchIfOverflowClear(instr *instruction) {
	if c.testFlag(flagOverflow) {
		c.pc += uint16(instr.bytes)
		return
	}
	c.pc++
	address, pageCrossed := c.getAddress(instr.addrMode)
	c.pc = address + 1

	c.cycleDelay++
	if pageCrossed {
		c.cycleDelay++
	}
}

func (c *Cpu) setCarry(instr *instruction) {
	c.setFlag(flagCarry)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) clearCarry(instr *instruction) {
	c.clearFlag(flagCarry)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) clearOverflow(instr *instruction) {
	c.clearFlag(flagOverflow)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) setDecimal(instr *instruction) {
	c.setFlag(flagDecimal)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) clearDecimal(instr *instruction) {
	c.clearFlag(flagDecimal)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) setInterruptDisable(instr *instruction) {
	c.setFlag(flagIntDisable)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) clearInterruptDisable(instr *instruction) {
	c.clearFlag(flagIntDisable)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) addWithCarry(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}

	result := uint16(c.a) + uint16(value)
	if c.testFlag(flagCarry) {
		result++
	}

	if result > 255 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	if result == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if (uint8(result)^c.a)&(uint8(result)^value)&0x80 > 0 {
		c.setFlag(flagOverflow)
	} else {
		c.clearFlag(flagOverflow)
	}

	if result&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.a = uint8(result)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) subtractWithCarry(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = sysbus.Read(c.pc + 1)
	} else {
		address, pageCrossed := c.getAddress(instr.addrMode)
		value = sysbus.Read(address)
		if pageCrossed && (instr.addrMode == addrModeAbsoluteX ||
			instr.addrMode == addrModeAbsoluteY ||
			instr.addrMode == addrModeIndirIndexY) {
			c.cycleDelay++
		}
	}

	result := int16(c.a) - int16(value)
	if !c.testFlag(flagCarry) {
		result--
	}

	if result >= 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}

	if result == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}

	if (uint8(result)^c.a)&(uint8(result)^^value)&0x80 > 0 {
		c.setFlag(flagOverflow)
	} else {
		c.clearFlag(flagOverflow)
	}

	if result&0x80 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}

	c.a = uint8(result)
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) noOperation(instr *instruction) {
	c.pc += uint16(instr.bytes)
}
