# typst-web — Demo for using Typst in a Go Web Application

# Overview

This project is a minimal demonstration of how to embed the **Typst** compiler into a Go application, execute it safely at runtime, and expose it via a simple web interface — **without** runtime downloads, build-time side effects, or external orchestration tools (except ko).

The result is a single container that has the web application that compiles user-submitted Typst documents and returns them as PDFs directly to the browser. The container has a specific Typst compiler version added at build time.

# Build

Fetch the typst binaries and fonts:

```command
go generate ./cmd/server
```

# Run

```command
docker run -p 8080:8080 $(go tool ko build --local --platform=linux/arm64 ./cmd/server)
```

# FAQ

## Why aren't you embedding the typst program into the Go binary?

At runtime, we would need to lay down the `typst` binary in the container's filesystem in order to be able to execute it, anyway, same for the fonts. Including all of these at build-time into the container image with Ko has the same effect, but with fewer steps.

## Why aren't you using a multi-stage Docker build?

The official [Typst Docker image](https://github.com/typst/typst/blob/main/Dockerfile) is based on Alpine, which we really don't need for a Go program.

We also need to specify which fonts to download and then get them added to the image. The `go:generate` statements, calling `go tool` with arguments, allow us to express all of that in an idiomatic way. On top of that, Ko bundles all assets for us at container build time.
