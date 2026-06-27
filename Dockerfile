FROM nvidia/cuda:13.3.0-base-ubuntu24.04 AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go ca-certificates git build-essential && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags "-X main.version=${VERSION}" -o /gpu-mcp-server ./cmd/gpu-mcp-server

FROM nvidia/cuda:13.3.0-base-ubuntu24.04
LABEL io.modelcontextprotocol.server.name="io.github.pmady/gpu-mcp-server"
COPY --from=builder /gpu-mcp-server /usr/local/bin/gpu-mcp-server
ENTRYPOINT ["gpu-mcp-server"]
