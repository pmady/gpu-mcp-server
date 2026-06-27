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

package server

import (
	"context"
	"testing"

	"github.com/pmady/gpu-mcp-server/gpu"
)

var testDevices = []gpu.Metrics{
	{
		Index:    0,
		UUID:     "GPU-aaaa-1111",
		Name:     "NVIDIA A100-SXM4-80GB",
		GPUUtil:  85,
		MemUtil:  70,
		MemUsed:  57344,
		MemTotal: 81920,
		TempC:    72,
		PowerW:   300,
		PowerCap: 400,
	},
	{
		Index:    1,
		UUID:     "GPU-bbbb-2222",
		Name:     "NVIDIA A100-SXM4-80GB",
		GPUUtil:  20,
		MemUtil:  15,
		MemUsed:  12288,
		MemTotal: 81920,
		TempC:    38,
		PowerW:   75,
		PowerCap: 400,
	},
}

var testProcs = []gpu.ProcessInfo{
	{PID: 12345, Name: "python", GPUIndex: 0, GPUUUID: "GPU-aaaa-1111", MemUsedMiB: 4096, Type: "compute"},
	{PID: 67890, Name: "tritonserver", GPUIndex: 1, GPUUUID: "GPU-bbbb-2222", MemUsedMiB: 16384, Type: "compute"},
	{PID: 222, Name: "Xorg", GPUIndex: 0, GPUUUID: "GPU-aaaa-1111", MemUsedMiB: 64, Type: "graphics"},
}

func newTestHandler() *Handler {
	m := gpu.NewMock(testDevices)
	m.Procs = testProcs
	return New(m, "test")
}

func TestListGPUs(t *testing.T) {
	h := newTestHandler()
	_, out, err := h.listGPUs(context.Background(), nil, ListGPUsInput{})
	if err != nil {
		t.Fatalf("listGPUs: %v", err)
	}
	if out.Count != 2 {
		t.Errorf("count = %d, want 2", out.Count)
	}
	if out.Devices[0].UUID != "GPU-aaaa-1111" {
		t.Errorf("devices[0].UUID = %q, want GPU-aaaa-1111", out.Devices[0].UUID)
	}
	if out.Devices[1].GPUUtil != 20 {
		t.Errorf("devices[1].GPUUtil = %d, want 20", out.Devices[1].GPUUtil)
	}
}

func TestListGPUs_Empty(t *testing.T) {
	h := New(gpu.NewMock(nil), "test")
	_, out, err := h.listGPUs(context.Background(), nil, ListGPUsInput{})
	if err != nil {
		t.Fatalf("listGPUs: %v", err)
	}
	if out.Count != 0 {
		t.Errorf("count = %d, want 0", out.Count)
	}
}

func TestGetMetrics_ByIndex(t *testing.T) {
	h := newTestHandler()
	idx := 0
	_, out, err := h.getMetrics(context.Background(), nil, GetMetricsInput{Index: &idx})
	if err != nil {
		t.Fatalf("getMetrics: %v", err)
	}
	if out.UUID != "GPU-aaaa-1111" {
		t.Errorf("UUID = %q, want GPU-aaaa-1111", out.UUID)
	}
	if out.TempC != 72 {
		t.Errorf("TempC = %d, want 72", out.TempC)
	}
}

func TestGetMetrics_ByUUID(t *testing.T) {
	h := newTestHandler()
	_, out, err := h.getMetrics(context.Background(), nil, GetMetricsInput{UUID: "GPU-bbbb-2222"})
	if err != nil {
		t.Fatalf("getMetrics: %v", err)
	}
	if out.GPUUtil != 20 {
		t.Errorf("GPUUtil = %d, want 20", out.GPUUtil)
	}
}

func TestGetMetrics_NoInput(t *testing.T) {
	h := newTestHandler()
	_, _, err := h.getMetrics(context.Background(), nil, GetMetricsInput{})
	if err == nil {
		t.Error("expected error when neither index nor uuid provided")
	}
}

