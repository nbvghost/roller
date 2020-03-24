package index_bytes

import (
	"fmt"
	"testing"
)

func TestGetBytes(t *testing.T) {
	b := make([]byte, 0)
	type args struct {
		Index uint32
		Value uint8
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{name: "TestGetBytes", args: args{Index: 0, Value: 99}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b = SetBytes(b, tt.args.Index, tt.args.Value)
			fmt.Println(GetBytes(b, tt.args.Index))
		})
	}
}
