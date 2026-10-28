# ccxt (Go-only fork)

A Go-only fork of [CCXT](https://github.com/ccxt/ccxt), cut from upstream tag
`v4.5.90` with a fresh history.

## Why

- **Go only.** Upstream ships one repository for many languages (about 8.5 GB of
  history). This fork keeps just the Go module, now at the repository root.
- **No CCXT fees, attribution or partner ids.** Upstream turns on a "builder fee" in
  several DEX integrations, sends CCXT broker/partner ids (params, headers, client
  order id prefixes such as Binance's `x-…`) and ships CCXT affiliate links. This fork
  removes all of it; `fork_no_builder_test.go` and `fork_no_partner_test.go` keep it
  out. A broker id is only sent if you set one yourself in `options`.
- **No binaries and no AI-agent files** in the tree.

## Install

```bash
go get github.com/rosaldo/ccxt/v5@v5.0.0
```

v5 is the first major of the fork (module path `github.com/rosaldo/ccxt/v5`, tags
`v5.X.Y`); it replaces the upstream-style `go/v4` layout. `VERSION` holds the current
version and `CHANGELOG.md` lists every release.

```go
package main

import (
	"fmt"

	ccxt "github.com/rosaldo/ccxt/v5"
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

Sub-packages: `github.com/rosaldo/ccxt/v5/pro` (WebSocket) and
`github.com/rosaldo/ccxt/v5/prediction` (prediction markets). `cli/` and `tests/` are
separate local modules (a CLI and upstream's offline test runner).

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
