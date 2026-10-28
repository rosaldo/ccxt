package ccxt

// Fork guarantee: no CCXT partner/broker id is sent by default — no prefix in the
// generated client order id, no broker param, tag or header. Offline (fake transport).

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
)

func testMarket(id, symbol, base, quote, kind string) map[string]any {
	return map[string]any{
		"id": id, "symbol": symbol, "base": base, "quote": quote, "baseId": base, "quoteId": quote,
		"type": kind, "spot": kind == "spot", "swap": false, "future": false, "option": false, "margin": false,
		"contract": false, "active": true, "precision": map[string]any{"amount": 0.00001, "price": 0.01},
		"limits": map[string]any{"amount": map[string]any{}, "price": map[string]any{}, "cost": map[string]any{}},
		"info":   map[string]any{"orderTypes": []any{"LIMIT", "MARKET"}, "instType": "SPOT"},
	}
}

func TestForkBinanceGeneratedClientOrderIdHasNoPartnerPrefix(t *testing.T) {
	ex := NewBinance(map[string]any{"apiKey": "key", "secret": "secret"})
	rt := &recordingTransport{reply: `{"symbol":"BTCUSDT","orderId":1,"clientOrderId":"x","status":"NEW","type":"LIMIT","side":"BUY"}`}
	ex.httpClient = &http.Client{Transport: rt}
	ex.SetMarkets([]any{testMarket("BTCUSDT", "BTC/USDT", "BTC", "USDT", "spot")})
	if _, err := ex.CreateOrder("BTC/USDT", "limit", "buy", 0.001, WithCreateOrderPrice(50000)); err != nil {
		t.Fatalf("createOrder: %v", err)
	}
	sent := rt.sent()
	m := regexp.MustCompile(`newClientOrderId=([^&\s]+)`).FindStringSubmatch(sent)
	if m == nil {
		t.Fatalf("no order sent:\n%s", sent)
	}
	if strings.HasPrefix(m[1], "x-") {
		t.Fatalf("generated client order id carries a partner prefix: %s", m[1])
	}
}

func TestForkOkxOrderHasNoBrokerTagNorClientIdPrefix(t *testing.T) {
	ex := NewOkx(map[string]any{"apiKey": "key", "secret": "secret", "password": "pass"})
	rt := &recordingTransport{reply: `{"code":"0","msg":"","data":[{"ordId":"1","clOrdId":"","sCode":"0","sMsg":""}]}`}
	ex.httpClient = &http.Client{Transport: rt}
	ex.SetMarkets([]any{testMarket("BTC-USDT", "BTC/USDT", "BTC", "USDT", "spot")})
	if _, err := ex.CreateOrder("BTC/USDT", "limit", "buy", 0.001, WithCreateOrderPrice(50000)); err != nil {
		t.Fatalf("createOrder: %v", err)
	}
	sent := rt.sent()
	if !strings.Contains(sent, "/api/v5/trade/") {
		t.Fatalf("no order sent:\n%s", sent)
	}
	for _, banned := range []string{`"tag"`, `"clOrdId"`, "6b9ad766b55dBCDE"} {
		if strings.Contains(sent, banned) {
			t.Fatalf("order sent with %s:\n%s", banned, sent)
		}
	}
}

// Source-level guard: none of the partner ids, broker tags or affiliate links upstream
// shipped by default may come back.
func TestForkSourceHasNoCcxtPartnerIdentifiers(t *testing.T) {
	banned := regexp.MustCompile(`x-TKT5PX2F|x-cvBPrNm9|x-xcKtGhcu|x-R4BD3S82|"6956"|ec6dd3a7dd982d0b|nfqkvdjp|10000700011|AA03022abc|9e58cc35-5b5e-4133-92ec-166e3f077cb8|"ccxtfutures"|CCXTMODE|42ce8902-8585-448c-a1e8-0371a6ca7ca8|6b9ad766b55dBCDE|CCXT123456|b-WEEX111125|bc830de7-50f3-460b-9ee0-f430f83f9dad|p4sve|177321641268789|APCA-PARTNER-ID|OPEN-API-PARTNER|"BROKER-ID"|"Referrer"|X-Gate-Channel-Id|^\t\t\t"referral": |[?&](brokerId|broker_id|channel_code|partner|affiliate)=`)
	walkGoSources(t, func(path string, line int, text string) {
		if banned.MatchString(text) {
			t.Errorf("%s:%d: %s", path, line, strings.TrimSpace(text))
		}
	})
}

// Bithumb WS "ticket" is the requester id Bithumb's own WS requires (apidocs.bithumb.com,
// "Ticket Field": 요청자를 식별 할 수 있는 값, 필수 O). It goes only to ws-api.bithumb.com and
// is kept; this pins it to a fixed "ccxt" or a UUID so no partner value slips in.
func TestForkBithumbWsTicketIsNotAPartnerId(t *testing.T) {
	src, err := os.ReadFile("pro/bithumb.go")
	if err != nil {
		t.Fatal(err)
	}
	tickets := regexp.MustCompile(`"ticket":\s*([^,\n]+)`).FindAllStringSubmatch(string(src), -1)
	if len(tickets) == 0 {
		t.Fatal("no ticket found in pro/bithumb.go")
	}
	for _, m := range tickets {
		if v := strings.TrimSpace(m[1]); v != `"ccxt"` && v != "this.Uuid()" {
			t.Errorf("unexpected bithumb ws ticket value: %s", v)
		}
	}
}
