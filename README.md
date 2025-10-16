# x

A pragmatic collection of Go utilities for building backends and CLIs: auth helpers, structured errors, rate limiting, process supervision, a tiny REST stack, S3 sync, registry helpers, logging glue, Redis lock, and assorted stdlib-style helpers.

* **Module:** `github.com/spcent/x`
* **License:** MIT
* **Go:** 1.20+
* **Tags:** see releases on GitHub / pkg.go.dev. ([github.com][1])

## Table of Contents

* [Install](#install)
* [Quick Start](#quick-start)
* [Packages](#packages)

  * [auth](#auth) · [buffer](#buffer) · [client](#client) · [concurrent](#concurrent) · [config](#config) · [email](#email) · [encoding (+ proto)](#encoding--proto) · [errcode](#errcode) · [flash](#flash) · [helper](#helper) · [limiter](#limiter) · [lock](#lock) · [logging](#logging) · [middleware](#middleware) · [netutil](#netutil) · [process](#process) · [redis](#redis) · [registry (+ nacos)](#registry--nacos) · [rest](#rest) · [rpc](#rpc) · [runner](#runner) · [runtime](#runtime) · [sign](#sign) · [spinlock](#spinlock) · [testutil](#testutil) · [time](#time) · [uploader](#uploader) · [utils](#utils) · [vector](#vector) · [version](#version)
* [Versioning & Stability](#versioning--stability)
* [Contributing](#contributing)
* [License](#license)

## Install

```bash
go get github.com/spcent/x@latest
```

## Quick Start

* **Embed build info at startup**

```go
v := version.Get()
fmt.Println(v.String())
```

Populate via `-ldflags` (repo URL, tag, commit, date, etc.). ([pkg.go.dev][2])

* **Guard hot paths with a token-bucket**

```go
tb := limiter.NewTokenBucket(10, 20) // 10 tokens/sec, burst 20
if !tb.Allow() { return errors.New("rate limited") }
```

([pkg.go.dev][3])

* **Return typed errors with HTTP/gRPC mapping**

```go
e := errcode.NewError(10001, "invalid token")
httpStatus := errcode.ToHTTPStatusCode(e.Code())
rpcCode   := errcode.ToRPCCode(e.Code())
```

([pkg.go.dev][4])

* **Sync static files to S3 (with include/exclude and callbacks)**

```go
eng := uploader.NewEngine(uploader.EngineConfig{
  SaveRoot:  "./public",
  VisitHost: "https://cdn.example.com",
  Excludes:  []string{"**/*.tmp"},
}, &uploader.S3Uploader{})
eng.TailRun("./public")
```

([pkg.go.dev][5])

## Packages

### auth

Helpers around authentication concerns (audit, email helpers, JWT, simple client). Source files include `audit.go`, `auth.go`, `client.go`, `email.go`, `jwt.go`. Use as low-ceremony building blocks in your own middleware/handlers. ([pkg.go.dev][6])

### buffer

Small buffering helpers (internal quality-of-life utilities). (Listed in repository tree.) ([github.com][1])

### client

Tiny HTTP/client helpers (per repo tree). Start here when you need minimal dependencies. ([github.com][1])

### concurrent

General-purpose concurrent execution helpers.

* **API highlight:** `ConcurrentExecute(ctx, ids, task, maxConcurrent)` — fan-out tasks with bounded concurrency; aggregates failures. ([pkg.go.dev][7])

### config

Configuration loader/glue around typical env/file setups (see package page for imports & surface). ([pkg.go.dev][8])

### email

Minimal email utilities:

* `Send(to, subject, body string) error`
* `ValidateEmail(email string) bool` (basic syntax validation)

```go
ok := email.ValidateEmail("alice@example.com")
if !ok { /* reject */ }
```

([pkg.go.dev][9])

### encoding (+ proto)

Encoding codecs and gRPC codec hooks; includes a `proto` subpackage for protobuf-related glue. Designed to be thread-safe and usable from concurrent goroutines. ([pkg.go.dev][10])

### errcode

Structured application errors with code → HTTP/gRPC status mapping, plus helpers to wrap/unwrap and attach details. Good for consistent API responses and RPC interoperability. ([pkg.go.dev][4])

### flash

“Flash” message helpers for web flows (per tree). Plug into your handler stack as needed. ([github.com][1])

### helper

Assorted helpers (per tree). Use sparingly; prefer specific packages when available. ([github.com][1])

### limiter

Token-bucket rate limiter.

```go
tb := limiter.NewTokenBucket(rate /*tokens/s*/, capacity /*burst*/ )
if tb.Allow() { /* proceed */ }
```

Docs describe behavior and use-cases. ([pkg.go.dev][3])

### lock

Redis-backed distributed lock.

* `type RedisLock`
* `func NewRedisLock(rdb *redis.Client, key string, expiration time.Duration) *RedisLock`

Use for coarse critical sections across instances. ([pkg.go.dev][11])

### logging

Logging glue/utilities to complement your preferred logger (slog/zap/etc.). (Per tree.) ([github.com][1])

### middleware

HTTP/RPC middleware utilities (per tree). Pair with `auth`, `errcode`, `rest`. ([github.com][1])

### netutil

Network utilities beyond the standard `net` package (not to be confused with `x/net/netutil`). Useful for small server helpers and limits. ([pkg.go.dev][12])

### process

A small local process supervisor with a simple remote client/server API. Handy during development or for small hosts without a full init system.

**Highlights**

* `(*Cli).StartGoBin(sourcePath, name string, keepAlive bool, args []string)`
* Start/stop/restart named processes; watch and resurrect processes on failure.

```go
cli := &process.Cli{}
cli.StartGoBin("github.com/you/svc", "svc", true, []string{"-p=8080"})
```

([pkg.go.dev][13])

### redis

Redis helpers (client setup and handy wrappers). See package page. ([pkg.go.dev][14])

### registry (+ nacos)

Service registry abstraction plus a Nacos implementation under `registry/nacos` for service discovery scenarios. ([pkg.go.dev][15])

### rest

Tiny REST “stack” intended for demos and small internal tools:

* `NewServer(dataDir, tmplDir, staticDir string)` → `http.Handler`
* `NewStore(dir string)` → CSV-backed store with CRUD
* Session helpers: `SignSession`, `VerifySession`
* In-process pub/sub `Broker` with `Publish/Subscribe/Unsubscribe`
* `Schema`⇄`Record` conversion & validation

```go
srv, _ := rest.NewServer("./data", "./templates", "./static")
http.ListenAndServe(":8080", srv)
```

Browse the full index for `Store`, `Schema`, `Broker`, and auth helpers. ([pkg.go.dev][16])

### rpc

Minimal RPC helper(s) (HTTP wiring, error type). Use alongside `errcode` for consistent error surfaces. (See package page for the exported types.) ([github.com][1])

### runner

Lightweight “runner” glue for CLI/service entrypoints. (Per tree.) ([github.com][1])

### runtime

Must-style helpers for init-time error handling:

* `Must2[T1 any, T2 any](v1 T1, v2 T2, err error) (T1, T2)`

Great for compact setup code that should fail fast. ([pkg.go.dev][17])

### sign

API request signing utilities with MD5 and HMAC(SHA1) signers and a pluggable signer interface. The README (Chinese) outlines goals: variability, timeliness, uniqueness, integrity; required params: `app_id`, `timestamp`, `sign`. ([pkg.go.dev][18])

### spinlock

A small spinlock experiment (educational/low-level). Prefer higher-level sync unless you *really* need this. ([pkg.go.dev][19])

### testutil

Helpers for tests (fixtures, small assertions). (Per tree.) ([github.com][1])

### time

Human-friendly formatter that accepts tokens like `YYYY`, `MM`, `DD`, `HH`, `mm`, `ss`, and day/month names:

```go
s := xtime.Format(time.Now(), "YYYY-MM-DD HH:mm:ss")
```

Docs list all tokens and examples. ([pkg.go.dev][20])

### uploader

Sync local files to cloud storage (AWS S3 driver included). Key types: `Engine`, `EngineConfig{SaveRoot, VisitHost, ForceSync, Excludes}`, `Object{Key,ETag,FilePath,Type}`, `Syncer`, `S3Uploader` with `Upload/Delete/ListObjects`.

```go
drv := &uploader.S3Uploader{}
eng := uploader.NewEngine(conf, drv)
eng.TailRun("./public")
```

Designed for “push on change” static hosting flows. ([pkg.go.dev][5])

### utils

Grab-bag of micro-helpers (per tree). Keep your imports focused; prefer dedicated packages above. ([github.com][1])

### vector

Small vector/math helpers (per tree). Useful in sims or scoring utilities. ([github.com][1])

### version

Collect and print build/VCS metadata (tag, commit, tree state, build date, repo URL). Pairs nicely with `-ldflags` in CI. ([pkg.go.dev][2])

## Versioning & Stability

The module ships tagged releases (e.g., `v1.1.2`) and individual packages on pkg.go.dev show stability badges. Pin a tag in your `go.mod` for reproducible builds. ([pkg.go.dev][2])

## Contributing

* Keep packages small and focused.
* Favor stdlib patterns (context, errors, io) and zero magic.
* Add tests and brief package docs; public identifiers should be documented.
* Avoid introducing heavy dependencies unless essential.

## License

MIT – see [`LICENSE`](./LICENSE). ([github.com][1])

---

### Notes & Sources

Most API details above come from the package pages on pkg.go.dev (they reflect the exported symbols and docs from this repo): `version`, `rest`, `uploader`, `errcode`, `limiter`, `process`, `auth`, `config`, `time`, `lock`, `encoding` (+ `encoding/proto`), `registry` (+ `registry/nacos`), `netutil`. ([pkg.go.dev][2])

[1]: https://github.com/spcent/x "GitHub - spcent/x: golang libs, include db, cache, file, rpc, command .etc"
[2]: https://pkg.go.dev/github.com/spcent/x/version "version package - github.com/spcent/x/version - Go Packages"
[3]: https://pkg.go.dev/github.com/spcent/x/limiter "limiter package - github.com/spcent/x ..."
[4]: https://pkg.go.dev/github.com/spcent/x/errcode "errcode package - github.com/spcent/x/ ..."
[5]: https://pkg.go.dev/github.com/spcent/x/uploader "uploader package - github.com/spcent/x/uploader - Go Packages"
[6]: https://pkg.go.dev/github.com/spcent/x/auth "auth package - github.com/spcent/x/ ..."
[7]: https://pkg.go.dev/github.com/spcent/x/concurrent "concurrent package - github.com/spcent/x/ ..."
[8]: https://pkg.go.dev/github.com/spcent/x/config "config package - github.com/spcent/x/ ..."
[9]: https://pkg.go.dev/github.com/spcent/x/email "email package - github.com/spcent/x ..."
[10]: https://pkg.go.dev/github.com/spcent/x/encoding "encoding package - github.com/spcent/x/ ..."
[11]: https://pkg.go.dev/github.com/spcent/x/lock "lock package - github.com/spcent/x/ ..."
[12]: https://pkg.go.dev/github.com/spcent/x/netutil "netutil package - github.com/spcent/x/ ..."
[13]: https://pkg.go.dev/github.com/spcent/x/process "process package - github.com/spcent/x ..."
[14]: https://pkg.go.dev/github.com/spcent/x/redis "redis package - github.com/spcent/x/ ..."
[15]: https://pkg.go.dev/github.com/spcent/x/registry "registry package - github.com/spcent/x/ ..."
[16]: https://pkg.go.dev/github.com/spcent/x/rest "rest package - github.com/spcent/x/rest - Go Packages"
[17]: https://pkg.go.dev/github.com/spcent/x/runtime "runtime package - github.com/spcent/x/ ..."
[18]: https://pkg.go.dev/github.com/spcent/x/sign "sign package - github.com/spcent/x/ ..."
[19]: https://pkg.go.dev/github.com/spcent/x/spinlock "spinlock package - github.com/spcent/x ..."
[20]: https://pkg.go.dev/github.com/spcent/x/time "time package - github.com/spcent/x/time - ..."
