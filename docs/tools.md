# Tools Reference

## list_gpus

Enumerate all GPUs (and MIG instances) on the node.

**Input:** none

**Output:**

```json
{
  "count": 2,
  "devices": [
    {
      "index": 0,
      "uuid": "GPU-aaaa-1111",
      "name": "NVIDIA A100-SXM4-80GB",
      "gpu_utilization_percent": 85,
      "memory_used_mib": 57344,
      "memory_total_mib": 81920
    }
  ]
}
```

## get_gpu_metrics

Detailed metrics for a single GPU by index or UUID.

**Input:**

| Field | Type | Description |
|-------|------|-------------|
| `index` | int (optional) | GPU device index (0-based) |
| `uuid` | string (optional) | GPU or MIG instance UUID |

Provide one of `index` or `uuid`.

**Output:**

```json
{
  "index": 0,
  "uuid": "GPU-aaaa-1111",
  "name": "NVIDIA A100-SXM4-80GB",
  "gpu_utilization_percent": 85,
  "memory_utilization_percent": 70,
  "memory_used_mib": 57344,
  "memory_total_mib": 81920,
  "temperature_celsius": 72,
  "power_draw_watts": 300,
  "power_limit_watts": 400,
  "pcie_tx_kbps": 0,
  "pcie_rx_kbps": 0,
  "nvlink_tx_mbps": 0,
  "nvlink_rx_mbps": 0
}
```

MIG instances add `is_mig`, `parent_gpu`, and `mig_profile` fields.

## get_gpu_processes

PID-level process attribution — who is using each GPU.

**Input:**

| Field | Type | Description |
|-------|------|-------------|
| `index` | int (optional) | Filter by GPU device index |
| `uuid` | string (optional) | Filter by GPU or MIG instance UUID |

**Output:**

```json
{
  "processes": [
    {
      "pid": 12345,
      "name": "python",
      "gpu_index": 0,
      "gpu_uuid": "GPU-aaaa-1111",
      "memory_used_mib": 4096,
      "type": "compute"
    }
  ]
}
```

The `type` field is `"compute"` or `"graphics"`.

## gpu_summary

Aggregate stats across all GPUs on the node.

**Input:** none

**Output:**

```json
{
  "device_count": 2,
  "avg_gpu_utilization": 52.5,
  "avg_memory_utilization": 42.5,
  "total_memory_used_mib": 69632,
  "total_memory_total_mib": 163840,
  "max_temperature_celsius": 72,
  "total_power_draw_watts": 375
}
```
