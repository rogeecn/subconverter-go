package converter

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/rogeecn/fabfile"
	"github.com/rogeecn/subconverter-go/internal/infra/config"
	"github.com/rogeecn/subconverter-go/internal/pkg/logger"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
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

	FocusConvey("Convert with sub links", t, func() {
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

			t.Logf("%# v", pretty.Formatter(resp))
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

			t.Logf("%# v", pretty.Formatter(resp))
		})

		FocusConvey("valid clash request convert to v2ray", func() {
			req := &ConvertRequest{
				Target: "v2ray",
				URLs: []string{
					"http://localhost:18080/clash",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			t.Logf("%# v", pretty.Formatter(resp))
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

			t.Logf("%# v", pretty.Formatter(resp))
		})
	})
}
