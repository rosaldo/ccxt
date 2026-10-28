# ccxt (Go-only fork)

A Go-only fork of [CCXT](https://github.com/ccxt/ccxt), cut from upstream tag
`v4.5.90` with a fresh history.

## Why

- **Go only.** Upstream ships one repository for eight languages (about 8.5 GB of
  history). This fork keeps just the Go module, so it is small and fast to clone.
- **No hidden fees or attribution.** Upstream enables a "builder fee" by default on
  several DEX integrations (approving a CCXT-owned builder address, attaching it to
  every order and setting a CCXT referral code). This fork removes that.
- **No binaries and no AI-agent files** in the tree.

## Use

```bash
go get github.com/rosaldo/ccxt/go/v4@latest
```

```go
import ccxt "github.com/rosaldo/ccxt/go/v4"
```

Versions are tagged `go/vX.Y.Z` (the module lives in `go/v4`; see
[go.dev/ref/mod](https://go.dev/ref/mod#vcs-version)). `VERSION` holds the current
one; `CHANGELOG.md` lists the fork's releases.

## Develop

```bash
scripts/commit.sh <type> "<message>"   # gate (build, vet, offline tests) + bump + tag
scripts/push.sh                         # owner only
```

## License

MIT, as upstream — see [LICENSE](LICENSE). Copyright of the original code belongs
to the CCXT authors.
