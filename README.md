# typst-web — Demo for Embedding Typst into a Go Web Application

# Build

Fetch the typst binaries and fonts:

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
