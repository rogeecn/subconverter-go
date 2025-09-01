package converter

import (
    "context"
    "encoding/base64"
    "net/http"
    "net/http/httptest"
    "testing"

    appcfg "github.com/subconverter/subconverter-go/internal/infra/config"
    applog "github.com/subconverter/subconverter-go/internal/pkg/logger"
)

func TestService_Convert_OnlyCustomLinks(t *testing.T) {
    // upstream returns base64 of a vmess link
    vmess := "vmess://" + base64.StdEncoding.EncodeToString([]byte(`{"add":"x.example","port":"443","id":"11111111-1111-1111-1111-111111111111","net":"ws","host":"x.example","path":"/ws","tls":"tls"}`))
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(vmess))))
    }))
    defer upstream.Close()

    cfg := appcfg.Load()
    cfg.Subscription.ExtraLinks = []string{upstream.URL}
    log := applog.New(applog.Config{Level:"debug", Format:"json", Output:"stdout"})
    svc := NewService(cfg, log)
    svc.RegisterGenerators()

    // request with empty URLs should still work due to ExtraLinks
    resp, err := svc.Convert(context.Background(), &ConvertRequest{Target: "clash", URLs: []string{}})
    if err != nil { t.Fatalf("convert error: %v", err) }
    if len(resp.Proxies) == 0 { t.Fatalf("expected proxies from custom links") }
}

func TestService_Convert_MergeCustomLinks(t *testing.T) {
    // two upstreams
    one := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte("trojan://pass@a.example:443?security=tls&type=ws&host=a.example&path=%2Fws#A"))))
    }))
    defer one.Close()
    two := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte("trojan://pass@b.example:443?security=tls&type=ws&host=b.example&path=%2Fws#B"))))
    }))
    defer two.Close()

    cfg := appcfg.Load()
    cfg.Subscription.ExtraLinks = []string{one.URL}
    log := applog.New(applog.Config{Level:"debug", Format:"json", Output:"stdout"})
    svc := NewService(cfg, log)
    svc.RegisterGenerators()

    resp, err := svc.Convert(context.Background(), &ConvertRequest{Target: "clash", URLs: []string{two.URL}})
    if err != nil { t.Fatalf("convert error: %v", err) }
    if len(resp.Proxies) < 2 { t.Fatalf("expected merged proxies from both urls and extra links") }
}

