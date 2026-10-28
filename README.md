# ccxt (Go-only fork)

A Go-only fork of [CCXT](https://github.com/ccxt/ccxt), cut from upstream tag
`v4.5.90` with a fresh history.

## Why

- **Go only.** Upstream ships one repository for many languages (about 8.5 GB of
  history). This fork keeps just the Go module, so it is small and fast to clone.
- **No CCXT fees or attribution.** Upstream turns on a "builder fee" by default in
  several DEX integrations (hyperliquid, aster, extended, grvt, lighter, nado,
  pacifica, plus the hyperliquid and polymarket prediction markets): it approves a
  CCXT-owned builder, attaches it to every order and sets a CCXT referral code.
  This fork removes all of that; `go/v4/fork_no_builder_test.go` keeps it out.
- **No binaries and no AI-agent files** in the tree.

## Install

```bash
go get github.com/rosaldo/ccxt/go/v4@v4.5.94
```

The module lives in `go/v4`, so version `v4.X.Y` is the git tag `go/v4.X.Y`
([go.dev/ref/mod](https://go.dev/ref/mod#vcs-version)). `@latest` picks the newest
tag; `VERSION` holds the current one and `CHANGELOG.md` lists every release.

```go
package main

import (
	"fmt"

	ccxt "github.com/rosaldo/ccxt/go/v4"
)

func main() {
	exchange := ccxt.NewBinance(nil)
	ticker, err := exchange.FetchTicker("BTC/USDT")
	if err != nil {
		panic(err)
	}
	fmt.Println(*ticker.Last)
}
```

Sub-packages: `github.com/rosaldo/ccxt/go/v4/pro` (WebSocket) and
`github.com/rosaldo/ccxt/go/v4/prediction` (prediction markets).

## Develop

```bash
scripts/commit.sh <type> "<message>"   # gate (build, vet, offline tests) + bump + tag
scripts/push.sh                         # owner only
```

## Credits and license

All exchange code comes from [CCXT](https://github.com/ccxt/ccxt) by Igor Kroitor and
the CCXT contributors; the API manual is the upstream
[wiki](https://github.com/ccxt/ccxt/wiki/Manual). MIT, as upstream — see
[LICENSE](LICENSE).
