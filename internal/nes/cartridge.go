package nes

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type Cartridge struct {
	programData   []uint8
	characterData []uint8

	mapperId            int
	programDataChunks   int
	characterDataChunks int
	horizontalMirror    bool
}

func NewCartridge(filePath string) (*Cartridge, error) {
	const errorMessage = "failed to read rom file: %w"

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(errorMessage, err)
	}
	defer file.Close()
	cartridge := &Cartridge{}

	err = cartridge.parseHeader(file)
	if err != nil {
		return nil, fmt.Errorf(errorMessage, err)
	}

	err = cartridge.parseProgramData(file)
	if err != nil {
		return nil, fmt.Errorf(errorMessage, err)
	}

	err = cartridge.parseCharacterData(file)
	if err != nil {
		return nil, fmt.Errorf(errorMessage, err)
	}

	return cartridge, nil
}

func (cartridge *Cartridge) parseHeader(file *os.File) error {
	header := make([]byte, 16)
	n, err := file.Read(header)
	if n != len(header) {
		return errors.New("unexpected end of file")
	}
	if err != nil {
		return err
	}

	cartridge.mapperId = int(header[7] & 0xF0)
	cartridge.mapperId |= int(header[6]) >> 4

	cartridge.programDataChunks = int(header[4])
	cartridge.characterDataChunks = int(header[5])
	cartridge.horizontalMirror = header[6]&0x01 > 0

	// skip trainer section if present
	if header[6]&0x04 > 0 {
		file.Seek(512, io.SeekCurrent)
	}

	return nil
}

func (cartridge *Cartridge) parseProgramData(file *os.File) error {
	cartridge.programData = make([]uint8, 16384*cartridge.programDataChunks)
	n, err := file.Read(cartridge.programData)
	if n != len(cartridge.programData) {
		return errors.New("unexpected end of file")
	}
	if err != nil {
		return err
	}
	return nil
}

func (cartridge *Cartridge) parseCharacterData(file *os.File) error {
	cartridge.characterData = make([]uint8, 8192*cartridge.characterDataChunks)
	if cartridge.characterDataChunks < 1 {
		return nil
	}
	n, err := file.Read(cartridge.characterData)
	if n != len(cartridge.characterData) {
		return errors.New("unexpected end of file")
	}
	if err != nil {
		return err
	}
	return nil
}

func (cartridge *Cartridge) ReadCharacterData(addr uint16) uint8 {
	// TODO: implement mapper
	return 0
}

func (cartridge *Cartridge) ReadProgramData(addr uint16) uint8 {
	// TODO: implement mapper
	return 0
}

func (cartridge *Cartridge) HasHorizontalNameTableMirroring() bool {
	return cartridge.horizontalMirror
}
