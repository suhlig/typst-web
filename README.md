# typst-web — Embedded Typst Web Compiler (Go)

# Build

Fetch the typst binaries:

```command
go generate ./cmd/server
```

# Run

```command
docker run -p 8080:8080 $(ko build --local --platform=linux/arm64 ./cmd/server)
```

# Overview

This project is a minimal demonstration of how to embed the **Typst** compiler into a Go application, execute it safely at runtime, and expose it via a simple web interface — **without** runtime downloads, build-time side effects, or external orchestration tools.

The result is a single Linux Go binary that:

- contains a specific Typst compiler version
- extracts it lazily at runtime
- compiles user-submitted Typst documents
- returns PDFs directly to the browser

## What This Project Does

- `GET /`
  - Serves a minimal HTML page with a large textarea for Typst input
- `POST /`
  - Receives Typst source
  - Compiles it using an embedded Typst binary
  - Returns the resulting PDF (`application/pdf`)

## Architecture Brief (Core Idea)

> Typst is treated as *source material*, not a dependency.

- It is fetched explicitly
- Embedded immutably
- Extracted ephemerally
- Executed as a disposable cache artifact

### Invariants

- `go build` never downloads anything
- Runtime never contacts the network
- Execution failures trigger regeneration
- The filesystem is never trusted for persistence

## Build-Time Phase: `go generate`

Prepare **all mutable inputs** before compilation:

- `internal/typstmeta/version.go` defines:
  - Typst version
  - Expected SHA256 digest
- `go generate` runs `cmd/typst-fetch`
- The generator:
  - Downloads the official Typst Linux binary
  - Verifies its checksum
  - Writes it atomically to `internal/embedded/typst/typst`

## Runtime Phase: Embedded Execution

At runtime, the server:

1. Writes the embedded Typst binary to:
   ```
   ~/.cache/typstweb/typst
   ```
2. Uses atomic write + rename
3. Sets executable permissions

The execution strategy is failure-driven:

- Attempt to execute Typst
- If execution fails:
  - Re-extract the binary
  - Retry once
- Second failure propagates

The cache directory is treated as disposable.

- It may be deleted at any time
- The program assumes nothing persists
- Deleting a running executable is safe on Linux (inode semantics)

This collapses GC, integrity checks, and versioning into one rule:

> If it runs, use it.
> If it doesn’t, regenerate it.

## Pros

- Deterministic builds
- No runtime downloads
- No Makefiles or shell scripts
- No IPC wrappers around Typst
- Safe under crashes, restarts, and concurrency
- Works with cross-compilation
- Minimal moving parts

## Cons and Trade-offs

- Embedded binary increases final executable size
- Git history grows over time (one binary per update)
- First execution pays extraction cost
- Linux-specific assumptions (inode semantics)

## Alternatives Considered

### Fetching During `go build`

- Breaks reproducibility
- Violates Go build guarantees
- Breaks build caching

### Runtime Download of Typst

- Network dependency
- Supply-chain risk
- Operational fragility

### Typst as a Separate Service / IPC Wrapper

- Additional moving parts
- More failure modes
- Harder deployment

### Binary-in-Go-Module Dependency

- Abuses module system
- Poor long-term scaling
- Obscures provenance