func TestGetMetrics_BadIndex(t *testing.T) {
	h := newTestHandler()
	idx := 99
	_, _, err := h.getMetrics(context.Background(), nil, GetMetricsInput{Index: &idx})
	if err == nil {
		t.Error("expected error for out-of-range index")
	}
}

func TestGetMetrics_BadUUID(t *testing.T) {
	h := newTestHandler()
	_, _, err := h.getMetrics(context.Background(), nil, GetMetricsInput{UUID: "GPU-nope"})
	if err == nil {
		t.Error("expected error for unknown UUID")
	}
}

func TestGPUSummary(t *testing.T) {
	h := newTestHandler()
	_, out, err := h.gpuSummary(context.Background(), nil, SummaryInput{})
	if err != nil {
		t.Fatalf("gpuSummary: %v", err)
	}
	if out.DeviceCount != 2 {
		t.Errorf("DeviceCount = %d, want 2", out.DeviceCount)
	}
	// avg GPU util should be (85+20)/2 = 52.5
	if out.AvgGPUUtil < 52.0 || out.AvgGPUUtil > 53.0 {
		t.Errorf("AvgGPUUtil = %f, want ~52.5", out.AvgGPUUtil)
	}
	if out.MaxTempC != 72 {
		t.Errorf("MaxTempC = %d, want 72", out.MaxTempC)
	}
	if out.TotalPowerW != 375 {
		t.Errorf("TotalPowerW = %d, want 375", out.TotalPowerW)
	}
	wantMemUsed := uint64(57344 + 12288)
	if out.TotalMemUsed != wantMemUsed {
		t.Errorf("TotalMemUsed = %d, want %d", out.TotalMemUsed, wantMemUsed)
	}
}

func TestGetProcesses_All(t *testing.T) {
	h := newTestHandler()
	_, out, err := h.getProcesses(context.Background(), nil, GetProcessesInput{})
	if err != nil {
		t.Fatalf("getProcesses: %v", err)
	}
	if out.Count != 3 {
		t.Errorf("count = %d, want 3", out.Count)
	}
}

func TestGetProcesses_FilterByIndex(t *testing.T) {
	h := newTestHandler()
	idx := 0
	_, out, err := h.getProcesses(context.Background(), nil, GetProcessesInput{Index: &idx})
	if err != nil {
		t.Fatalf("getProcesses: %v", err)
	}
	if out.Count != 2 {
		t.Errorf("count = %d, want 2 for index 0", out.Count)
	}
	for _, p := range out.Processes {
		if p.GPUIndex != 0 {
			t.Errorf("process %d has index %d, want 0", p.PID, p.GPUIndex)
		}
	}
}

func TestGetProcesses_FilterByUUID(t *testing.T) {
	h := newTestHandler()
	_, out, err := h.getProcesses(context.Background(), nil, GetProcessesInput{UUID: "GPU-bbbb-2222"})
	if err != nil {
		t.Fatalf("getProcesses: %v", err)
	}
	if out.Count != 1 {
		t.Fatalf("count = %d, want 1", out.Count)
	}
	if out.Processes[0].PID != 67890 {
		t.Errorf("PID = %d, want 67890", out.Processes[0].PID)
	}
}

func TestGetProcesses_Empty(t *testing.T) {
	h := New(gpu.NewMock(nil), "test")
	_, out, err := h.getProcesses(context.Background(), nil, GetProcessesInput{})
	if err != nil {
		t.Fatalf("getProcesses: %v", err)
	}
	if out.Count != 0 {
		t.Errorf("count = %d, want 0", out.Count)
	}
}

func TestGPUSummary_Empty(t *testing.T) {
	h := New(gpu.NewMock(nil), "test")
	_, out, err := h.gpuSummary(context.Background(), nil, SummaryInput{})
	if err != nil {
		t.Fatalf("gpuSummary: %v", err)
	}
	if out.DeviceCount != 0 {
		t.Errorf("DeviceCount = %d, want 0", out.DeviceCount)
	}
}
