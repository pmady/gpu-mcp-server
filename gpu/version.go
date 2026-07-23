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

// formatCUDAVersion converts NVML's encoded CUDA version int
// (major*1000 + minor*10) into a "major.minor" string, e.g. 12040 -> "12.4".
func formatCUDAVersion(v int) string {
	return fmt.Sprintf("%d.%d", v/1000, (v%1000)/10)
}
