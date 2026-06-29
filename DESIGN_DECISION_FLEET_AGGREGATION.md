# Design Decision: GPU Fleet Aggregation Strategy for Multi-Node Kubernetes Clusters

| Field       | Value                          |
|-------------|--------------------------------|
| Status      | Proposed                                          |
| Date        | 2026-06-28                                        |
| Issue       | [#29](https://github.com/pmady/gpu-mcp-server/issues/29) |
| Deciders    | Pavan Madduri ([@pmady](https://github.com/pmady)) — lead maintainer |
| Reviewed by | Open for community input via GitHub issue         |

---

## Context

`gpu-mcp-server` currently runs as a DaemonSet pod on each GPU node and exposes an MCP interface for querying that node's GPUs. This design works well for single-node scenarios. In multi-node Kubernetes clusters, agents must connect to each `gpu-mcp-server` instance separately to get a full picture of fleet-wide GPU utilization.

There is no unified endpoint that:
- Lists all GPUs across all nodes
- Reports per-node GPU metrics in a single call
- Exposes fleet-level summaries (e.g. `gpu_summary`) across the cluster

This gap forces agents to implement their own discovery and fan-out logic, which is error-prone and duplicated across every consumer.

---

## Decision Drivers

- **Operational simplicity** — minimize infrastructure burden on end users
- **Reliability** — avoid introducing new single points of failure
- **Scope** — keep `gpu-mcp-server` focused; avoid scope creep
- **Kubernetes nativeness** — align with existing Kubernetes patterns where possible
- **Maintainability** — limit the complexity contributors must reason about

---

## Options Considered

### Option A: Built-in Coordinator (Aggregator Instance)

Run a dedicated `gpu-mcp-server` instance in a new `coordinator` mode. This coordinator:

1. Discovers peer `gpu-mcp-server` pods via the Kubernetes API or headless DNS
2. Fans out `list_gpus` and `get_gpu_metrics` requests to all peers in parallel
3. Merges and returns a single consolidated response
4. Tags each GPU result with the originating node identifier

**Pros:**
- Zero extra infrastructure — one MCP endpoint serves the full fleet
- Node-to-GPU attribution is first-class in responses
- `gpu_summary` aggregation works out of the box across the cluster
- Better developer experience for agents — single connection, no discovery logic

**Cons:**
- Adds Kubernetes API dependency (`client-go` or equivalent) to the server
- Coordinator is a new single point of failure for fleet-wide queries
- Fan-out requires handling partial failures, timeouts, and retries
- Significant increase in codebase complexity; harder for new contributors
- Requires new RBAC roles (`get`/`list` on Pods) for the coordinator
- At large node counts, fan-out tail latency and partial-failure handling become harder to keep correct

---

### Option B: Gateway Pattern

Keep `gpu-mcp-server` stateless and single-node. Provide and document integration with an MCP gateway layer (e.g. `agentgateway`) that distributes requests across multiple `gpu-mcp-server` instances. The project would ship a reference gateway configuration as part of the implementation (similar to existing Helm chart support), rather than leaving deployment to users to figure out from scratch.

**Pros:**
- Server remains simple, stateless, and easy to reason about
- Reuses battle-tested distributed infrastructure (routing, retries, load balancing)
- No new Kubernetes API dependency
- Scales horizontally without coordinator bottleneck
- Failure of one node's server doesn't affect other nodes' responses
- At larger cluster scale, a gateway's existing load-balancing/health-checking pays for itself instead of being reimplemented in a coordinator

**Cons:**
- Requires an additional gateway component to be deployed (project-provided config, but still a new moving part)
- Node identification metadata must be configured or injected at the gateway layer
- `gpu_summary` aggregation across nodes requires gateway-level logic or agent-side merging
- Less "plug-and-play" out of the box than a single self-contained binary

---

## Cost Comparison

| | Option A: Built-in Coordinator | Option B: Gateway Pattern |
|---|---|---|
| Code changes | New coordinator mode, K8s API client, fan-out/retry/timeout logic, RBAC | One `node` field injected via downward API (~1-2 hrs) |
| New infrastructure | None additional, but coordinator itself is new infra to run and scale | A gateway (e.g. `agentgateway`) must be deployed and operated |
| Maintenance burden | High — partial failures, peer discovery, and SPOF handling are owned by this project long-term | Low — distribution logic lives in a separately maintained gateway project |
| Time to ship | Days (new subsystem, RBAC, integration tests) | Hours of code + gateway setup time |


## Consequences

### If Option A (Built-in Coordinator) is chosen:
- A new `--mode=coordinator` flag (or similar) will be introduced
- Coordinator requires a `ServiceAccount` with Pod `list`/`get` RBAC
- Fan-out timeout and partial-failure behavior must be specified
- New integration tests covering multi-node simulation are required
- `gpu_summary` must be extended to aggregate across node responses

### If Option B (Gateway Pattern) is chosen:
- A fleet-aggregation guide will document gateway integration
- `gpu-mcp-server` response schemas must include a `node` metadata field to support gateway-level attribution
- Reference configuration for `agentgateway` (or equivalent) will be provided by the project
- No changes to core server logic are required

---

## Acceptance Criteria

The following must be satisfied before this decision is considered resolved:

- [ ] Design decision documented (built-in coordinator vs gateway) — **this document**
- [ ] Implementation or documentation of the chosen approach is merged
- [ ] `gpu_summary` works correctly across multiple nodes
- [ ] Each GPU response includes a `node` identification field (node name or UUID)

---

## References

- [pmady/gpu-mcp-server#29](https://github.com/pmady/gpu-mcp-server/issues/29) — original issue
