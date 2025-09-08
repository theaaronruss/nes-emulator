package cpu

type addressMode int

// addressing modes
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

// opcode mnemonics
const (
	adc  = "ADC"
	and  = "AND"
	asl  = "ASL"
	bcc  = "BCC"
	bcs  = "BCS"
	beq  = "BEQ"
	bit  = "BIT"
	bmi  = "BMI"
	bne  = "BNE"
	bpl  = "BPL"
	brk  = "BRK"
	bvc  = "BVC"
	bvs  = "BVS"
	clc  = "CLC"
	cld  = "CLD"
	cli  = "CLI"
	clv  = "CLV"
	cmp  = "CMP"
	cpx  = "CPX"
	cpy  = "CPY"
	dec  = "DEC"
	dex  = "DEX"
	dey  = "DEY"
	eor  = "EOR"
	inc  = "INC"
	inx  = "INX"
	iny  = "INY"
	jmp  = "JMP"
	jsr  = "JSR"
	lda  = "LDA"
	ldx  = "LDX"
	ldy  = "LDY"
	lsr  = "LSR"
	nop  = "NOP"
	ora  = "ORA"
	pha  = "PHA"
	php  = "PHP"
	pla  = "PLA"
	plp  = "PLP"
	rol  = "ROL"
	ror  = "ROR"
	rti  = "RTI"
	rts  = "RTS"
	sbc  = "SBC"
	sec  = "SEC"
	sed  = "SED"
	sei  = "SEI"
	sta  = "STA"
	stx  = "STX"
	sty  = "STY"
	tax  = "TAX"
	tay  = "TAY"
	tsx  = "TSX"
	txa  = "TXA"
	txs  = "TXS"
	tya  = "TYA"
	inop = "*NOP"
	ilax = "*LAX"
	isax = "*SAX"
	isbc = "*SBC"
	idcp = "*DCP"
	iisb = "*ISB"
	islo = "*SLO"
	irla = "*RLA"
	isre = "*SRE"
	irra = "*RRA"
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*instruction)
}

