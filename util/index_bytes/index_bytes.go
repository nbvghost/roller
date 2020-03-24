package index_bytes

func SetBytes(source []byte, Index uint32, Value uint8) []byte {

	p := (len(source) - 1) - int(Index)
	if p < 0 {
		source = append(source, make([]byte, -p)...)
	}
	source[Index] = Value
	return source
}
func GetBytes(source []byte, Index uint32) uint8 {
	p := (len(source) - 1) - int(Index)
	if p < 0 {
		return 0
	}
	return source[Index]
}
