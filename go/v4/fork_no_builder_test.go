package ccxt

// Fork guarantee: no integration attributes the user's orders to a CCXT-owned
// builder/integrator, charges a builder fee or sets a CCXT referral code — not by
// default and not as a side effect of initialization. Everything here is offline:
// a fake http.RoundTripper answers every request and records what was sent.

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

type recordingTransport struct {
	mu     sync.Mutex
	bodies []string
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := ""
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		body = string(raw)
	}
	rt.mu.Lock()
	rt.bodies = append(rt.bodies, req.URL.String()+" "+body)
	rt.mu.Unlock()
	reply := `{"status":"ok","response":{"type":"default"}}`
	if strings.Contains(body, `"type":"order"`) {
		reply = `{"status":"ok","response":{"type":"order","data":{"statuses":[{"resting":{"oid":1}}]}}}`
	}
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(reply)),
		Request:    req,
	}, nil
}

func (rt *recordingTransport) sent() string {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return strings.Join(rt.bodies, "\n")
}

func newOfflineHyperliquid(t *testing.T) (*Hyperliquid, *recordingTransport) {
	t.Helper()
	ex := NewHyperliquid(map[string]any{
		"privateKey":    "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		"walletAddress": "0x1111111111111111111111111111111111111111",
	})
	rt := &recordingTransport{}
	ex.httpClient = &http.Client{Transport: rt}
	ex.SetMarkets([]any{map[string]any{
		"id": "0", "symbol": "BTC/USDC:USDC", "base": "BTC", "quote": "USDC", "settle": "USDC",
		"baseId": 0, "quoteId": "USDC", "settleId": "USDC", "type": "swap", "swap": true,
		"spot": false, "contract": true, "linear": true, "active": true,
		"precision": map[string]any{"amount": 0.00001, "price": 0.1},
		"limits":    map[string]any{"amount": map[string]any{}, "price": map[string]any{}, "cost": map[string]any{}},
		"info":      map[string]any{"szDecimals": "5", "maxLeverage": "50"},
	}})
	return ex, rt
}

func assertNoBuilderTraffic(t *testing.T, sent string) {
	t.Helper()
	for _, banned := range []string{"approveBuilderFee", "setReferrer", `"builder"`, "0x6530512a6c89c7cfcebc3ba7fcd9ada5f30827a6"} {
		if strings.Contains(strings.ToLower(sent), strings.ToLower(banned)) {
			t.Fatalf("request sent with %q:\n%s", banned, sent)
		}
	}
}

func TestForkHyperliquidInitializeSendsNoBuilderApprovalNorReferrer(t *testing.T) {
	ex, rt := newOfflineHyperliquid(t)
	if r := <-ex.InitializeClientAsync(); r.Err != nil {
		t.Fatalf("initializeClient: %v", r.Err)
	}
	assertNoBuilderTraffic(t, rt.sent())
}

func TestForkHyperliquidCreateOrderPayloadHasNoBuilder(t *testing.T) {
	ex, rt := newOfflineHyperliquid(t)
	price := 50000.0
	if _, err := ex.CreateOrder("BTC/USDC:USDC", "limit", "buy", 0.001, WithCreateOrderPrice(price)); err != nil {
		t.Fatalf("createOrder: %v", err)
	}
	sent := rt.sent()
	if !strings.Contains(sent, `"type":"order"`) {
		t.Fatalf("no order was sent through the fake transport:\n%s", sent)
	}
	assertNoBuilderTraffic(t, sent)
}

// Source-level guard for every integration that shipped a CCXT builder/referral by
// default: none of the hardcoded attribution identifiers may come back.
func TestForkSourceHasNoCcxtBuilderIdentifiers(t *testing.T) {
	banned := regexp.MustCompile(`(?i)builderfee|0x6530512A6c89C7cfCEbC3BA7fcD9aDa5f30827a6|0x1F5877C19e3777Cfd15F9d57253eA4aA5254Ec39|0x21d2a053495994b1132a38cd1171acec40c6741e|0xea409de8b037bb6ac664b6d12d6831b03cb04a37|"257624"|\b718718\b|"CCXT1"|"builderCode":\s*"CCXT"|"builder":\s*"4500"`)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(src), "\n") {
			// the one allowed hit: hyperliquid parseTrade adds a third-party builder's charge to the fill fee
			if path == "hyperliquid.go" && strings.Contains(line, `this.SafeString(trade, "builderFee")`) {
				continue
			}
			if banned.MatchString(line) {
				t.Errorf("%s:%d: %s", path, i+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
