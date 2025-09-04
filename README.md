# SubConverter Go

SubConverter 的 Go 语言版本实现，提供高性能的代理订阅转换服务。

## 特性

- **高性能**: 基于 Go 语言的高并发处理能力
- **多协议支持**: 支持 SS、SSR、VMess、Trojan、Hysteria 等主流协议
- **多格式输出**: 支持 Clash、Surge、Quantumult、Loon 等格式
- **云原生**: 支持容器化部署，Kubernetes 友好
- **缓存支持**: 内存缓存，零依赖，快速稳定
- **API 友好**: RESTful API 设计，支持批量处理

## 快速开始

### 安装

```bash
# 克隆项目
git clone https://github.com/rogeecn/subconverter-go.git
cd subconverter-go

# 安装依赖
go mod tidy

# 构建
go build -o subconverter cmd/subconverter/main.go
```

### 运行

```bash
# 直接运行
./subconverter

# 使用配置文件
./subconverter --config configs/config.yaml

# 指定端口
./subconverter --port 8080
```

### Docker 运行

```bash
# 构建镜像
docker build -t subconverter-go .

# 运行容器
docker run -p 8080:8080 subconverter-go
```

## API 使用

仅保留两个 HTTP 路由：
- GET `/`：执行订阅转换（通过查询参数）
- GET `/health`：健康检查

### 转换订阅（GET，用于 Clash 等客户端）

客户端可直接填写合并订阅地址，无需手动下载配置：

```text
# 合并多个订阅为 Clash（target 默认 clash）
http://localhost:8080/?url=https://a.example/sub&url=https://b.example/sub&sort=1&udp=1

# 也可使用逗号分隔：
http://localhost:8080/?urls=https://a.example/sub,https://b.example/sub

# 过滤与重命名：
http://localhost:8080/?url=https://a/sub&include=HK&include=JP&rename=香港->HK&rename=日本->JP
```

支持的查询参数（常用）：
- `target`: 输出格式（`clash`/`surge`/`quantumult`/`loon`/`v2ray`/`surfboard`），默认 `clash`
- `url`: 订阅/节点链接（可重复）
- `urls`: 逗号分隔的链接
- `include`/`include_remarks`: 包含关键字（可重复 / 逗号分隔）
- `exclude`/`exclude_remarks`: 排除关键字（可重复 / 逗号分隔）
- `rename`: 重命名规则，形如 `旧->新`（可重复）
- `emoji`: 名称加表情，形如 `匹配:😊`（可重复）
- `sort`: `1/true` 开启排序
- `udp`: `1/true` 开启 UDP
- `rule`/`rules`: 追加自定义规则行（可重复 / 逗号分隔）
- `base_template`/`base`: 指定基础模板名称

### 配置额外独立节点（与订阅合并）

在配置文件中通过 `subscription.extra_links` 增加用户自定义的节点/订阅链接，它们会与请求中的 `url/urls` 一并合并后转换；当请求未提供任何订阅链接时，如果存在 `extra_links`，也会直接转换这些链接。

支持混合：`ss://`、`trojan://` 等独立节点链接，以及 `https://` 订阅链接。

示例（configs/config.yaml）：

```yaml
subscription:
  extra_links:
    # 独立节点链接示例
    - ss://YWVzLTI1Ni1nY206dGVzdEAxMjcuMC4wLjE6ODM4OA==#LOCAL
    - trojan://password@example.com:443?security=tls#TROJAN
    # 额外订阅链接，也可混合
    - https://upstream.example.com/sub
```

客户端可直接使用（未传 url 也可）：

```
http://localhost:8080/?target=clash
```

### 应用 base 规则（rules_dir 下的规则文件）

支持像 tindy2013/subconverter 一样，将仓库 `base/rules` 下的规则文件应用到生成的配置中。

- 配置文件方式（默认规则，所有请求生效）：

```yaml
generator:
  rules_dir: ./base/rules
  rule_files:
    - path: DivineEngine/Surge/Ruleset/Unbreak.list
      policy: DIRECT
    - path: ACL4SSR/Clash/ProxyMedia.list
      policy: "🚀 节点选择"
```

- 请求参数方式（按请求动态指定）：

说明：
- 规则文件路径为相对 `generator.rules_dir` 的相对路径。
- 每行规则若已含策略（逗号≥2），则保持原样；否则会追加 `,policy`（未提供 `policy` 时默认 `,DIRECT`）。
- 仅支持配置方式（`generator.rule_files`）。GET 路由不再支持按请求动态指定规则文件；如需临时追加少量规则，可使用 `rule`/`rules` 查询参数直接追加规则行。

### 快速验证（仅用 extra_links）

1. 在 `configs/config.yaml` 配置 `subscription.extra_links`（可混合 ss://、trojan://、https://）。
2. 启动服务：`./subconverter --config configs/config.yaml`（或 Docker 方式）。
3. 访问（不传 url，默认 target=clash 也可显式指定）：
   - `http://localhost:8080/?target=clash`