var opcodes = [256]instruction{
	0x00: {brk, addrModeImplied, 2, 7, forceBreak},
	0x01: {ora, addrModeIndexIndirX, 2, 6, bitwiseOr},
	0x03: {islo, addrModeIndexIndirX, 2, 8, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x04: {inop, addrModeZeroPage, 2, 3, illegalNoOperation},                        // illegal opcode
	0x05: {ora, addrModeZeroPage, 2, 3, bitwiseOr},
	0x06: {asl, addrModeZeroPage, 2, 5, arithmeticShiftLeft},
	0x07: {islo, addrModeZeroPage, 2, 5, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x08: {php, addrModeImplied, 1, 3, pushProcessorStatus},
	0x09: {ora, addrModeImmediate, 2, 2, bitwiseOr},
	0x0A: {asl, addrModeAccumulator, 1, 2, arithmeticShiftLeft},
	0x0C: {inop, addrModeAbsolute, 3, 4, illegalNoOperation}, // illegal opcode
	0x0D: {ora, addrModeAbsolute, 3, 4, bitwiseOr},
	0x0E: {asl, addrModeAbsolute, 3, 6, arithmeticShiftLeft},
	0x0F: {islo, addrModeAbsolute, 3, 6, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x10: {bpl, addrModeRelative, 2, 2, branchIfPlus},
	0x11: {ora, addrModeIndirIndexY, 2, 5, bitwiseOr},
	0x13: {islo, addrModeIndirIndexY, 2, 8, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x14: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},                       // illegal opcode
	0x15: {ora, addrModeZeroPageX, 2, 4, bitwiseOr},
	0x16: {asl, addrModeZeroPageX, 2, 6, arithmeticShiftLeft},
	0x17: {islo, addrModeZeroPageX, 2, 6, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x18: {clc, addrModeImplied, 1, 2, clearCarry},
	0x19: {ora, addrModeAbsoluteY, 3, 4, bitwiseOr},
	0x1A: {inop, addrModeImplied, 1, 2, illegalNoOperation},                       // illegal opcode
	0x1B: {islo, addrModeAbsoluteY, 3, 7, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x1C: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},                     // illegal opcode
	0x1D: {ora, addrModeAbsoluteX, 3, 4, bitwiseOr},
	0x1E: {asl, addrModeAbsoluteX, 3, 7, arithmeticShiftLeft},
	0x1F: {islo, addrModeAbsoluteX, 3, 7, illegalArithmeticShiftLeftAndBitwiseOr}, // illegal opcode
	0x20: {jsr, addrModeAbsolute, 3, 6, jumpToSubroutine},
	0x21: {and, addrModeIndexIndirX, 2, 6, bitwiseAnd},
	0x23: {irla, addrModeIndexIndirX, 2, 8, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x24: {bit, addrModeZeroPage, 2, 3, bitTest},
	0x25: {and, addrModeZeroPage, 2, 3, bitwiseAnd},
	0x26: {rol, addrModeZeroPage, 2, 5, rotateLeft},
	0x27: {irla, addrModeZeroPage, 2, 5, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x28: {plp, addrModeImplied, 1, 4, pullProcessorStatus},
	0x29: {and, addrModeImmediate, 2, 2, bitwiseAnd},
	0x2A: {rol, addrModeAccumulator, 1, 2, rotateLeft},
	0x2C: {bit, addrModeAbsolute, 3, 4, bitTest},
	0x2D: {and, addrModeAbsolute, 3, 4, bitwiseAnd},
	0x2E: {rol, addrModeAbsolute, 3, 6, rotateLeft},
	0x2F: {irla, addrModeAbsolute, 3, 6, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x30: {bmi, addrModeRelative, 2, 2, branchIfMinus},
	0x31: {and, addrModeIndirIndexY, 2, 5, bitwiseAnd},
	0x33: {irla, addrModeIndirIndexY, 2, 8, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x34: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},               // illegal opcode
	0x35: {and, addrModeZeroPageX, 2, 4, bitwiseAnd},
	0x36: {rol, addrModeZeroPageX, 2, 6, rotateLeft},
	0x37: {irla, addrModeZeroPageX, 2, 6, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x38: {sec, addrModeImplied, 1, 2, setCarry},
	0x39: {and, addrModeAbsoluteY, 3, 4, bitwiseAnd},
	0x3A: {inop, addrModeImplied, 1, 2, illegalNoOperation},               // illegal opcode
	0x3B: {irla, addrModeAbsoluteY, 3, 7, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x3C: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},             // illegal opcode
	0x3D: {and, addrModeAbsoluteX, 3, 4, bitwiseAnd},
	0x3E: {rol, addrModeAbsoluteX, 3, 7, rotateLeft},
	0x3F: {irla, addrModeAbsoluteX, 3, 7, illegalRotateLeftAndBitwiseAnd}, // illegal opcode
	0x40: {rti, addrModeImplied, 1, 6, returnFromInterrupt},
	0x41: {eor, addrModeIndexIndirX, 2, 6, bitwiseXor},
	0x43: {isre, addrModeIndexIndirX, 2, 8, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x44: {inop, addrModeZeroPage, 2, 3, illegalNoOperation},                       // illegal opcode
	0x45: {eor, addrModeZeroPage, 2, 3, bitwiseXor},
	0x46: {lsr, addrModeZeroPage, 2, 5, logicalShiftRight},
	0x47: {isre, addrModeZeroPage, 2, 5, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x48: {pha, addrModeImplied, 1, 3, pushA},
	0x49: {eor, addrModeImmediate, 2, 2, bitwiseXor},
	0x4A: {lsr, addrModeAccumulator, 1, 2, logicalShiftRight},
	0x4C: {jmp, addrModeAbsolute, 3, 3, jump},
	0x4D: {eor, addrModeAbsolute, 3, 4, bitwiseXor},
	0x4E: {lsr, addrModeAbsolute, 3, 6, logicalShiftRight},
	0x4F: {isre, addrModeAbsolute, 3, 6, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x50: {bvc, addrModeRelative, 2, 2, branchIfOverflowClear},
	0x51: {eor, addrModeIndirIndexY, 2, 5, bitwiseXor},
	0x53: {isre, addrModeIndirIndexY, 2, 8, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x54: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},                      // illegal opcode
	0x55: {eor, addrModeZeroPageX, 2, 4, bitwiseXor},
	0x56: {lsr, addrModeZeroPageX, 2, 6, logicalShiftRight},
	0x57: {isre, addrModeZeroPageX, 2, 6, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x58: {cli, addrModeImplied, 1, 2, clearInterruptDisable},
	0x59: {eor, addrModeAbsoluteY, 3, 4, bitwiseXor},
	0x5A: {inop, addrModeImplied, 1, 2, illegalNoOperation},                      // illegal opcode
	0x5B: {isre, addrModeAbsoluteY, 3, 7, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x5C: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},                    // illegal opcode
	0x5D: {eor, addrModeAbsoluteX, 3, 4, bitwiseXor},
	0x5E: {lsr, addrModeAbsoluteX, 3, 7, logicalShiftRight},
	0x5F: {isre, addrModeAbsoluteX, 3, 7, illegalLogicalShiftRightAndBitwiseXor}, // illegal opcode
	0x60: {rts, addrModeImplied, 1, 6, returnFromSubroutine},
	0x61: {adc, addrModeIndexIndirX, 2, 6, addWithCarry},
	0x63: {irra, addrModeIndexIndirX, 2, 8, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x64: {inop, addrModeZeroPage, 2, 3, illegalNoOperation},                   // illegal opcode
	0x65: {adc, addrModeZeroPage, 2, 3, addWithCarry},
	0x66: {ror, addrModeZeroPage, 2, 5, rotateRight},
	0x67: {irra, addrModeZeroPage, 2, 5, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x68: {pla, addrModeImplied, 1, 4, pullA},
	0x69: {adc, addrModeImmediate, 2, 2, addWithCarry},
	0x6A: {ror, addrModeAccumulator, 1, 2, rotateRight},
	0x6C: {jmp, addrModeIndirect, 3, 5, jump},
	0x6D: {adc, addrModeAbsolute, 3, 4, addWithCarry},
	0x6E: {ror, addrModeAbsolute, 3, 6, rotateRight},
	0x6F: {irra, addrModeAbsolute, 3, 6, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x70: {bvs, addrModeRelative, 2, 2, branchIfOverflowSet},
	0x71: {adc, addrModeIndirIndexY, 2, 5, addWithCarry},
	0x73: {irra, addrModeIndirIndexY, 2, 8, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x74: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},                  // illegal opcode
	0x75: {adc, addrModeZeroPageX, 2, 4, addWithCarry},
	0x76: {ror, addrModeZeroPageX, 2, 6, rotateRight},
	0x77: {irra, addrModeZeroPageX, 2, 6, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x78: {sei, addrModeImplied, 1, 2, setInterruptDisable},
	0x79: {adc, addrModeAbsoluteY, 3, 4, addWithCarry},
	0x7A: {inop, addrModeImplied, 1, 2, illegalNoOperation},                  // illegal opcode
	0x7B: {irra, addrModeAbsoluteY, 3, 7, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x7C: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},                // illegal opcode
	0x7D: {adc, addrModeAbsoluteX, 3, 4, addWithCarry},
	0x7E: {ror, addrModeAbsoluteX, 3, 7, rotateRight},
	0x7F: {irra, addrModeAbsoluteX, 3, 7, illegalRotateRightAndAddWithCarry}, // illegal opcode
	0x80: {inop, addrModeImmediate, 2, 2, illegalNoOperation},                // illegal opcode
	0x81: {sta, addrModeIndexIndirX, 2, 6, storeA},
	0x83: {isax, addrModeIndexIndirX, 2, 6, illegalStoreAAndX}, // illegal opcode
	0x84: {sty, addrModeZeroPage, 2, 3, storeY},
	0x85: {sta, addrModeZeroPage, 2, 3, storeA},
	0x86: {stx, addrModeZeroPage, 2, 3, storeX},
	0x87: {isax, addrModeZeroPage, 2, 3, illegalStoreAAndX}, // illegal opcode
	0x88: {dey, addrModeImplied, 1, 2, decrementY},
	0x8A: {txa, addrModeImplied, 1, 2, transferXToA},
	0x8C: {sty, addrModeAbsolute, 3, 4, storeY},
	0x8D: {sta, addrModeAbsolute, 3, 4, storeA},
	0x8E: {stx, addrModeAbsolute, 3, 4, storeX},
	0x8F: {isax, addrModeAbsolute, 3, 4, illegalStoreAAndX}, // illegal opcode
	0x90: {bcc, addrModeRelative, 2, 2, branchIfCarryClear},
	0x91: {sta, addrModeIndirIndexY, 2, 6, storeA},
	0x94: {sty, addrModeZeroPageX, 2, 4, storeY},
	0x95: {sta, addrModeZeroPageX, 2, 4, storeA},
	0x96: {stx, addrModeZeroPageY, 2, 4, storeX},
	0x97: {isax, addrModeZeroPageY, 2, 4, illegalStoreAAndX}, // illegal opcode
	0x98: {tya, addrModeImplied, 1, 2, transferYToA},
	0x99: {sta, addrModeAbsoluteY, 3, 5, storeA},
	0x9A: {txs, addrModeImplied, 1, 2, transferXToStackPointer},
	0x9D: {sta, addrModeAbsoluteX, 3, 5, storeA},
	0xA0: {ldy, addrModeImmediate, 2, 2, loadY},
	0xA1: {lda, addrModeIndexIndirX, 2, 6, loadA},
	0xA2: {ldx, addrModeImmediate, 2, 2, loadX},
	0xA3: {ilax, addrModeIndexIndirX, 2, 6, illegalLoadALoadX}, // illegal opcode
	0xA4: {ldy, addrModeZeroPage, 2, 3, loadY},
	0xA5: {lda, addrModeZeroPage, 2, 3, loadA},
	0xA6: {ldx, addrModeZeroPage, 2, 3, loadX},
	0xA7: {ilax, addrModeZeroPage, 2, 3, illegalLoadALoadX}, // illegal opcode
	0xA8: {tay, addrModeImplied, 1, 2, transferAToY},
	0xA9: {lda, addrModeImmediate, 2, 2, loadA},
	0xAA: {tax, addrModeImplied, 1, 2, transferAToX},
	0xAC: {ldy, addrModeAbsolute, 3, 4, loadY},
	0xAD: {lda, addrModeAbsolute, 3, 4, loadA},
	0xAE: {ldx, addrModeAbsolute, 3, 4, loadX},
	0xAF: {ilax, addrModeAbsolute, 3, 4, illegalLoadALoadX}, // illegal opcode
	0xB0: {bcs, addrModeRelative, 2, 2, branchIfCarrySet},
	0xB1: {lda, addrModeIndirIndexY, 2, 5, loadA},
	0xB3: {ilax, addrModeIndirIndexY, 2, 5, illegalLoadALoadX}, // illegal opcode
	0xB4: {ldy, addrModeZeroPageX, 2, 4, loadY},
	0xB5: {lda, addrModeZeroPageX, 2, 4, loadA},
	0xB6: {ldx, addrModeZeroPageY, 2, 4, loadX},
	0xB7: {ilax, addrModeZeroPageY, 2, 4, illegalLoadALoadX}, // illegal opcode
	0xB8: {clv, addrModeImplied, 1, 2, clearOverflow},
	0xB9: {lda, addrModeAbsoluteY, 3, 4, loadA},
	0xBA: {tsx, addrModeImplied, 1, 2, transferStackPointerToX},
	0xBC: {ldy, addrModeAbsoluteX, 3, 4, loadY},
	0xBD: {lda, addrModeAbsoluteX, 3, 4, loadA},
	0xBE: {ldx, addrModeAbsoluteY, 3, 4, loadX},
	0xBF: {ilax, addrModeAbsoluteY, 3, 4, illegalLoadALoadX}, // illegal opcode
	0xC0: {cpy, addrModeImmediate, 2, 2, compareY},
	0xC1: {cmp, addrModeIndexIndirX, 2, 6, compareA},
	0xC3: {idcp, addrModeIndexIndirX, 2, 8, illegalDecrementAndCompare}, // illegal opcode
	0xC4: {cpy, addrModeZeroPage, 2, 3, compareY},
	0xC5: {cmp, addrModeZeroPage, 2, 3, compareA},
	0xC6: {dec, addrModeZeroPage, 2, 5, decrementMemory},
	0xC7: {idcp, addrModeZeroPage, 2, 5, illegalDecrementAndCompare}, // illegal opcode
	0xC8: {iny, addrModeImplied, 1, 2, incrementY},
	0xC9: {cmp, addrModeImmediate, 2, 2, compareA},
	0xCA: {dex, addrModeImplied, 1, 2, decrementX},
	0xCC: {cpy, addrModeAbsolute, 3, 4, compareY},
	0xCD: {cmp, addrModeAbsolute, 3, 4, compareA},
	0xCE: {dec, addrModeAbsolute, 3, 6, decrementMemory},
	0xCF: {idcp, addrModeAbsolute, 3, 6, illegalDecrementAndCompare}, // illegal opcode
	0xD0: {bne, addrModeRelative, 2, 2, branchIfNotEqual},
	0xD1: {cmp, addrModeIndirIndexY, 2, 5, compareA},
	0xD3: {idcp, addrModeIndirIndexY, 2, 8, illegalDecrementAndCompare}, // illegal opcode
	0xD4: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},           // illegal opcode
	0xD5: {cmp, addrModeZeroPageX, 2, 4, compareA},
	0xD6: {dec, addrModeZeroPageX, 2, 6, decrementMemory},
	0xD7: {idcp, addrModeZeroPageX, 2, 6, illegalDecrementAndCompare}, // illegal opcode
	0xD8: {cld, addrModeImplied, 1, 2, clearDecimal},
	0xD9: {cmp, addrModeAbsoluteY, 3, 4, compareA},
	0xDA: {inop, addrModeImplied, 1, 2, illegalNoOperation},           // illegal opcode
	0xDB: {idcp, addrModeAbsoluteY, 3, 7, illegalDecrementAndCompare}, // illegal opcode
	0xDC: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},         // illegal opcode
	0xDD: {cmp, addrModeAbsoluteX, 3, 4, compareA},
	0xDE: {dec, addrModeAbsoluteX, 3, 7, decrementMemory},
	0xDF: {idcp, addrModeAbsoluteX, 3, 7, illegalDecrementAndCompare}, // illegal opcode
	0xE0: {cpx, addrModeImmediate, 2, 2, compareX},
	0xE1: {sbc, addrModeIndexIndirX, 2, 6, subtractWithCarry},
	0xE3: {iisb, addrModeIndexIndirX, 2, 8, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xE4: {cpx, addrModeZeroPage, 2, 3, compareX},
	0xE5: {sbc, addrModeZeroPage, 2, 3, subtractWithCarry},
	0xE6: {inc, addrModeZeroPage, 2, 5, incrementMemory},
	0xE7: {iisb, addrModeZeroPage, 2, 5, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xE8: {inx, addrModeImplied, 1, 2, incrementX},
	0xE9: {sbc, addrModeImmediate, 2, 2, subtractWithCarry},
	0xEA: {nop, addrModeImplied, 1, 2, noOperation},
	0xEB: {isbc, addrModeImmediate, 2, 2, subtractWithCarry},
	0xEC: {cpx, addrModeAbsolute, 3, 4, compareX},
	0xED: {sbc, addrModeAbsolute, 3, 4, subtractWithCarry},
	0xEE: {inc, addrModeAbsolute, 3, 6, incrementMemory},
	0xEF: {iisb, addrModeAbsolute, 3, 6, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xF0: {beq, addrModeRelative, 2, 2, branchIfEqual},
	0xF1: {sbc, addrModeIndirIndexY, 2, 5, subtractWithCarry},
	0xF3: {iisb, addrModeIndirIndexY, 2, 8, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xF4: {inop, addrModeZeroPageX, 2, 4, illegalNoOperation},                  // illegal opcode
	0xF5: {sbc, addrModeZeroPageX, 2, 4, subtractWithCarry},
	0xF6: {inc, addrModeZeroPageX, 2, 6, incrementMemory},
	0xF7: {iisb, addrModeZeroPageX, 2, 6, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xF8: {sed, addrModeImplied, 1, 2, setDecimal},
	0xF9: {sbc, addrModeAbsoluteY, 3, 4, subtractWithCarry},
	0xFA: {inop, addrModeImplied, 1, 2, illegalNoOperation},                  // illegal opcode
	0xFB: {iisb, addrModeAbsoluteY, 3, 7, illegalIncrementSubtractWithCarry}, // illegal opcode
	0xFC: {inop, addrModeAbsoluteX, 3, 4, illegalNoOperation},                // illegal opcode
	0xFD: {sbc, addrModeAbsoluteX, 3, 4, subtractWithCarry},
	0xFE: {inc, addrModeAbsoluteX, 3, 7, incrementMemory},
	0xFF: {iisb, addrModeAbsoluteX, 3, 7, illegalIncrementSubtractWithCarry}, // illegal opcode
}
