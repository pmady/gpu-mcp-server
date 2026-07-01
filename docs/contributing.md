# Contributing

See the full [CONTRIBUTING.md](https://github.com/pmady/gpu-mcp-server/blob/main/CONTRIBUTING.md)
on GitHub for detailed instructions.

## Quick start

```bash
git clone https://github.com/pmady/gpu-mcp-server.git
cd gpu-mcp-server
make test    # runs with no GPU (mock collector)
make lint    # golangci-lint
```

## Requirements

- All commits must be signed off (DCO): `git commit -s`
- Tests must pass: `make test`
- Linter must pass: `make lint`
- Follow existing code style (terse Go, minimal comments)

## AI-assisted contributions

See [AI_GUIDELINES.md](https://github.com/pmady/gpu-mcp-server/blob/main/AI_GUIDELINES.md).
Human verification is mandatory. Disclose AI usage in the PR description.
