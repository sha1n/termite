package termite

import "testing"

func TestTruncateString(t *testing.T) {
	type args struct {
		s      string
		maxLen int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "below threshold", args: args{s: "hello world!", maxLen: 20}, want: "hello world!"},
		{name: "above threshold", args: args{s: "hello world!", maxLen: 6}, want: "hell.."},
		{name: "zero length", args: args{s: "hello", maxLen: 0}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateString(tt.args.s, tt.args.maxLen); got != tt.want {
				t.Errorf("TruncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}
