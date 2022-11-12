package path

import (
	"reflect"
	"testing"
)

type enc struct {
	len int
	x   uint64
}

func Benchmark_Encode_split(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encode{val: 1111111111111111111, len: maxPathSegments}.split(17)
	}
}

func Benchmark_Encode_set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encode{val: 1111111111111111111, len: maxPathSegments}.set(10, 5)
	}
}

func Benchmark_Encode_new(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		useEncode(createEncode())
	}
}

func useEncode(e enc) {

}
func createEncode() enc {
	return enc{
		len: 5,
		x:   11111,
	}
}

func TestEncode_split(t *testing.T) {
	type want struct {
		l encode
		r encode
	}
	tests := []struct {
		name  string
		e     encode
		index int
		want  want
	}{
		{
			name:  "zero",
			e:     encode{},
			index: 0,
			want: want{
				l: encode{},
				r: encode{},
			},
		},
		{
			name:  "zero: index 1",
			e:     encode{},
			index: 1,
			want: want{
				l: encode{},
				r: encode{},
			},
		},
		{
			name:  "split middle",
			e:     encode{val: 123456, len: 6},
			index: 2,
			want: want{
				l: encode{val: 123, len: 3},
				r: encode{val: 456, len: 3},
			},
		},
		{
			name:  "split last",
			e:     encode{val: 123456789, len: 9},
			index: 7,
			want: want{
				l: encode{val: 12345678, len: 8},
				r: encode{val: 9, len: 1},
			},
		},
		{
			name:  "split out of bounds",
			e:     encode{val: 123456789, len: 9},
			index: 8,
			want: want{
				l: encode{val: 123456789, len: 9},
				r: encode{},
			},
		}, {
			name:  "split max: last",
			e:     encode{val: 1111111111111111111, len: 19},
			index: 17,
			want: want{
				l: encode{val: 111111111111111111, len: 18},
				r: encode{val: 1, len: 1},
			},
		}, {
			name:  "split max: first",
			e:     encode{val: 1111111111111111111, len: 19},
			index: 0,
			want: want{
				l: encode{val: 1, len: 1},
				r: encode{val: 111111111111111111, len: 18},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, r := tt.e.split(tt.index)
			got := want{
				l: l,
				r: r,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestEncode_set(t *testing.T) {
	type args struct {
		index int
		value MatchType
	}
	tests := []struct {
		name string
		e    encode
		args args
		want encode
	}{
		{
			name: "zero",
			e:    encode{},
			args: args{
				index: 0,
				value: 3,
			},
			want: encode{val: 3, len: 1},
		},
		{
			name: "zero index set",
			e:    encode{val: 12345, len: 5},
			args: args{
				index: 0,
				value: 3,
			},
			want: encode{val: 32345, len: 5},
		},
		{
			name: "last index set",
			e:    encode{val: 12345, len: 5},
			args: args{
				index: 4,
				value: 3,
			},
			want: encode{val: 12343, len: 5},
		}, {
			name: "max len: last index set",
			e:    encode{val: 1111111111111111111, len: 19},
			args: args{
				index: 18,
				value: 3,
			},
			want: encode{val: 1111111111111111113, len: 19},
		}, {
			name: "max len: first index set",
			e:    encode{val: 1111111111111111111, len: 19},
			args: args{
				index: 0,
				value: 3,
			},
			want: encode{val: 3111111111111111111, len: 19},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.set(tt.args.index, tt.args.value); got != tt.want {
				t.Errorf("set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncode_append(t *testing.T) {
	tests := []struct {
		name   string
		fields encode
		args   MatchType
		want   encode
	}{
		{
			name:   "zero: append 1",
			fields: encode{},
			args:   1,
			want:   encode{val: 1, len: 1},
		},
		{
			name:   "one: append another one",
			fields: encode{val: 1, len: 1},
			args:   2,
			want:   encode{val: 12, len: 2},
		},
		{
			name:   "max: append another one",
			fields: encode{val: 1111111111111111111, len: 19},
			args:   2,
			want:   encode{val: 1111111111111111111, len: 19},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.append(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("append() = %v, want %v", got, tt.want)
			}
		})
	}
}
