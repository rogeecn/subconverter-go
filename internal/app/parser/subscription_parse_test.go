package parser

import (
    "context"
    "encoding/base64"
    "testing"
)

func TestManager_Parse_Base64Subscription(t *testing.T) {
    m := NewManager()
    ctx := context.Background()

    raw := "vmess://" + base64.StdEncoding.EncodeToString([]byte(`{"add":"example.com","port":"443","id":"11111111-1111-1111-1111-111111111111","net":"ws","host":"example.com","path":"/ws","tls":"tls"}`)) + "\n" +
        "trojan://pass@example.org:443?security=tls&type=ws&host=ex.org&path=%2Fws#t1"
    b64 := base64.StdEncoding.EncodeToString([]byte(raw))

    proxies, err := m.Parse(ctx, b64)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if len(proxies) != 2 { t.Fatalf("expected 2 proxies, got %d", len(proxies)) }
}

func TestVLESSParser_RealityAndWS(t *testing.T) {
    p := NewVLESSParser()
    ctx := context.Background()

    uri := "vless://11111111-1111-1111-1111-111111111111@host:443?encryption=none&security=reality&flow=xtls-rprx-vision&fp=chrome&pbk=pubkey&sid=abcd&sni=example.com&spx=%2F#v1"
    list, err := p.Parse(ctx, uri)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    v := list[0]
    if v.Security != "reality" || v.TLS == TLSNone || v.RealityPublicKey != "pubkey" || v.SNI != "example.com" {
        t.Fatalf("vless reality fields not parsed correctly: %+v", v)
    }

    uri2 := "vless://11111111-1111-1111-1111-111111111111@host:443?encryption=none&security=tls&type=ws&host=ex.com&path=%2Fws&sni=ex.com#v2"
    list2, err := p.Parse(ctx, uri2)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    v2 := list2[0]
    if v2.Path == "" || v2.Host == "" || v2.TLS == TLSNone {
        t.Fatalf("vless ws/tls fields not parsed correctly: %+v", v2)
    }
}

