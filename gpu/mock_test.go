/*
Copyright 2026 The gpu-mcp-server Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gpu

import "testing"

// --- All() ---
func TestMock_All_NilDevices(t *testing.T) {
	m := NewMock(nil)
	got, err := m.All()
	if err != nil {
		t.Fatalf("All() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("All() = %v, want empty", got)
	}
}

func TestMock_All_EmptySlice(t *testing.T) {
	m := NewMock([]Metrics{})
	got, err := m.All()
	if err != nil {
		t.Fatalf("All() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("All() = %v, want empty", got)
	}
}

func TestMock_All_ReturnsAllDevices(t *testing.T) {
	devices := []Metrics{
		{Index: 0, UUID: "GPU-a"},
		{Index: 1, UUID: "GPU-b"},
	}
	m := NewMock(devices)
	got, err := m.All()
	if err != nil {
		t.Fatalf("All() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("All() returned %d devices, want 2", len(got))
	}
	if got[0].UUID != "GPU-a" || got[1].UUID != "GPU-b" {
		t.Errorf("All() = %+v, want devices in insertion order", got)
	}
}

// --- ByIndex() ---

func TestMock_ByIndex(t *testing.T) {
	devices := []Metrics{
		{Index: 0, UUID: "GPU-a"},
		{Index: 1, UUID: "GPU-b"},
	}

	tests := []struct {
		name    string
		index   int
		wantErr bool
	}{
		{name: "negative index", index: -1, wantErr: true},
		{name: "first valid index", index: 0, wantErr: false},
		{name: "last valid index", index: 1, wantErr: false},
		{name: "index equal to length", index: 2, wantErr: true},
		{name: "index far out of range", index: 99, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock(devices)
			got, err := m.ByIndex(tt.index)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ByIndex(%d) error = %v, wantErr %v", tt.index, err, tt.wantErr)
			}
			if !tt.wantErr && got.UUID != devices[tt.index].UUID {
				t.Errorf("ByIndex(%d).UUID = %q, want %q", tt.index, got.UUID, devices[tt.index].UUID)
			}
		})
	}
}

func TestMock_ByIndex_EmptyDeviceList(t *testing.T) {
	m := NewMock(nil)
	if _, err := m.ByIndex(0); err == nil {
		t.Error("ByIndex(0) on empty device list: expected error, got nil")
	}
}

// --- ByUUID() ---

func TestMock_ByUUID_Found(t *testing.T) {
	devices := []Metrics{
		{Index: 0, UUID: "GPU-a"},
		{Index: 1, UUID: "GPU-b"},
	}
	m := NewMock(devices)

	got, err := m.ByUUID("GPU-b")
	if err != nil {
		t.Fatalf(`ByUUID("GPU-b") error = %v, want nil`, err)
	}
	if got.Index != 1 || got.UUID != "GPU-b" {
		t.Errorf(`ByUUID("GPU-b") = %+v, want device at index 1`, got)
	}
}

func TestMock_ByUUID_NotFound(t *testing.T) {
	m := NewMock([]Metrics{{Index: 0, UUID: "GPU-a"}})
	if _, err := m.ByUUID("GPU-does-not-exist"); err == nil {
		t.Error("ByUUID: expected error for unknown UUID, got nil")
	}
}

func TestMock_ByUUID_EmptyString(t *testing.T) {
	m := NewMock([]Metrics{{Index: 0, UUID: "GPU-a"}})
	if _, err := m.ByUUID(""); err == nil {
		t.Error(`ByUUID(""): expected error, got nil`)
	}
}

func TestMock_ByUUID_DuplicateUUIDs(t *testing.T) {
	// Two devices deliberately share a UUID here to pin down current
	// behavior: ByUUID returns the first match rather than erroring.
	// If duplicate detection is ever added to Mock, update this test
	// to reflect the new contract.
	devices := []Metrics{
		{Index: 0, UUID: "GPU-dup", GPUUtil: 10},
		{Index: 1, UUID: "GPU-dup", GPUUtil: 90},
	}
	m := NewMock(devices)

	got, err := m.ByUUID("GPU-dup")
	if err != nil {
		t.Fatalf(`ByUUID("GPU-dup") error = %v, want nil`, err)
	}
	if got.Index != 0 {
		t.Errorf(`ByUUID("GPU-dup").Index = %d, want 0 (first match)`, got.Index)
	}
}

// --- Count() ---

func TestMock_Count(t *testing.T) {
	tests := []struct {
		name    string
		devices []Metrics
		want    int
	}{
		{name: "nil devices", devices: nil, want: 0},
		{name: "empty slice", devices: []Metrics{}, want: 0},
		{name: "two devices", devices: []Metrics{{Index: 0}, {Index: 1}}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock(tt.devices)
			got, err := m.Count()
			if err != nil {
				t.Fatalf("Count() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("Count() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Processes() ---

func TestMock_Processes_Nil(t *testing.T) {
	m := NewMock(nil)
	got, err := m.Processes()
	if err != nil {
		t.Fatalf("Processes() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Processes() = %v, want empty", got)
	}
}

func TestMock_Processes_ReturnsConfiguredProcs(t *testing.T) {
	m := NewMock(nil)
	m.Procs = []ProcessInfo{{PID: 123, Name: "python", GPUIndex: 0}}

	got, err := m.Processes()
	if err != nil {
		t.Fatalf("Processes() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].PID != 123 {
		t.Errorf("Processes() = %+v, want one process with PID 123", got)
	}
}

// --- Close() ---

func TestMock_Close(t *testing.T) {
	m := NewMock(nil)
	if err := m.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}
