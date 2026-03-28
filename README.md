# typst-web — Demo for using Typst in a Go Web Application

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

The result is a single container that has the web application that compiles user-submitted Typst documents and returns them as PDFs directly to the browser. The container has a specific Typst compiler version added at build time.
