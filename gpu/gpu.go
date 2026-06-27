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

// Metrics is a point-in-time snapshot for a single GPU or MIG instance.
type Metrics struct {
	Index    int    `json:"index"`
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	GPUUtil  uint32 `json:"gpu_utilization_percent"`
	MemUtil  uint32 `json:"memory_utilization_percent"`
	MemUsed  uint64 `json:"memory_used_mib"`
	MemTotal uint64 `json:"memory_total_mib"`
	TempC    uint32 `json:"temperature_celsius"`
	PowerW   uint32 `json:"power_draw_watts"`
	PowerCap uint32 `json:"power_limit_watts"`
	// bus throughput
	PCIeTxKBps   uint32 `json:"pcie_tx_kbps"`
	PCIeRxKBps   uint32 `json:"pcie_rx_kbps"`
	NVLinkTxMBps uint64 `json:"nvlink_tx_mbps"`
	NVLinkRxMBps uint64 `json:"nvlink_rx_mbps"`
	// MIG fields — zero for non-MIG GPUs
	IsMIG      bool   `json:"is_mig,omitempty"`
	ParentGPU  int    `json:"parent_gpu,omitempty"`
	MIGProfile string `json:"mig_profile,omitempty"`
}

// ProcessInfo describes a single process using GPU resources.
type ProcessInfo struct {
	PID        uint32 `json:"pid"`
	Name       string `json:"name"`
	GPUIndex   int    `json:"gpu_index"`
	GPUUUID    string `json:"gpu_uuid,omitempty"`
	MemUsedMiB uint64 `json:"memory_used_mib"`
	// Type is "compute" or "graphics".
	Type string `json:"type"`
}

// Collector reads GPU metrics from the node.
// Implementations must be safe for concurrent use.
type Collector interface {
	All() ([]Metrics, error)
	ByIndex(index int) (Metrics, error)
	ByUUID(uuid string) (Metrics, error)
	Count() (int, error)
	// Processes returns the processes currently using each GPU on the node.
	Processes() ([]ProcessInfo, error)
	Close() error
}
