package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rogeecn/subconverter-go/internal/app/converter"
	appcfg "github.com/rogeecn/subconverter-go/internal/infra/config"
	applog "github.com/rogeecn/subconverter-go/internal/pkg/logger"
)

func TestGETConvert_MergeAndReturnYAML(t *testing.T) {
	// prepare fake subscription endpoints
	vmess := "vmess://" + base64.StdEncoding.EncodeToString([]byte(`{"v":"2","ps":"VM-x","add":"x.example","port":"443","id":"11111111-1111-1111-1111-111111111111","net":"ws","host":"x.example","path":"/ws","tls":"tls"}`))
	trojan := "trojan://pass@y.example:443?security=tls&type=ws&host=y.example&path=%2Fws#T1"
	content := base64.StdEncoding.EncodeToString([]byte(vmess + "\n" + trojan))
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	defer upstream.Close()

	// boot app
	cfg := appcfg.Load()
	log := applog.New(applog.Config{Level: "debug", Format: "json", Output: "stdout"})
	svc := converter.NewService(cfg, log)
	svc.RegisterGenerators()
	rt := NewRouter(svc, cfg)
	rt.SetupRoutes()

	// issue GET request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/convert?target=clash&url="+upstream.URL+"&sort=1&udp=1", nil)
	resp, err := rt.App().Test(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-yaml") {
		t.Fatalf("unexpected content type: %s", ct)
	}
}
