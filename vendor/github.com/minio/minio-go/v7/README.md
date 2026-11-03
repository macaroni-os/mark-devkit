<h1 align="center">
  <a href="https://silo.pgsty.com/">
    <img src=".github/silo-logo.svg" alt="Silo" width="240">
  </a>
</h1>

<h1 align="center">Silo Go SDK</h1>

<p align="center">
  <strong>Go client SDK for Silo and Amazon S3-compatible object storage</strong>
</p>

<p align="center">
  <a href="https://silo.pgsty.com/">Website</a> ·
  <a href="https://silo.pgsty.com/developers/go/minio-go/">Documentation</a> ·
  <a href="https://pkg.go.dev/github.com/minio/minio-go/v7">API Reference</a> ·
  <a href="https://github.com/pgsty/silo-go/releases">Releases</a> ·
  <a href="https://github.com/pgsty/silo-go/tree/master/examples">Examples</a> ·
  <a href="https://silo.pgsty.com/about/security/">Security</a>
</p>

<p align="center">
  <a href="https://silo.pgsty.com/"><img alt="Website" src="https://img.shields.io/badge/Website-silo.pgsty.com-1d588c"></a>
  <a href="https://github.com/pgsty/silo-go/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/pgsty/silo-go?include_prereleases&label=release&logo=github"></a>
  <a href="go.mod"><img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/pgsty/silo-go?logo=go"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/badge/license-Apache%202.0-blue"></a>
</p>

> [!IMPORTANT]
> `pgsty/silo-go` is an independent, community-maintained fork of the open-source MinIO Go SDK, published by [Pigsty](https://pigsty.io/) for [Silo](https://silo.pgsty.com/). It is not affiliated with, endorsed by, or sponsored by MinIO, Inc. “MinIO” is used only to identify the upstream project, retained compatibility interfaces, and project lineage.

## Overview

Silo Go SDK provides the Go client used by Silo, `mcli`, Silo Console, and other applications that connect to Amazon S3-compatible object storage. It supports bucket and object operations, multipart transfers, presigned requests, notifications, encryption, lifecycle rules, retention, replication, and S3-compatible authentication.

This fork follows the same compatibility rule as the Silo server: **the project and its delivery surface are renamed; public interfaces stay stable.** Existing Go source continues to import `github.com/minio/minio-go/v7` and use the `minio` package name.

## Compatibility Contract

| Surface | Current contract |
| :-- | :-- |
| Project repository | [`github.com/pgsty/silo-go`](https://github.com/pgsty/silo-go) |
| Go module and imports | `github.com/minio/minio-go/v7` |
| Go package name | `minio` |
| Silo fork selection | `replace github.com/minio/minio-go/v7 => github.com/pgsty/silo-go/v7 v7.3.1` |
| Wire protocol | Amazon S3-compatible API and AWS Signature V2/V4 |
| User-Agent compatibility | The existing `MinIO (...) minio-go/v7.x` form is retained |

The module path, import paths, exported Go identifiers, environment variables, protocol headers, and User-Agent form are compatibility interfaces, not product branding. Changing them would split the Go package ecosystem or break integrations, so this release deliberately preserves them.

## Install the Silo Fork

Select the maintained fork in the application module while keeping existing imports unchanged:

```go
require github.com/minio/minio-go/v7 v7.3.1

replace github.com/minio/minio-go/v7 => github.com/pgsty/silo-go/v7 v7.3.1
```

Or apply the same selection from the command line:

```bash
go mod edit -require=github.com/minio/minio-go/v7@v7.3.1
go mod edit -replace=github.com/minio/minio-go/v7=github.com/pgsty/silo-go/v7@v7.3.1
go mod tidy
```

The replacement repository deliberately retains the original module declaration so Go source, exported types, and import identity stay compatible. Go applies its code under the left-hand module identity shown above. Go also ignores `replace` directives declared by dependency modules, so applications and top-level services that require the fork must carry this replacement in their own `go.mod`.

## Quick Start

The compatibility import path and `minio` package name remain unchanged:

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	client, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds: credentials.NewStaticV4(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			"",
		),
		Secure: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := client.MakeBucket(ctx, "demo", minio.MakeBucketOptions{}); err != nil {
		exists, checkErr := client.BucketExists(ctx, "demo")
		if checkErr != nil || !exists {
			log.Fatal(err)
		}
	}

	_, err = client.FPutObject(ctx, "demo", "README.md", "README.md", minio.PutObjectOptions{
		ContentType: "text/markdown",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

Run it against a local Silo server:

```bash
export S3_ACCESS_KEY=minioadmin
export S3_SECRET_KEY=change-me-long-password
go run .
```

Use TLS and scoped credentials in production. See the [Go client guide](https://silo.pgsty.com/developers/go/minio-go/), compatibility-path [API reference](https://pkg.go.dev/github.com/minio/minio-go/v7), and maintained [`examples`](examples/) for the complete surface.

## Development

```bash
go test ./...
make lint
```

`make lint` runs the pinned GolangCI-Lint release without adding its analyzer dependency tree to the SDK's published `go.mod`. A cold lint run needs network access or a pre-warmed Go module cache. Functional tests require a compatible object-storage server and are documented in [`CLAUDE.md`](CLAUDE.md).

## Related Projects

| Repository | Description |
| :-- | :-- |
| [`pgsty/silo`](https://github.com/pgsty/silo) | Silo object-storage server |
| [`pgsty/mc`](https://github.com/pgsty/mc) | Silo command-line client, shipped as `mcli` |
| [`pgsty/silo-console`](https://github.com/pgsty/silo-console) | Web administration console for Silo |
| [`pgsty/silo-pkg`](https://github.com/pgsty/silo-pkg) | Shared Go packages maintained for the Silo forks |

## License, Attribution, and Trademark

Silo Go SDK is distributed under the [Apache License 2.0](LICENSE). This fork derives from [`minio/minio-go`](https://github.com/minio/minio-go); upstream copyright and third-party notices remain preserved in [`NOTICE`](NOTICE), [`CREDITS`](CREDITS), source headers, and Git history.

MinIO is a trademark of MinIO, Inc. The name is used here only to identify the upstream project and compatibility lineage. Pigsty, Silo, and `pgsty/silo-go` are independent community efforts and are not affiliated with, endorsed by, or sponsored by MinIO, Inc.
