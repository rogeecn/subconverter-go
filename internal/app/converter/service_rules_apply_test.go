package converter

import (
    "context"
    "path/filepath"
    "strings"
    "testing"

    appcfg "github.com/rogeecn/subconverter-go/internal/infra/config"
    applog "github.com/rogeecn/subconverter-go/internal/pkg/logger"
)

// Verify that rule files under base/rules are loaded and policies applied to Clash rules
func TestService_Convert_ApplyRuleFiles(t *testing.T) {
    cfg := appcfg.Load()
    // Ensure rules_dir points to repo base/rules regardless of test working dir
    cfg.Generator.RulesDir = filepath.Join("..", "..", "..", "base", "rules")
    log := applog.New(applog.Config{Level: "debug", Format: "json", Output: "stdout"})
    svc := NewService(cfg, log)
    svc.RegisterGenerators()

    req := &ConvertRequest{
        Target: "clash",
        URLs:   []string{"trojan://pass@host.example:443?security=tls#T1"},
        Options: Options{
            RuleFiles: []RuleFileRef{{
                Path:   "DivineEngine/Surge/Ruleset/Unbreak.list",
                Policy: "DIRECT",
            }},
        },
    }

    resp, err := svc.Convert(context.Background(), req)
    if err != nil {
        t.Fatalf("convert error: %v", err)
    }
    if resp.Format != "clash" {
        t.Fatalf("unexpected format: %s", resp.Format)
    }
    if !strings.Contains(resp.Config, "DOMAIN,fonts.googleapis.com,DIRECT") &&
        !strings.Contains(resp.Config, "DOMAIN,fonts.gstatic.com,DIRECT") {
        t.Fatalf("expected loaded rule with policy not found in config")
    }
}
