package cpu

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexIndirX
	addrModeIndirIndexY
	addrModeImplicit
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
	0x00: {"BRK", addrModeImplicit, 2, 7, (*Cpu).forceBreak},
	0x01: {"ORA", addrModeIndexIndirX, 2, 6, (*Cpu).bitwiseOr},
	0x05: {"ORA", addrModeZeroPage, 2, 3, (*Cpu).bitwiseOr},
}
