//go:build cgo && linux

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

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

const maxNVLinks = 18

// NVML collects metrics through the NVIDIA Management Library.
type NVML struct {
	mu sync.Mutex
}

var _ Collector = (*NVML)(nil)

func NewNVML() (*NVML, error) {
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml init: %s", nvml.ErrorString(ret))
	}
	slog.Info("nvml initialized")
	return &NVML{}, nil
}

func (n *NVML) Close() error {
	if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
		return fmt.Errorf("nvml shutdown: %s", nvml.ErrorString(ret))
	}
	return nil
}

func (n *NVML) Count() (int, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("device count: %s", nvml.ErrorString(ret))
	}
	return count, nil
}

func (n *NVML) All() ([]Metrics, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	count, err := n.Count()
	if err != nil {
		return nil, err
	}

	var out []Metrics
	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			slog.Warn("skipping device", "index", i, "err", nvml.ErrorString(ret))
			continue
		}
		if migEnabled(dev) {
			ms, err := n.migInstances(dev, i)
			if err != nil {
				slog.Warn("mig enumeration failed", "gpu", i, "err", err)
				continue
			}
			out = append(out, ms...)
		} else {
			m, err := n.readDevice(i)
			if err != nil {
				slog.Warn("skipping device", "index", i, "err", err)
				continue
			}
			out = append(out, m)
		}
	}
	return out, nil
}

func (n *NVML) ByIndex(index int) (Metrics, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.readDevice(index)
}

func (n *NVML) ByUUID(uuid string) (Metrics, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	dev, ret := nvml.DeviceGetHandleByUUID(uuid)
	if ret != nvml.SUCCESS {
		return Metrics{}, fmt.Errorf("uuid %q: %s", uuid, nvml.ErrorString(ret))
	}
	if strings.HasPrefix(uuid, "MIG-") {
		return n.readMIGDevice(dev, -1, 0, Metrics{})
	}
	idx, ret := dev.GetIndex()
	if ret != nvml.SUCCESS {
		return Metrics{}, fmt.Errorf("index for %q: %s", uuid, nvml.ErrorString(ret))
	}
	return n.readDevice(idx)
}

func (n *NVML) readDevice(index int) (Metrics, error) {
	dev, ret := nvml.DeviceGetHandleByIndex(index)
	if ret != nvml.SUCCESS {
		return Metrics{}, fmt.Errorf("handle %d: %s", index, nvml.ErrorString(ret))
	}

	m := Metrics{Index: index}

	if v, ret := dev.GetUUID(); ret == nvml.SUCCESS {
		m.UUID = v
	}
	if v, ret := dev.GetName(); ret == nvml.SUCCESS {
		m.Name = v
	}
	if u, ret := dev.GetUtilizationRates(); ret == nvml.SUCCESS {
		m.GPUUtil = u.Gpu
		m.MemUtil = u.Memory
	}
	if mi, ret := dev.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.MemUsed = mi.Used / (1024 * 1024)
		m.MemTotal = mi.Total / (1024 * 1024)
	}
	if v, ret := dev.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		m.TempC = v
	}
	if v, ret := dev.GetPowerUsage(); ret == nvml.SUCCESS {
		m.PowerW = v / 1000
	}
	if v, ret := dev.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		m.PowerCap = v / 1000
	}
	if v, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); ret == nvml.SUCCESS {
		m.PCIeTxKBps = v
	}
	if v, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); ret == nvml.SUCCESS {
		m.PCIeRxKBps = v
	}

	var txKB, rxKB uint64
	for link := 0; link < maxNVLinks; link++ {
		tx, rx, ret := nvml.DeviceGetNvLinkUtilizationCounter(dev, link, 0)
		if ret != nvml.SUCCESS {
			continue
		}
		txKB += tx
		rxKB += rx
	}
	m.NVLinkTxMBps = txKB / 1024
	m.NVLinkRxMBps = rxKB / 1024

	return m, nil
}

