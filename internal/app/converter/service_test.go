package converter

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rogeecn/fabfile"
	"github.com/rogeecn/subconverter-go/internal/app/generator"
	"github.com/rogeecn/subconverter-go/internal/infra/config"
	"github.com/rogeecn/subconverter-go/internal/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestService_SupportedFormats(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})

	service := NewService(cfg, log)
	service.RegisterGenerators()

	formats := service.SupportedFormats()
	assert.Contains(t, formats, "clash")
	assert.Contains(t, formats, "surge")
	assert.Contains(t, formats, "quantumult")
	assert.Contains(t, formats, "loon")
	assert.Contains(t, formats, "v2ray")
	assert.Contains(t, formats, "surfboard")
}

func startHttpServer() *http.Server {
	http.Handle("/clash", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := fabfile.MustRead("fixtures/clash.txt")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	http.Handle("/v2ray", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := fabfile.MustRead("fixtures/v2ray.txt")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
	go http.ListenAndServe(":18080", nil)
	time.Sleep(time.Second)
	server := &http.Server{Addr: ":18080"}
	go func() {
		_ = server.ListenAndServe()
	}()
	time.Sleep(time.Second)
	return server
}

func TestService_Convert_SubLinks(t *testing.T) {
	srv := startHttpServer()
	defer srv.Shutdown(context.Background())

	Convey("Convert with sub links", t, func() {
		cfg := &config.Config{
			Cache: config.CacheConfig{TTL: 300},
		}
		log := logger.New(logger.Config{Level: "debug", Format: "text", Output: "stdout"})

		service := NewService(cfg, log)
		service.RegisterGenerators()

		Convey("invalid target", func() {
			req := &ConvertRequest{
				Target: "invalid",
				URLs:   []string{"https://example.com/subscription"},
			}
			_, err := service.Convert(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("empty URLs and extra links", func() {
			req := &ConvertRequest{
				Target: "clash",
				URLs:   []string{},
			}
			_, err := service.Convert(context.Background(), req)
			So(err, ShouldNotBeNil)
		})

		Convey("valid clash request convert to clash", func() {
			req := &ConvertRequest{
				Target: "clash",
				URLs: []string{
					"http://localhost:18080/clash",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			// Auto-validate generated Clash YAML content
			assert.Equal(t, "clash", resp.Format)
			assertClashConfig(t, resp)
		})

		Convey("valid v2ray request convert to clash", func() {
			req := &ConvertRequest{
				Target: "clash",
				URLs: []string{
					"http://localhost:18080/v2ray",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			// Auto-validate generated Clash YAML content
			assert.Equal(t, "clash", resp.Format)
			assertClashConfig(t, resp)
		})

		Convey("valid clash request convert to v2ray", func() {
			req := &ConvertRequest{
				Target: "v2ray",
				URLs: []string{
					"http://localhost:18080/clash",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			// Auto-validate V2Ray base64 subscription content
			assert.Equal(t, "v2ray", resp.Format)
			assertV2RaySubscription(t, resp)
		})

		Convey("valid v2ray request convert to v2ray", func() {
			req := &ConvertRequest{
				Target: "v2ray",
				URLs: []string{
					"http://localhost:18080/v2ray",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			// Auto-validate V2Ray base64 subscription content
			assert.Equal(t, "v2ray", resp.Format)
			assertV2RaySubscription(t, resp)
		})
	})
}

func TestService_Convert_RenameAndEmojiRules(t *testing.T) {
    // Two direct links with names to be transformed
    // SIP002: ss://BASE64(method:password)@host:port#name
    ss := "ss://YWVzLTI1Ni1nY206dGVzdA==@127.0.0.1:8388#US-01"
    trojan := "trojan://pass@us.example.com:443?security=tls#US-02"

    cfg := &config.Config{Cache: config.CacheConfig{TTL: 300}}
    log := logger.New(logger.Config{Level: "debug", Format: "text", Output: "stdout"})

    svc := NewService(cfg, log)
    svc.RegisterGenerators()

    req := &ConvertRequest{
        Target: "clash",
        URLs:   []string{ss, trojan},
        Options: Options{
            RenameRules: []generator.RenameRule{
                {Match: "US-", Replace: "美国-"},
            },
            EmojiRules: []generator.EmojiRule{
                {Match: "美国", Emoji: "🇺🇸"},
            },
            Sort: false,
        },
    }

    resp, err := svc.Convert(context.Background(), req)
    if err != nil {
        t.Fatalf("convert error: %v", err)
    }

    // Expect transformed names
    expected := map[string]bool{
        "🇺🇸 美国-01": false,
        "🇺🇸 美国-02": false,
    }
    for _, p := range resp.Proxies {
        if _, ok := expected[p.Name]; ok {
            expected[p.Name] = true
        }
        // ensure no original pattern remains
        if strings.Contains(p.Name, "US-") {
            t.Fatalf("proxy name was not renamed: %s", p.Name)
        }
    }
    for name, ok := range expected {
        if !ok {
            t.Fatalf("expected renamed proxy not found: %s", name)
        }
    }

    // Ensure Clash YAML contains the renamed/emojified names
    var m map[string]interface{}
    if err := yaml.Unmarshal([]byte(resp.Config), &m); err != nil {
        t.Fatalf("invalid clash yaml: %v", err)
    }
    arr, _ := m["proxies"].([]interface{})
    have := map[string]bool{}
    for _, it := range arr {
        if mp, ok := it.(map[string]interface{}); ok {
            if n, ok2 := mp["name"].(string); ok2 {
                have[n] = true
            }
        }
    }
    for name := range expected {
        if !have[name] {
            t.Fatalf("clash yaml does not include proxy name: %s", name)
        }
    }
}

// assertClashConfig validates that resp.Config is valid Clash YAML with expected structure
func assertClashConfig(t *testing.T, resp *ConvertResponse) {
	t.Helper()
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(resp.Config), &m); err != nil {
		t.Fatalf("clash config should be valid YAML: %v\nconfig: %s", err, resp.Config)
	}

	// Basic top-level keys
	if _, ok := m["proxies"]; !ok {
		t.Fatalf("clash config missing 'proxies' key")
	}
	if _, ok := m["proxy-groups"]; !ok {
		t.Fatalf("clash config missing 'proxy-groups' key")
	}
	if _, ok := m["rules"]; !ok {
		t.Fatalf("clash config missing 'rules' key")
	}

	// Proxies array should be present and non-empty; match resp.Proxies length
	proxies, ok := m["proxies"].([]interface{})
	if !ok {
		t.Fatalf("clash config 'proxies' should be an array")
	}
	if len(resp.Proxies) == 0 {
		t.Fatalf("response Proxies should not be empty")
	}
	if len(proxies) != len(resp.Proxies) {
		t.Fatalf("mismatched proxies count: yaml=%d resp=%d", len(proxies), len(resp.Proxies))
	}

	// Spot-check required fields exist for first proxy
	first, _ := proxies[0].(map[string]interface{})
	if first == nil {
		t.Fatalf("clash proxy entry should be a map")
	}
	for _, k := range []string{"name", "type", "server", "port"} {
		if _, ok := first[k]; !ok {
			t.Fatalf("clash proxy missing key '%s'", k)
		}
	}
}

// assertV2RaySubscription validates that resp.Config is a base64 subscription that decodes
// to newline-delimited links and matches the number of proxies in the response.
func assertV2RaySubscription(t *testing.T, resp *ConvertResponse) {
	t.Helper()
	dec, err := base64.StdEncoding.DecodeString(resp.Config)
	if err != nil {
		// try padding fix
		if m := len(resp.Config) % 4; m != 0 {
			padded := resp.Config + strings.Repeat("=", 4-m)
			dec, err = base64.StdEncoding.DecodeString(padded)
		}
	}
	if err != nil {
		t.Fatalf("v2ray config should be valid base64: %v", err)
	}
	text := string(dec)
	lines := 0
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		// Accept common schemes
		if strings.HasPrefix(ln, "vmess://") || strings.HasPrefix(ln, "vless://") ||
			strings.HasPrefix(ln, "trojan://") || strings.HasPrefix(ln, "ss://") {
			lines++
		} else {
			t.Fatalf("unexpected line in v2ray subscription: %s", ln)
		}
	}
	if lines == 0 {
		t.Fatalf("v2ray subscription should contain at least 1 link")
	}
	if lines != len(resp.Proxies) {
		t.Fatalf("mismatched link count: decoded=%d resp=%d", lines, len(resp.Proxies))
	}
}