4. 返回应为 YAML。命令行验证：
   - `curl -I "http://localhost:8080/?target=clash" | grep Content-Type`
   - `curl -s "http://localhost:8080/?target=clash" | head -n 20`
5. 合并使用（将请求 url 与 extra_links 一并合并）：
   - `http://localhost:8080/?target=clash&url=https://example.com/sub`

### 健康检查

```bash
curl http://localhost:8080/health
```

### 客户端示例（Clash / Surge / Quantumult X / Loon）

- Clash / Clash.Meta（推荐）

  - 路径：Settings → Profiles → New Profile（URL）
  - 示例：
    - `http://<host>:8080/sub?url=https://a.example/sub&url=https://b.example/sub`
    - 仅用 extra_links：`http://<host>:8080/sub?target=clash`
  - 提示：支持别名 `/sub`，与 `/api/v1/convert` 等价。

- Surge（iOS/macOS）

  - 路径：Profile → Install from URL（或在已有配置中使用远程片段）
  - 示例（整份配置）：
    - `http://<host>:8080/api/v1/convert?target=surge&url=https://a.example/sub`
  - 若只想生成节点片段，可在规则侧另行合并（取决于现有配置结构）。

- Quantumult X（iOS）

  - 路径：Settings → Configuration → Download Configuration（或 Servers 导入远程）
  - 示例：
    - `http://<host>:8080/api/v1/convert?target=quantumult&url=https://a.example/sub`

- Loon（iOS）

  - 路径：Configuration → Remote（或从 URL 导入）
  - 示例：
    - `http://<host>:8080/api/v1/convert?target=loon&url=https://a.example/sub`

- 其他格式
  - V2Ray JSON：`http://<host>:8080/api/v1/convert?target=v2ray&url=...`
  - Surfboard：`http://<host>:8080/api/v1/convert?target=surfboard&url=...`

> 说明：不同客户端的“从 URL 导入/远程配置”入口名称略有差异，请以客户端当前版本实际界面为准。URL 参数可与前文一致（include/exclude/rename/emoji/sort/udp/base_template 等）。

## CLI 工具

### 安装 CLI

```bash
go install ./cmd/subctl
```

### 使用示例

```bash
# 转换订阅
subctl convert -u https://example.com/subscription -t clash -o config.yaml

# 使用自定义配置
subctl convert -u https://example.com/subscription -c configs/config.yaml
```

## 项目结构

```
subconverter-go/
├── cmd/
│   ├── subconverter/    # 主服务程序
│   ├── subctl/         # CLI工具
│   └── subworker/      # 后台任务
├── internal/
│   ├── app/
│   │   ├── converter/  # 转换服务
│   │   ├── parser/     # 协议解析
│   │   └── generator/  # 配置生成
│   ├── domain/
│   │   ├── proxy/      # 代理实体
│   │   ├── ruleset/    # 规则集实体
│   │   └── subscription/ # 订阅实体
│   ├── infra/
│   │   ├── cache/      # 缓存实现
│   │   ├── config/     # 配置管理
│   │   ├── http/       # HTTP客户端
│   │   └── storage/    # 存储抽象
│   └── pkg/
│       ├── logger/     # 日志封装
│       ├── errors/     # 错误处理
│       └── validator/  # 参数验证
├── configs/            # 配置文件
├── test/               # 测试文件
└── docs/               # 文档
```

## 开发指南

### 环境要求

- Go 1.21+
- Docker (可选)

### 开发运行

```bash
# 安装开发依赖
go mod tidy

# 运行测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 生成代码覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 致谢

感谢原始 SubConverter 项目的贡献者和社区支持。

### 常见错误与排查（客户端侧）

- TLS/SNI 不一致：证书域名校验失败。确保 `sni/servername` 与证书一致；WS 的 `Host` 头与反代配置匹配。
- WS 路径/编码：`path` 需以 `/` 开头；URI 中路径应做 URL 编码（`/ws` → `%2Fws`）。
- Base64 兼容：订阅常见移除 `=` 或使用 URL-safe Base64；解码前可补齐 `=` 或替换 `-_` 为 `+/`。
- Reality/XTLS 支持：需 Clash.Meta 或支持 Reality 的客户端；`flow/fp/pbk/sid/spx` 与服务端一致。
- 证书校验：测试或自签证书可开启 `skip-cert-verify/insecure`，生产建议正确配置完整信任链。
- Hy2 ALPN：与服务端一致（如 `h3`）；不一致会导致握手失败。
- DNS/解析：域名污染或解析异常时，客户端配置可信 DNS/Bootstrap（DoH/DoQ 亦可）。
- 时间同步：系统时间偏差会导致 TLS 失败；建议开启 NTP。

快速自检命令：

- `openssl s_client -connect host:443 -servername sni` 观察证书与握手。
- `curl -vk --resolve example.com:443:ip https://example.com` 验证 SNI/证书与反代。
