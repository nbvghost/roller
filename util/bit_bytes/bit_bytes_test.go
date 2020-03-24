package bit_bytes

import (
	"encoding/hex"
	"fmt"
	"testing"
)

var b = make([]byte, 0)

func TestGetBytesState(t *testing.T) {

	type args struct {
		source   []byte
		BitIndex uint32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "TestGetBytesState", args: args{source: b, BitIndex: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBytesState(tt.args.source, tt.args.BitIndex); got != tt.want {
				t.Errorf("GetBytesState() = %v, want %v", got, tt.want)
			}
		})
	}

	fmt.Println(hex.EncodeToString(b))
}

func TestSetBytesOff(t *testing.T) {
	type args struct {
		source   []byte
		BitIndex uint32
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestSetBytesOff", args: args{source: b, BitIndex: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b = SetBytesOff(b, 0)
			fmt.Println(hex.EncodeToString(b))
		})
	}
	fmt.Println(hex.EncodeToString(b))
}

func TestSetBytesOn(t *testing.T) {

	var b = make([]byte, 0)

	type args struct {
		BitIndex uint32
	}
	tests := []struct {
		name string
		args args
	}{
		//{name:"TestSetBytesOn",args:args{BitIndex:0}},
		//{name:"TestSetBytesOn",args:args{BitIndex:5}},
		{name: "TestSetBytesOn", args: args{BitIndex: 5}},
		//{name:"TestSetBytesOn",args:args{BitIndex:23}},
		//{name:"TestSetBytesOn",args:args{BitIndex:1000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b = SetBytesOn(b, tt.args.BitIndex)
			fmt.Println(hex.EncodeToString(b))
			fmt.Println(GetBytesState(b, tt.args.BitIndex))
			b = SetBytesOff(b, tt.args.BitIndex)
			fmt.Println(GetBytesState(b, tt.args.BitIndex))
			fmt.Println(hex.EncodeToString(b), float64(len(b))/float64(1024))
		})
	}

}
