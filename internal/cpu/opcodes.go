package cpu

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexIndirX
	addrModeIndirIndexY
	addrModeImplied
	addrModeAccumulator
	addrModeImmediate
	addrModeZeroPage
	addrModeAbsolute
	addrModeRelative
	addrModeIndirect
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*Cpu, *instruction)
}

var opcodes = [256]instruction{
	0x00: {"BRK", addrModeImplied, 2, 7, (*Cpu).forceBreak},
	0x01: {"ORA", addrModeIndexIndirX, 2, 6, (*Cpu).bitwiseOr},
	0x05: {"ORA", addrModeZeroPage, 2, 3, (*Cpu).bitwiseOr},
	0x06: {"ASL", addrModeZeroPage, 2, 5, (*Cpu).arithmeticShiftLeft},
	0x08: {"PHP", addrModeImplied, 1, 3, (*Cpu).pushProcessorStatus},
	0x09: {"ORA", addrModeImmediate, 2, 2, (*Cpu).bitwiseOr},
	0x0A: {"ASL", addrModeAccumulator, 1, 2, (*Cpu).arithmeticShiftLeft},
	0x0D: {"ORA", addrModeAbsolute, 3, 4, (*Cpu).bitwiseOr},
	0x0E: {"ASL", addrModeAbsolute, 3, 6, (*Cpu).arithmeticShiftLeft},
	0x10: {"BPL", addrModeRelative, 2, 0, (*Cpu).branchIfPlus},
	0x11: {"ORA", addrModeIndirIndexY, 2, 0, (*Cpu).bitwiseOr},
	0x15: {"ORA", addrModeZeroPageX, 2, 4, (*Cpu).bitwiseOr},
	0x16: {"ASL", addrModeZeroPageX, 2, 6, (*Cpu).arithmeticShiftLeft},
	0x18: {"CLC", addrModeImplied, 1, 2, (*Cpu).clearCarry},
	0x19: {"ORA", addrModeAbsoluteY, 3, 0, (*Cpu).bitwiseOr},
	0x1D: {"ORA", addrModeAbsoluteX, 3, 0, (*Cpu).bitwiseOr},
	0x1E: {"ASL", addrModeAbsoluteX, 3, 7, (*Cpu).arithmeticShiftLeft},
	0x20: {"JSR", addrModeAbsolute, 3, 6, (*Cpu).jumpToSubroutine},
	0x21: {"AND", addrModeIndexIndirX, 2, 6, (*Cpu).bitwiseAnd},
	0x24: {"BIT", addrModeZeroPage, 2, 3, (*Cpu).bitTest},
	0x25: {"AND", addrModeZeroPage, 2, 3, (*Cpu).bitwiseAnd},
	0x26: {"ROL", addrModeZeroPage, 2, 5, (*Cpu).rotateLeft},
	0x28: {"PLP", addrModeImplied, 1, 4, (*Cpu).pullProcessorStatus},
	0x29: {"AND", addrModeImmediate, 2, 2, (*Cpu).bitwiseAnd},
	0x2A: {"ROL", addrModeAccumulator, 1, 2, (*Cpu).rotateLeft},
	0x2C: {"BIT", addrModeAbsolute, 3, 4, (*Cpu).bitTest},
	0x2D: {"AND", addrModeAbsolute, 3, 4, (*Cpu).bitwiseAnd},
	0x2E: {"ROL", addrModeAbsolute, 3, 6, (*Cpu).rotateLeft},
	0x30: {"BMI", addrModeRelative, 2, 0, (*Cpu).branchIfMinus},
	0x31: {"AND", addrModeIndirIndexY, 2, 0, (*Cpu).bitwiseAnd},
	0x35: {"AND", addrModeZeroPageX, 2, 4, (*Cpu).bitwiseAnd},
	0x36: {"ROL", addrModeZeroPageX, 2, 6, (*Cpu).rotateLeft},
	0x38: {"SEC", addrModeImplied, 1, 2, (*Cpu).setCarry},
	0x39: {"AND", addrModeAbsoluteY, 3, 0, (*Cpu).bitwiseAnd},
	0x3D: {"AND", addrModeAbsoluteX, 3, 0, (*Cpu).bitwiseAnd},
	0x3E: {"ROL", addrModeAbsoluteX, 3, 7, (*Cpu).rotateLeft},
	0x40: {"RTI", addrModeImplied, 1, 6, (*Cpu).returnFromInterrupt},
	0x41: {"EOR", addrModeIndexIndirX, 2, 6, (*Cpu).bitwiseXor},
	0x45: {"EOR", addrModeZeroPage, 2, 3, (*Cpu).bitwiseXor},
	0x46: {"LSR", addrModeZeroPage, 2, 5, (*Cpu).logicalShiftRight},
	0x48: {"PHA", addrModeImplied, 1, 3, (*Cpu).pushA},
	0x49: {"EOR", addrModeImmediate, 2, 2, (*Cpu).bitwiseXor},
	0x4A: {"LSR", addrModeAccumulator, 1, 2, (*Cpu).logicalShiftRight},
	0x4C: {"JMP", addrModeAbsolute, 3, 3, (*Cpu).jump},
	0x4D: {"EOR", addrModeAbsolute, 3, 4, (*Cpu).bitwiseXor},
	0x4E: {"LSR", addrModeAbsolute, 3, 6, (*Cpu).logicalShiftRight},
	0x50: {"BVC", addrModeRelative, 2, 0, (*Cpu).branchIfOverflowClear},
	0x51: {"EOR", addrModeIndirIndexY, 2, 0, (*Cpu).bitwiseXor},
	0x55: {"EOR", addrModeZeroPageX, 2, 4, (*Cpu).bitwiseXor},
	0x56: {"LSR", addrModeZeroPageX, 2, 6, (*Cpu).logicalShiftRight},
	0x58: {"CLI", addrModeImplied, 1, 2, (*Cpu).clearInterruptDisable},
	0x59: {"EOR", addrModeAbsoluteY, 3, 0, (*Cpu).bitwiseXor},
	0x5D: {"EOR", addrModeAbsoluteX, 3, 0, (*Cpu).bitwiseXor},
	0x5E: {"LSR", addrModeAbsoluteX, 3, 7, (*Cpu).logicalShiftRight},
	0x60: {"RTS", addrModeImplied, 1, 6, (*Cpu).returnFromSubroutine},
}
