package converter

import (
	"context"
	"testing"

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

func TestService_Convert_SubLinks(t *testing.T) {
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
			t.Logf("got err: %+v", err)
		})

		Convey("empty URLs and extra links", func() {
			req := &ConvertRequest{
				Target: "clash",
				URLs:   []string{},
			}
			_, err := service.Convert(context.Background(), req)
			So(err, ShouldNotBeNil)
			t.Logf("got err: %+v", err)
		})

		FocusConvey("valid clash request", func() {
			req := &ConvertRequest{
				Target: "clash",
				URLs: []string{
					"https://r64mx8i.waimaody.cc/sub/7dd9519e6e3fbea2/clash",
				},
			}
			resp, err := service.Convert(context.Background(), req)
			So(err, ShouldBeNil)

			t.Logf("resp: %+v", resp)
		})
	})
}
