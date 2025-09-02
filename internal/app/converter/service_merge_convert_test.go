package converter

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcfg "github.com/rogeecn/subconverter-go/internal/infra/config"
	applog "github.com/rogeecn/subconverter-go/internal/pkg/logger"
)

// helper to craft a simple vmess url
func makeVMess(add, uuid string) string {
	js := `{"v":"2","ps":"VM-` + add + `","add":"` + add + `","port":"443","id":"` + uuid + `","net":"ws","type":"none","host":"` + add + `","path":"/ws","tls":"tls"}`
	return "vmess://" + base64.StdEncoding.EncodeToString([]byte(js))
}

func makeTrojan(host string) string {
	return "trojan://pass@" + host + ":443?security=tls&type=ws&host=" + host + "&path=%2Fws#TR-" + host
}

func TestService_Convert_MergeMultipleURLs_ToClash(t *testing.T) {
	// prepare two subscription endpoints (base64-encoded multi-line)
	sub1 := makeVMess("a.example", "11111111-1111-1111-1111-111111111111") + "\n" + makeTrojan("a.example")
	sub2 := makeVMess("b.example", "22222222-2222-2222-2222-222222222222") + "\n" + makeTrojan("b.example")
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(sub1))))
	}))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString([]byte(sub2))))
	}))
	defer srv2.Close()

	// init service
	cfg := appcfg.Load()
	log := applog.New(applog.Config{Level: "debug", Format: "json", Output: "stdout"})
	svc := NewService(cfg, log)
	svc.RegisterGenerators()

	// convert
	resp, err := svc.Convert(context.Background(), &ConvertRequest{
		Target:  "clash",
		URLs:    []string{srv1.URL, srv2.URL},
		Options: Options{Sort: true, UDP: true},
	})
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if resp == nil || len(resp.Proxies) != 4 {
		t.Fatalf("expected 4 proxies, got %d", len(resp.Proxies))
	}
	if resp.Format != "clash" {
		t.Fatalf("unexpected format: %s", resp.Format)
	}
	if !(strings.Contains(resp.Config, "vmess") || strings.Contains(resp.Config, "type: vmess")) {
		t.Fatalf("clash config does not contain vmess section: %s", resp.Config)
	}
}
