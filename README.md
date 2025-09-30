# NES Emulator

## Stuff to address later

* Changing the NMI enable bit in PPUCTRL from 0 to 1 while the vblank flag in
PPUSTATUS is 1 will immediately trigger an NMI
* Look into returning data from palette memory immediately instead of going through
read buffer
