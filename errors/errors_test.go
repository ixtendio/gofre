package errors

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewErrInvalidRequestWithMessage(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with message",
			args: args{msg: "an error"},
			want: "an error",
		},
		{
			name: "without message",
			args: args{msg: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBadRequestWithMessage(tt.args.msg); got.Error() != tt.want {
				t.Errorf("NewBadRequestWithMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrInvalidRequest(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with message",
			args: args{err: errors.New("another error")},
			want: "another error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBadRequest(tt.args.err); got.Error() != tt.want {
				t.Errorf("NewBadRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrObjectNotFoundWithMessage(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with message",
			args: args{msg: "an error"},
			want: "an error",
		},
		{
			name: "without message",
			args: args{msg: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewObjectNotFoundWithMessage(tt.args.msg); got.Error() != tt.want {
				t.Errorf("NewObjectNotFoundWithMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrObjectNotFound(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want ErrObjectNotFound
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewObjectNotFound(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewObjectNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrObjectNotFound1(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with message",
			args: args{err: errors.New("another error")},
			want: "another error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewObjectNotFound(tt.args.err); got.Error() != tt.want {
				t.Errorf("NewObjectNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