// Processes enumerates the compute and graphics processes running on every
// GPU on the node. MIG instance IDs are reported through the parent device.
func (n *NVML) Processes() ([]ProcessInfo, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	count, err := n.Count()
	if err != nil {
		return nil, err
	}

	var out []ProcessInfo
	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			slog.Warn("skipping device", "index", i, "err", nvml.ErrorString(ret))
			continue
		}
		uuid, ret := dev.GetUUID()
		if ret != nvml.SUCCESS {
			uuid = ""
		}

		if procs, ret := dev.GetComputeRunningProcesses(); ret == nvml.SUCCESS {
			out = append(out, convertProcs(procs, i, uuid, "compute")...)
		} else {
			slog.Warn("compute processes failed", "gpu", i, "err", nvml.ErrorString(ret))
		}
		if procs, ret := dev.GetGraphicsRunningProcesses(); ret == nvml.SUCCESS {
			out = append(out, convertProcs(procs, i, uuid, "graphics")...)
		} else {
			slog.Warn("graphics processes failed", "gpu", i, "err", nvml.ErrorString(ret))
		}
	}
	return out, nil
}

func convertProcs(procs []nvml.ProcessInfo, index int, uuid, kind string) []ProcessInfo {
	out := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		name, ret := nvml.SystemGetProcessName(int(p.Pid))
		if ret != nvml.SUCCESS {
			name = ""
		}
		out = append(out, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			GPUIndex:   index,
			GPUUUID:    uuid,
			MemUsedMiB: p.UsedGpuMemory / (1024 * 1024),
			Type:       kind,
		})
	}
	return out
}

// --- MIG helpers ---

func migEnabled(dev nvml.Device) bool {
	mode, _, ret := dev.GetMigMode()
	return ret == nvml.SUCCESS && mode == nvml.DEVICE_MIG_ENABLE
}

func (n *NVML) migInstances(dev nvml.Device, parent int) ([]Metrics, error) {
	shared := n.readPhysical(dev, parent)
	var out []Metrics
	for i := 0; ; i++ {
		mig, ret := dev.GetMigDeviceHandleByIndex(i)
		if ret != nvml.SUCCESS {
			break
		}
		m, err := n.readMIGDevice(mig, parent, i, shared)
		if err != nil {
			slog.Warn("mig read failed", "gpu", parent, "instance", i, "err", err)
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (n *NVML) readMIGDevice(dev nvml.Device, parent, idx int, shared Metrics) (Metrics, error) {
	m := Metrics{
		Index:        idx,
		IsMIG:        true,
		ParentGPU:    parent,
		TempC:        shared.TempC,
		PowerW:       shared.PowerW,
		PowerCap:     shared.PowerCap,
		PCIeTxKBps:   shared.PCIeTxKBps,
		PCIeRxKBps:   shared.PCIeRxKBps,
		NVLinkTxMBps: shared.NVLinkTxMBps,
		NVLinkRxMBps: shared.NVLinkRxMBps,
	}
	if v, ret := dev.GetUUID(); ret == nvml.SUCCESS {
		m.UUID = v
	}
	if v, ret := dev.GetName(); ret == nvml.SUCCESS {
		m.Name = v
		if strings.HasPrefix(v, "MIG ") {
			m.MIGProfile = strings.TrimPrefix(v, "MIG ")
		}
	}
	if u, ret := dev.GetUtilizationRates(); ret == nvml.SUCCESS {
		m.GPUUtil = u.Gpu
		m.MemUtil = u.Memory
	}
	if mi, ret := dev.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.MemUsed = mi.Used / (1024 * 1024)
		m.MemTotal = mi.Total / (1024 * 1024)
	}
	return m, nil
}

func (n *NVML) readPhysical(dev nvml.Device, index int) Metrics {
	m := Metrics{Index: index}
	if v, ret := dev.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		m.TempC = v
	}
	if v, ret := dev.GetPowerUsage(); ret == nvml.SUCCESS {
		m.PowerW = v / 1000
	}
	if v, ret := dev.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		m.PowerCap = v / 1000
	}
	if v, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_TX_BYTES); ret == nvml.SUCCESS {
		m.PCIeTxKBps = v
	}
	if v, ret := dev.GetPcieThroughput(nvml.PCIE_UTIL_RX_BYTES); ret == nvml.SUCCESS {
		m.PCIeRxKBps = v
	}
	var txKB, rxKB uint64
	for link := 0; link < maxNVLinks; link++ {
		tx, rx, ret := nvml.DeviceGetNvLinkUtilizationCounter(dev, link, 0)
		if ret != nvml.SUCCESS {
			continue
		}
		txKB += tx
		rxKB += rx
	}
	m.NVLinkTxMBps = txKB / 1024
	m.NVLinkRxMBps = rxKB / 1024
	return m
}
