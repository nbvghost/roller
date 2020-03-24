package bit_bytes

//4294967295
func SetBytesOn(source []byte, BitIndex uint32) []byte {

	byteIndex := uint32(BitIndex / 8)
	byteBitIndex := uint8(BitIndex % 8)
	p := (len(source) - 1) - int(byteIndex)
	if p < 0 {
		source = append(source, make([]byte, -p)...)
	}
	byte := uint8(source[byteIndex])
	//11101011
	//   10000
	byte = byte | (1 << byteBitIndex)
	source[byteIndex] = byte

	return source
}
func SetBytesOff(source []byte, BitIndex uint32) []byte {
	byteIndex := uint32(BitIndex / 8)
	byteBitIndex := uint8(BitIndex % 8)
	p := (len(source) - 1) - int(byteIndex)
	if p < 0 {
		source = append(source, make([]byte, -p)...)
	}
	byte := uint8(source[byteIndex])
	//11101011
	//   10000
	byte = byte | (1 << byteBitIndex)
	//11111011
	//   10000
	byte = byte ^ (1 << byteBitIndex)
	source[byteIndex] = byte
	return source
}
func GetBytesState(source []byte, BitIndex uint32) bool {
	byteIndex := uint32(BitIndex / 8)
	byteBitIndex := uint8(BitIndex % 8)
	p := (len(source) - 1) - int(byteIndex)
	if p < 0 {
		return false
	}
	byte := uint8(source[byteIndex])

	if byte<<(7-byteBitIndex)>>7 == 0 {
		return false
	} else {
		return true
	}

}
