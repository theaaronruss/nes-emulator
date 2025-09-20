package nes

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	headerSize        int = 16
	programDataSize   int = 16384
	characterDataSize int = 8192
)

type Cartridge struct {
	programData   []byte
	characterData []byte
}

func NewCartridge(filePath string) (*Cartridge, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cartridge := &Cartridge{}

	err = cartridge.parseProgramData(file)
	if err != nil {
		return nil, err
	}

	err = cartridge.parseCharacterData(file)
	if err != nil {
		return nil, err
	}

	return cartridge, nil
}

func (cartridge *Cartridge) ReadProgramData(address uint16) uint8 {
	return cartridge.programData[address%uint16(programDataSize)]
}

func (cartridge *Cartridge) MustReadCharacterData(address uint16) uint8 {
	if address >= uint16(characterDataSize) {
		error := fmt.Sprintf("game cartridge character data invalid address: 0x%X",
			address)
		panic(error)
	}
	return cartridge.characterData[address%uint16(characterDataSize)]
}

func (cartridge *Cartridge) parseProgramData(file *os.File) error {
	_, err := file.Seek(int64(headerSize), io.SeekStart)
	if err != nil {
		return err
	}

	cartridge.programData = make([]byte, programDataSize)
	n, err := file.Read(cartridge.programData)
	if n != programDataSize {
		return errors.New("unexpected end of file")
	}
	if err != nil {
		return err
	}

	return nil
}

func (cartridge *Cartridge) parseCharacterData(file *os.File) error {
	_, err := file.Seek(int64(headerSize+programDataSize), io.SeekStart)
	if err != nil {
		return err
	}

	cartridge.characterData = make([]byte, characterDataSize)
	n, err := file.Read(cartridge.characterData)
	if n != characterDataSize {
		return errors.New("unexpected end of file")
	}
	if err != nil {
		return err
	}

	return nil
}
