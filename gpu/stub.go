//go:build !cgo || !linux

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

import "errors"

// NewNVML is a stub for platforms without NVML support.
// Build on Linux with CGO_ENABLED=1 for the real implementation.
func NewNVML() (*NVML, error) {
	return nil, errors.New("NVML requires Linux with CGO enabled — use gpu.NewMock() for testing")
}

type NVML struct{}

func (n *NVML) All() ([]Metrics, error)             { return nil, errors.New("nvml unavailable") }
func (n *NVML) ByIndex(index int) (Metrics, error)  { return Metrics{}, errors.New("nvml unavailable") }
func (n *NVML) ByUUID(uuid string) (Metrics, error) { return Metrics{}, errors.New("nvml unavailable") }
func (n *NVML) Count() (int, error)                 { return 0, errors.New("nvml unavailable") }
func (n *NVML) Processes() ([]ProcessInfo, error)   { return nil, errors.New("nvml unavailable") }
func (n *NVML) Close() error                        { return nil }
