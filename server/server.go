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
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pmady/gpu-mcp-server/gpu"
)

// -- tool input types --

type ListGPUsInput struct{}

type GetMetricsInput struct {
	Index *int   `json:"index,omitempty" jsonschema:"GPU device index (0-based)"`
	UUID  string `json:"uuid,omitempty" jsonschema:"GPU or MIG instance UUID"`
}

type SummaryInput struct{}

type GetProcessesInput struct {
	Index *int   `json:"index,omitempty" jsonschema:"filter by GPU device index (0-based)"`
	UUID  string `json:"uuid,omitempty" jsonschema:"filter by GPU or MIG instance UUID"`
}

// -- tool output types --

type GPUListItem struct {
	Index    int    `json:"index" jsonschema:"device index"`
	UUID     string `json:"uuid" jsonschema:"device UUID"`
	Name     string `json:"name" jsonschema:"GPU model name"`
	GPUUtil  uint32 `json:"gpu_utilization_percent" jsonschema:"compute utilization 0-100"`
	MemUsed  uint64 `json:"memory_used_mib" jsonschema:"memory in use"`
	MemTotal uint64 `json:"memory_total_mib" jsonschema:"total device memory"`
}

type ListGPUsOutput struct {
	Count   int           `json:"count" jsonschema:"number of GPU devices"`
	Devices []GPUListItem `json:"devices" jsonschema:"GPU device list"`
}

type SummaryOutput struct {
	DeviceCount   int     `json:"device_count" jsonschema:"number of GPUs"`
	AvgGPUUtil    float64 `json:"avg_gpu_utilization" jsonschema:"mean GPU utilization across devices"`
	AvgMemUtil    float64 `json:"avg_memory_utilization" jsonschema:"mean memory utilization across devices"`
	TotalMemUsed  uint64  `json:"total_memory_used_mib" jsonschema:"aggregate memory in use"`
	TotalMemTotal uint64  `json:"total_memory_total_mib" jsonschema:"aggregate device memory"`
	MaxTempC      uint32  `json:"max_temperature_celsius" jsonschema:"hottest GPU temperature"`
	TotalPowerW   uint32  `json:"total_power_draw_watts" jsonschema:"aggregate power consumption"`
}

type GetProcessesOutput struct {
	Count     int               `json:"count" jsonschema:"number of GPU processes"`
	Processes []gpu.ProcessInfo `json:"processes" jsonschema:"per-process GPU usage"`
}

// Handler wires GPU metrics into MCP tools.
type Handler struct {
	collector gpu.Collector
	srv       *mcp.Server
}

// New creates an MCP server and registers the GPU tools.
func New(c gpu.Collector, version string) *Handler {
	h := &Handler{
		collector: c,
		srv: mcp.NewServer(&mcp.Implementation{
			Name:    "gpu-mcp-server",
			Version: version,
		}, nil),
	}

	mcp.AddTool(h.srv, &mcp.Tool{
		Name:        "list_gpus",
		Description: "List all NVIDIA GPUs on this machine with utilization and memory info",
	}, h.listGPUs)

	mcp.AddTool(h.srv, &mcp.Tool{
		Name:        "get_gpu_metrics",
		Description: "Get detailed metrics for a specific GPU by index or UUID",
	}, h.getMetrics)

	mcp.AddTool(h.srv, &mcp.Tool{
		Name:        "gpu_summary",
		Description: "Get aggregate GPU statistics across all devices on this node",
	}, h.gpuSummary)

	mcp.AddTool(h.srv, &mcp.Tool{
		Name:        "get_gpu_processes",
		Description: "List processes using GPU resources, optionally filtered by GPU index or UUID",
	}, h.getProcesses)

	return h
}

// Run starts the MCP server over stdio. Blocks until the client disconnects.
func (h *Handler) Run(ctx context.Context) error {
	return h.srv.Run(ctx, &mcp.StdioTransport{})
}

func (h *Handler) listGPUs(ctx context.Context, req *mcp.CallToolRequest, input ListGPUsInput) (*mcp.CallToolResult, ListGPUsOutput, error) {
	devices, err := h.collector.All()
	if err != nil {
		return nil, ListGPUsOutput{}, fmt.Errorf("reading GPUs: %w", err)
	}

	out := ListGPUsOutput{
		Count:   len(devices),
		Devices: make([]GPUListItem, len(devices)),
	}
	for i, d := range devices {
		out.Devices[i] = GPUListItem{
			Index:    d.Index,
			UUID:     d.UUID,
			Name:     d.Name,
			GPUUtil:  d.GPUUtil,
			MemUsed:  d.MemUsed,
			MemTotal: d.MemTotal,
		}
	}
	return nil, out, nil
}

func (h *Handler) getMetrics(ctx context.Context, req *mcp.CallToolRequest, input GetMetricsInput) (*mcp.CallToolResult, gpu.Metrics, error) {
	switch {
	case input.UUID != "":
		m, err := h.collector.ByUUID(input.UUID)
		return nil, m, err
	case input.Index != nil:
		m, err := h.collector.ByIndex(*input.Index)
		return nil, m, err
	default:
		return nil, gpu.Metrics{}, fmt.Errorf("provide either 'index' or 'uuid'")
	}
}

func (h *Handler) getProcesses(ctx context.Context, req *mcp.CallToolRequest, input GetProcessesInput) (*mcp.CallToolResult, GetProcessesOutput, error) {
	procs, err := h.collector.Processes()
	if err != nil {
		return nil, GetProcessesOutput{}, fmt.Errorf("reading GPU processes: %w", err)
	}

	filtered := procs[:0:0]
	for _, p := range procs {
		switch {
		case input.UUID != "":
			if p.GPUUUID == input.UUID {
				filtered = append(filtered, p)
			}
		case input.Index != nil:
			if p.GPUIndex == *input.Index {
				filtered = append(filtered, p)
			}
		default:
			filtered = append(filtered, p)
		}
	}

	return nil, GetProcessesOutput{Count: len(filtered), Processes: filtered}, nil
}

func (h *Handler) gpuSummary(ctx context.Context, req *mcp.CallToolRequest, input SummaryInput) (*mcp.CallToolResult, SummaryOutput, error) {
	devices, err := h.collector.All()
	if err != nil {
		return nil, SummaryOutput{}, fmt.Errorf("reading GPUs: %w", err)
	}
	if len(devices) == 0 {
		return nil, SummaryOutput{}, nil
	}

	var out SummaryOutput
	out.DeviceCount = len(devices)
	var gpuSum, memSum uint64
	for _, d := range devices {
		gpuSum += uint64(d.GPUUtil)
		memSum += uint64(d.MemUtil)
		out.TotalMemUsed += d.MemUsed
		out.TotalMemTotal += d.MemTotal
		out.TotalPowerW += d.PowerW
		if d.TempC > out.MaxTempC {
			out.MaxTempC = d.TempC
		}
	}
	n := float64(len(devices))
	out.AvgGPUUtil = float64(gpuSum) / n
	out.AvgMemUtil = float64(memSum) / n
	return nil, out, nil
}
