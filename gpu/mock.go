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

import "fmt"

// Mock is a test double backed by a static device list.
type Mock struct {
	Devices []Metrics
	Procs   []ProcessInfo
}

var _ Collector = (*Mock)(nil)

func NewMock(devices []Metrics) *Mock { return &Mock{Devices: devices} }

func (m *Mock) All() ([]Metrics, error) { return m.Devices, nil }

func (m *Mock) ByIndex(index int) (Metrics, error) {
	if index < 0 || index >= len(m.Devices) {
		return Metrics{}, fmt.Errorf("index %d out of range [0, %d)", index, len(m.Devices))
	}
	return m.Devices[index], nil
}

func (m *Mock) ByUUID(uuid string) (Metrics, error) {
	for _, d := range m.Devices {
		if d.UUID == uuid {
			return d, nil
		}
	}
	return Metrics{}, fmt.Errorf("no device with UUID %q", uuid)
}

func (m *Mock) Processes() ([]ProcessInfo, error) { return m.Procs, nil }

func (m *Mock) Count() (int, error) { return len(m.Devices), nil }
func (m *Mock) Close() error        { return nil }
