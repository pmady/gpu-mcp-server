package gpu

import "testing"

func TestFormatCUDAVersion(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "CUDA 12.4", input: 12040, want: "12.4"},
		{name: "CUDA 11.8", input: 11080, want: "11.8"},
		{name: "zero minor", input: 12000, want: "12.0"},
		{name: "CUDA 10.2", input: 10020, want: "10.2"},
		{name: "zero value", input: 0, want: "0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCUDAVersion(tt.input); got != tt.want {
				t.Errorf("formatCUDAVersion(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
