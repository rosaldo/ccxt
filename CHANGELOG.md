# Changelog

> Generated from git by `scripts/gen-changelog.sh` — do not edit by hand.

- **v5.0.1** (2026-10-28) · `test` — remove upstream brokerId suite (tests/base, --idTests) that asserted the old CCXT partner ids; neutralize partner prefix in tests/types orders fixture
- **v5.0.0** (2026-10-28) · `major` — Go module moves to the repository root as github.com/rosaldo/ccxt/v5 (was go/v4); tags are now vX.Y.Z
- **v4.5.95** (2026-10-28) · `fix` — remove every CCXT partner id: affiliate links, broker ids, partner headers and client order id prefixes
- **v4.5.94** (2026-10-28) · `docs` — README: install with go get github.com/rosaldo/ccxt/go/v4, usage, credits
- **v4.5.93** (2026-10-28) · `fix` — finish the module rename: protobuf go_package descriptors and the cli build
- **v4.5.92** (2026-10-28) · `test` — gate also runs the offline base suites (go/tests --baseTests, REST and WS)
- **v4.5.91** (2026-10-28) · `fix` — remove CCXT builder fees, builder attribution and referral codes from 9 integrations
- **v4.5.90** (2026-10-28) · `chore` — Go-only fork cut from upstream ccxt v4.5.90
