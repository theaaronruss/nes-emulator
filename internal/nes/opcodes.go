package nes

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexedIndir
	addrModeIndirIndexed
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
	idcp = "*DCP"
	iisb = "*ISB"
	ilax = "*LAX"
	inop = "*NOP"
	irla = "*RLA"
	irra = "*RRA"
	isax = "*SAX"
	isbc = "*SBC"
	islo = "*SLO"
	isre = "*SRE"
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*cpu, addressMode, uint16)
}

var opcodes = [256]instruction{
	0x00: {brk, addrModeImplied, 2, 7, (*cpu).brk},
	0x01: {ora, addrModeIndexedIndir, 2, 6, (*cpu).ora},
	0x03: {islo, addrModeIndexedIndir, 2, 8, (*cpu).slo},
	0x04: {inop, addrModeZeroPage, 2, 3, (*cpu).nop},
	0x05: {ora, addrModeZeroPage, 2, 3, (*cpu).ora},
	0x06: {asl, addrModeZeroPage, 2, 5, (*cpu).asl},
	0x07: {islo, addrModeZeroPage, 2, 5, (*cpu).slo},
	0x08: {php, addrModeImplied, 1, 3, (*cpu).php},
	0x09: {ora, addrModeImmediate, 2, 2, (*cpu).ora},
	0x0A: {asl, addrModeAccumulator, 1, 2, (*cpu).asl},
	0x0C: {inop, addrModeAbsolute, 3, 4, (*cpu).nop},
	0x0D: {ora, addrModeAbsolute, 3, 4, (*cpu).ora},
	0x0E: {asl, addrModeAbsolute, 3, 6, (*cpu).asl},
	0x0F: {islo, addrModeAbsolute, 3, 6, (*cpu).slo},
	0x10: {bpl, addrModeRelative, 2, 2, (*cpu).bpl},
	0x11: {ora, addrModeIndirIndexed, 2, 5, (*cpu).ora},
	0x13: {islo, addrModeIndirIndexed, 2, 8, (*cpu).slo},
	0x14: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0x15: {ora, addrModeZeroPageX, 2, 4, (*cpu).ora},
	0x16: {asl, addrModeZeroPageX, 2, 6, (*cpu).asl},
	0x17: {islo, addrModeZeroPageX, 2, 6, (*cpu).slo},
	0x18: {clc, addrModeImplied, 1, 2, (*cpu).clc},
	0x19: {ora, addrModeAbsoluteY, 3, 4, (*cpu).ora},
	0x1A: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0x1B: {islo, addrModeAbsoluteY, 3, 7, (*cpu).slo},
	0x1C: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0x1D: {ora, addrModeAbsoluteX, 3, 4, (*cpu).ora},
	0x1E: {asl, addrModeAbsoluteX, 3, 7, (*cpu).asl},
	0x1F: {islo, addrModeAbsoluteX, 3, 7, (*cpu).slo},
	0x20: {jsr, addrModeAbsolute, 3, 6, (*cpu).jsr},
	0x21: {and, addrModeIndexedIndir, 2, 6, (*cpu).and},
	0x23: {irla, addrModeIndexedIndir, 2, 8, (*cpu).rla},
	0x24: {bit, addrModeZeroPage, 2, 3, (*cpu).bit},
	0x25: {and, addrModeZeroPage, 2, 3, (*cpu).and},
	0x26: {rol, addrModeZeroPage, 2, 5, (*cpu).rol},
	0x27: {irla, addrModeZeroPage, 2, 5, (*cpu).rla},
	0x28: {plp, addrModeImplied, 1, 4, (*cpu).plp},
	0x29: {and, addrModeImmediate, 2, 2, (*cpu).and},
	0x2A: {rol, addrModeAccumulator, 1, 2, (*cpu).rol},
	0x2C: {bit, addrModeAbsolute, 3, 4, (*cpu).bit},
	0x2D: {and, addrModeAbsolute, 3, 4, (*cpu).and},
	0x2E: {rol, addrModeAbsolute, 3, 6, (*cpu).rol},
	0x2F: {irla, addrModeAbsolute, 3, 6, (*cpu).rla},
	0x30: {bmi, addrModeRelative, 2, 2, (*cpu).bmi},
	0x31: {and, addrModeIndirIndexed, 2, 5, (*cpu).and},
	0x33: {irla, addrModeIndirIndexed, 2, 8, (*cpu).rla},
	0x34: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0x35: {and, addrModeZeroPageX, 2, 4, (*cpu).and},
	0x36: {rol, addrModeZeroPageX, 2, 6, (*cpu).rol},
	0x37: {irla, addrModeZeroPageX, 2, 6, (*cpu).rla},
	0x38: {sec, addrModeImplied, 1, 2, (*cpu).sec},
	0x39: {and, addrModeAbsoluteY, 3, 4, (*cpu).and},
	0x3A: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0x3B: {irla, addrModeAbsoluteY, 3, 7, (*cpu).rla},
	0x3C: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0x3D: {and, addrModeAbsoluteX, 3, 4, (*cpu).and},
	0x3E: {rol, addrModeAbsoluteX, 3, 7, (*cpu).rol},
	0x3F: {irla, addrModeAbsoluteX, 3, 7, (*cpu).rla},
	0x40: {rti, addrModeImplied, 1, 6, (*cpu).rti},
	0x41: {eor, addrModeIndexedIndir, 2, 6, (*cpu).eor},
	0x43: {isre, addrModeIndexedIndir, 2, 8, (*cpu).sre},
	0x44: {inop, addrModeZeroPage, 2, 3, (*cpu).nop},
	0x45: {eor, addrModeZeroPage, 2, 3, (*cpu).eor},
	0x46: {lsr, addrModeZeroPage, 2, 5, (*cpu).lsr},
	0x47: {isre, addrModeZeroPage, 2, 5, (*cpu).sre},
	0x48: {pha, addrModeImplied, 1, 3, (*cpu).pha},
	0x49: {eor, addrModeImmediate, 2, 2, (*cpu).eor},
	0x4A: {lsr, addrModeAccumulator, 1, 2, (*cpu).lsr},
	0x4C: {jmp, addrModeAbsolute, 3, 3, (*cpu).jmp},
	0x4D: {eor, addrModeAbsolute, 3, 4, (*cpu).eor},
	0x4E: {lsr, addrModeAbsolute, 3, 6, (*cpu).lsr},
	0x4F: {isre, addrModeAbsolute, 3, 6, (*cpu).sre},
	0x50: {bvc, addrModeRelative, 2, 2, (*cpu).bvc},
	0x51: {eor, addrModeIndirIndexed, 2, 5, (*cpu).eor},
	0x53: {isre, addrModeIndirIndexed, 2, 8, (*cpu).sre},
	0x54: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0x55: {eor, addrModeZeroPageX, 2, 4, (*cpu).eor},
	0x56: {lsr, addrModeZeroPageX, 2, 6, (*cpu).lsr},
	0x57: {isre, addrModeZeroPageX, 2, 6, (*cpu).sre},
	0x58: {cli, addrModeImplied, 1, 2, (*cpu).cli},
	0x59: {eor, addrModeAbsoluteY, 3, 4, (*cpu).eor},
	0x5A: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0x5B: {isre, addrModeAbsoluteY, 3, 7, (*cpu).sre},
	0x5C: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0x5D: {eor, addrModeAbsoluteX, 3, 4, (*cpu).eor},
	0x5E: {lsr, addrModeAbsoluteX, 3, 7, (*cpu).lsr},
	0x5F: {isre, addrModeAbsoluteX, 3, 7, (*cpu).sre},
	0x60: {rts, addrModeImplied, 1, 6, (*cpu).rts},
	0x61: {adc, addrModeIndexedIndir, 2, 6, (*cpu).adc},
	0x63: {irra, addrModeIndexedIndir, 2, 8, (*cpu).rra},
	0x64: {inop, addrModeZeroPage, 2, 3, (*cpu).nop},
	0x65: {adc, addrModeZeroPage, 2, 3, (*cpu).adc},
	0x66: {ror, addrModeZeroPage, 2, 5, (*cpu).ror},
	0x67: {irra, addrModeZeroPage, 2, 5, (*cpu).rra},
	0x68: {pla, addrModeImplied, 1, 4, (*cpu).pla},
	0x69: {adc, addrModeImmediate, 2, 2, (*cpu).adc},
	0x6A: {ror, addrModeAccumulator, 1, 2, (*cpu).ror},
	0x6C: {jmp, addrModeIndirect, 3, 5, (*cpu).jmp},
	0x6D: {adc, addrModeAbsolute, 3, 4, (*cpu).adc},
	0x6E: {ror, addrModeAbsolute, 3, 6, (*cpu).ror},
	0x6F: {irra, addrModeAbsolute, 3, 6, (*cpu).rra},
	0x70: {bvs, addrModeRelative, 2, 2, (*cpu).bvs},
	0x71: {adc, addrModeIndirIndexed, 2, 5, (*cpu).adc},
	0x73: {irra, addrModeIndirIndexed, 2, 8, (*cpu).rra},
	0x74: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0x75: {adc, addrModeZeroPageX, 2, 4, (*cpu).adc},
	0x76: {ror, addrModeZeroPageX, 2, 6, (*cpu).ror},
	0x77: {irra, addrModeZeroPageX, 2, 6, (*cpu).rra},
	0x78: {sei, addrModeImplied, 1, 2, (*cpu).sei},
	0x79: {adc, addrModeAbsoluteY, 3, 4, (*cpu).adc},
	0x7A: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0x7B: {irra, addrModeAbsoluteY, 3, 7, (*cpu).rra},
	0x7C: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0x7D: {adc, addrModeAbsoluteX, 3, 4, (*cpu).adc},
	0x7E: {ror, addrModeAbsoluteX, 3, 7, (*cpu).ror},
	0x7F: {irra, addrModeAbsoluteX, 3, 7, (*cpu).rra},
	0x80: {inop, addrModeImmediate, 2, 2, (*cpu).nop},
	0x81: {sta, addrModeIndexedIndir, 2, 6, (*cpu).sta},
	0x83: {isax, addrModeIndexedIndir, 2, 6, (*cpu).sax},
	0x84: {sty, addrModeZeroPage, 2, 3, (*cpu).sty},
	0x85: {sta, addrModeZeroPage, 2, 3, (*cpu).sta},
	0x86: {stx, addrModeZeroPage, 2, 3, (*cpu).stx},
	0x87: {isax, addrModeZeroPage, 2, 3, (*cpu).sax},
	0x88: {dey, addrModeImplied, 1, 2, (*cpu).dey},
	0x8A: {txa, addrModeImplied, 1, 2, (*cpu).txa},
	0x8C: {sty, addrModeAbsolute, 3, 4, (*cpu).sty},
	0x8D: {sta, addrModeAbsolute, 3, 4, (*cpu).sta},
	0x8E: {stx, addrModeAbsolute, 3, 4, (*cpu).stx},
	0x8F: {isax, addrModeAbsolute, 3, 4, (*cpu).sax},
	0x90: {bcc, addrModeRelative, 2, 2, (*cpu).bcc},
	0x91: {sta, addrModeIndirIndexed, 2, 6, (*cpu).sta},
	0x94: {sty, addrModeZeroPageX, 2, 4, (*cpu).sty},
	0x95: {sta, addrModeZeroPageX, 2, 4, (*cpu).sta},
	0x96: {stx, addrModeZeroPageY, 2, 4, (*cpu).stx},
	0x97: {isax, addrModeZeroPageY, 2, 4, (*cpu).sax},
	0x98: {tya, addrModeImplied, 1, 2, (*cpu).tya},
	0x99: {sta, addrModeAbsoluteY, 3, 5, (*cpu).sta},
	0x9A: {txs, addrModeImplied, 1, 2, (*cpu).txs},
	0x9D: {sta, addrModeAbsoluteX, 3, 5, (*cpu).sta},
	0xA0: {ldy, addrModeImmediate, 2, 2, (*cpu).ldy},
	0xA1: {lda, addrModeIndexedIndir, 2, 6, (*cpu).lda},
	0xA2: {ldx, addrModeImmediate, 2, 2, (*cpu).ldx},
	0xA3: {ilax, addrModeIndexedIndir, 2, 6, (*cpu).lax},
	0xA4: {ldy, addrModeZeroPage, 2, 3, (*cpu).ldy},
	0xA5: {lda, addrModeZeroPage, 2, 3, (*cpu).lda},
	0xA6: {ldx, addrModeZeroPage, 2, 3, (*cpu).ldx},
	0xA7: {ilax, addrModeZeroPage, 2, 3, (*cpu).lax},
	0xA8: {tay, addrModeImplied, 1, 2, (*cpu).tay},
	0xA9: {lda, addrModeImmediate, 2, 2, (*cpu).lda},
	0xAA: {tax, addrModeImplied, 1, 2, (*cpu).tax},
	0xAC: {ldy, addrModeAbsolute, 3, 4, (*cpu).ldy},
	0xAD: {lda, addrModeAbsolute, 3, 4, (*cpu).lda},
	0xAE: {ldx, addrModeAbsolute, 3, 4, (*cpu).ldx},
	0xAF: {ilax, addrModeAbsolute, 3, 4, (*cpu).lax},
	0xB0: {bcs, addrModeRelative, 2, 2, (*cpu).bcs},
	0xB1: {lda, addrModeIndirIndexed, 2, 5, (*cpu).lda},
	0xB3: {ilax, addrModeIndirIndexed, 2, 5, (*cpu).lax},
	0xB4: {ldy, addrModeZeroPageX, 2, 4, (*cpu).ldy},
	0xB5: {lda, addrModeZeroPageX, 2, 4, (*cpu).lda},
	0xB6: {ldx, addrModeZeroPageY, 2, 4, (*cpu).ldx},
	0xB7: {ilax, addrModeZeroPageY, 2, 4, (*cpu).lax},
	0xB8: {clv, addrModeImplied, 1, 2, (*cpu).clv},
	0xB9: {lda, addrModeAbsoluteY, 3, 4, (*cpu).lda},
	0xBA: {tsx, addrModeImplied, 1, 2, (*cpu).tsx},
	0xBC: {ldy, addrModeAbsoluteX, 3, 4, (*cpu).ldy},
	0xBD: {lda, addrModeAbsoluteX, 3, 4, (*cpu).lda},
	0xBE: {ldx, addrModeAbsoluteY, 3, 4, (*cpu).ldx},
	0xBF: {ilax, addrModeAbsoluteY, 3, 4, (*cpu).lax},
	0xC0: {cpy, addrModeImmediate, 2, 2, (*cpu).cpy},
	0xC1: {cmp, addrModeIndexedIndir, 2, 6, (*cpu).cmp},
	0xC3: {idcp, addrModeIndexedIndir, 2, 8, (*cpu).dcp},
	0xC4: {cpy, addrModeZeroPage, 2, 3, (*cpu).cpy},
	0xC5: {cmp, addrModeZeroPage, 2, 3, (*cpu).cmp},
	0xC6: {dec, addrModeZeroPage, 2, 5, (*cpu).dec},
	0xC7: {idcp, addrModeZeroPage, 2, 5, (*cpu).dcp},
	0xC8: {iny, addrModeImplied, 1, 2, (*cpu).iny},
	0xC9: {cmp, addrModeImmediate, 2, 2, (*cpu).cmp},
	0xCA: {dex, addrModeImplied, 1, 2, (*cpu).dex},
	0xCC: {cpy, addrModeAbsolute, 3, 4, (*cpu).cpy},
	0xCD: {cmp, addrModeAbsolute, 3, 4, (*cpu).cmp},
	0xCE: {dec, addrModeAbsolute, 3, 6, (*cpu).dec},
	0xCF: {idcp, addrModeAbsolute, 3, 6, (*cpu).dcp},
	0xD0: {bne, addrModeRelative, 2, 2, (*cpu).bne},
	0xD1: {cmp, addrModeIndirIndexed, 2, 5, (*cpu).cmp},
	0xD3: {idcp, addrModeIndirIndexed, 2, 8, (*cpu).dcp},
	0xD4: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0xD5: {cmp, addrModeZeroPageX, 2, 4, (*cpu).cmp},
	0xD6: {dec, addrModeZeroPageX, 2, 6, (*cpu).dec},
	0xD7: {idcp, addrModeZeroPageX, 2, 6, (*cpu).dcp},
	0xD8: {cld, addrModeImplied, 1, 2, (*cpu).cld},
	0xD9: {cmp, addrModeAbsoluteY, 3, 4, (*cpu).cmp},
	0xDA: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0xDB: {idcp, addrModeAbsoluteY, 3, 7, (*cpu).dcp},
	0xDC: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0xDD: {cmp, addrModeAbsoluteX, 3, 4, (*cpu).cmp},
	0xDE: {dec, addrModeAbsoluteX, 3, 7, (*cpu).dec},
	0xDF: {idcp, addrModeAbsoluteX, 3, 7, (*cpu).dcp},
	0xE0: {cpx, addrModeImmediate, 2, 2, (*cpu).cpx},
	0xE1: {sbc, addrModeIndexedIndir, 2, 6, (*cpu).sbc},
	0xE3: {iisb, addrModeIndexedIndir, 2, 8, (*cpu).isb},
	0xE4: {cpx, addrModeZeroPage, 2, 3, (*cpu).cpx},
	0xE5: {sbc, addrModeZeroPage, 2, 3, (*cpu).sbc},
	0xE6: {inc, addrModeZeroPage, 2, 5, (*cpu).inc},
	0xE7: {iisb, addrModeZeroPage, 2, 5, (*cpu).isb},
	0xE8: {inx, addrModeImplied, 1, 2, (*cpu).inx},
	0xE9: {sbc, addrModeImmediate, 2, 2, (*cpu).sbc},
	0xEA: {nop, addrModeImplied, 1, 2, (*cpu).nop},
	0xEB: {isbc, addrModeImmediate, 2, 2, (*cpu).sbc},
	0xEC: {cpx, addrModeAbsolute, 3, 4, (*cpu).cpx},
	0xED: {sbc, addrModeAbsolute, 3, 4, (*cpu).sbc},
	0xEE: {inc, addrModeAbsolute, 3, 6, (*cpu).inc},
	0xEF: {iisb, addrModeAbsolute, 3, 6, (*cpu).isb},
	0xF0: {beq, addrModeRelative, 2, 2, (*cpu).beq},
	0xF1: {sbc, addrModeIndirIndexed, 2, 5, (*cpu).sbc},
	0xF3: {iisb, addrModeIndirIndexed, 2, 8, (*cpu).isb},
	0xF4: {inop, addrModeZeroPageX, 2, 4, (*cpu).nop},
	0xF5: {sbc, addrModeZeroPageX, 2, 4, (*cpu).sbc},
	0xF6: {inc, addrModeZeroPageX, 2, 6, (*cpu).inc},
	0xF7: {iisb, addrModeZeroPageX, 2, 6, (*cpu).isb},
	0xF8: {sed, addrModeImplied, 1, 2, (*cpu).sed},
	0xF9: {sbc, addrModeAbsoluteY, 3, 4, (*cpu).sbc},
	0xFA: {inop, addrModeImplied, 1, 2, (*cpu).nop},
	0xFB: {iisb, addrModeAbsoluteY, 3, 7, (*cpu).isb},
	0xFC: {inop, addrModeAbsoluteX, 3, 4, (*cpu).nop},
	0xFD: {sbc, addrModeAbsoluteX, 3, 4, (*cpu).sbc},
	0xFE: {inc, addrModeAbsoluteX, 3, 7, (*cpu).inc},
	0xFF: {iisb, addrModeAbsoluteX, 3, 7, (*cpu).isb},
}
