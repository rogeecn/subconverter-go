# VPN 订阅内容格式说明

本文档旨在说明 Shadowsocks、V2Ray、Clash 等不同 VPN 客户端的订阅内容格式。

## 1. 概述

VPN 订阅是一种方便用户获取和更新服务器节点信息的方式。服务提供商通常会提供一个 URL，用户将此 URL 添加到客户端后，客户端会从中下载服务器列表。

订阅内容本身通常是一个文本文件，其内容经过 Base64 编码。解码后，可以看到所有服务器节点的连接信息。

## 2. Shadowsocks

Shadowsocks 使用 `ss://` URI scheme 来表示单个服务器配置。

### 2.1. URI 格式

根据 [SIP002](https://shadowsocks.org/guide/sip002/) 规范，`ss://` 的格式如下：

```
ss://method:password@hostname:port#remark
```

- `method`: 加密方法，例如 `aes-256-gcm`
- `password`: 密码
- `hostname`: 服务器地址
- `port`: 服务器端口
- `remark`: 节点备注（可选）

为了方便传输，通常会将 `method:password` 部分进行 Base64 编码：

```
ss://<base64-encoded-method-and-password>@hostname:port#remark
```

**示例:**

```
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@server.example.com:8388#My-Server
```

其中 `YWVzLTI1Ni1nY206cGFzc3dvcmQ=` 是 `aes-256-gcm:password` 的 Base64 编码。

### 2.2. 订阅格式

Shadowsocks 的订阅文件内容是多个 `ss://` 链接，每行一个。然后将整个文本文件内容进行 Base64 编码。

**解码后的示例:**

```
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@server1.example.com:8388#Server1
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@server2.example.com:8388#Server2
```

## 3. V2Ray

V2Ray 使用 `vmess://` URI scheme。

### 3.1. URI 格式

`vmess://` 的内容是一个 Base64 编码的 JSON 对象，其中包含了 V2Ray 节点的详细配置。

**JSON 对象示例:**

```json
{
  "v": "2",
  "ps": "My-V2Ray-Server",
  "add": "server.example.com",
  "port": "443",
  "id": "your-uuid",
  "aid": "0",
  "net": "ws",
  "type": "none",
  "host": "server.example.com",
  "path": "/path",
  "tls": "tls"
}
```

将此 JSON 对象进行 Base64 编码后，附加到 `vmess://` 后面，即构成了 V2Ray 的节点链接。

### 3.2. 订阅格式

与 Shadowsocks 类似，V2Ray 的订阅文件内容是多个 `vmess://` 链接，每行一个。然后将整个文本文件内容进行 Base64 编码。

**解码后的示例:**

```
vmess://<base64-encoded-json-1>
vmess://<base64-encoded-json-2>
```

## 4. Clash

Clash 使用 YAML 格式的配置文件，功能更为强大，支持更复杂的规则和策略。

### 4.1. 配置文件格式

Clash 的订阅链接返回的是一个完整的 YAML 格式的配置文件。这个文件不仅包含了服务器节点信息，还定义了代理组、路由规则等。

**YAML 配置文件片段示例:**

```yaml
proxies:
  - name: "My-SS-Server"
    type: ss
    server: server.example.com
    port: 8388
    cipher: aes-256-gcm
    password: "password"
  - name: "My-V2Ray-Server"
    type: vmess
    server: server.example.com
    port: 443
    uuid: "your-uuid"
    alterId: 0
    cipher: auto
    tls: true
    network: ws
    ws-opts:
      path: "/path"
      headers:
        Host: server.example.com

proxy-groups:
  - name: "Proxy"
    type: select
    proxies:
      - "My-SS-Server"
      - "My-V2Ray-Server"

rules:
  - DOMAIN-SUFFIX,google.com,Proxy
  - MATCH,DIRECT
```

### 4.2. 订阅格式

Clash 的订阅链接直接指向这个 YAML 配置文件。客户端下载该文件后，会替换掉本地的配置文件。

## 5. Trojan

Trojan 使用 `trojan://` URI scheme，通常以 TLS 为基础传输，也支持基于 WebSocket/HTTP 的伪装参数。

### 5.1. URI 格式

```
trojan://password@hostname:port?security=tls&type=ws&host=example.com&path=%2Fws#remark
```

- `password`: 认证密码
- `hostname:port`: 服务器地址与端口（一般为 443）
- `security`: 传输安全，`tls` 或 `none`
- `type`: 传输类型，如 `tcp`、`ws`
- `host`/`sni`: TLS SNI/WS Host 头
- `path`: WebSocket 路径（需 URL 编码）

### 5.2. 订阅格式

多个 `trojan://` 链接按行排列，常见做法是对整个文本进行 Base64 编码后下发。

## 6. VLESS

VLESS 使用 `vless://` URI scheme，常与 TLS/Reality/WS 等传输组合。

### 6.1. URI 格式

```
vless://uuid@hostname:port?encryption=none&security=tls&type=ws&host=example.com&path=%2Fvless#remark
```

- `uuid`: 客户端 ID
- `encryption`: 固定 `none`
- 其他查询参数与 VMess/Trojan 类似（`security`、`type`、`host`、`path` 等）

### 6.2. 订阅格式

与 VMess 类似，多条 `vless://` 链接按行排列，通常整体 Base64 编码下发。

## 7. Hysteria / Hysteria2

本项目支持 `hysteria://`、`hysteria2://`。

### 7.1. Hysteria 示例

```
hysteria://hostname:port?auth=token&sni=example.com&alpn=h3#remark
```

### 7.2. Hysteria2 示例

```
hysteria2://hostname:port?auth=token&insecure=0&sni=example.com#remark
```

- 常见参数：`auth`（认证）、`sni`、`alpn`、`insecure` 等；不同客户端对参数支持略有差异。
- 订阅通常为多行链接整体 Base64 编码。

## 8. ShadowsocksR (SSR)

SSR 使用 `ssr://`，其内容本身是一个 Base64 编码的复合字符串，包含多项参数。

### 8.1. 链接结构（解码后）

```
hostname:port:protocol:method:obfs:base64(password)/?obfsparam=...&protoparam=...&remarks=...&group=...
```

示例：

```
ssr://<base64-encoded-string>
```

## 9. Snell

Snell 使用 `snell://`。

示例：

```
snell://chacha20-ietf-poly1305:password@hostname:port?obfs=http&obfs-host=example.com#remark
```

订阅按常见约定为多行 `snell://` 链接，整体 Base64 编码。

## 10. 总结

| 客户端        | 节点表示           | 订阅内容 (解码后)              | 格式                  |
| ------------- | ------------------ | ------------------------------ | --------------------- |
| Shadowsocks   | `ss://` URI        | 每行一个 `ss://` 链接          | Base64 编码的纯文本   |
| V2Ray (VMess) | `vmess://` URI     | 每行一个 `vmess://` 链接       | Base64 编码的纯文本   |
| VLESS         | `vless://` URI     | 每行一个 `vless://` 链接       | Base64 编码的纯文本   |
| Trojan        | `trojan://` URI    | 每行一个 `trojan://` 链接      | Base64 编码的纯文本   |
| Hysteria/Hy2  | `hysteria://`/`hysteria2://` | 多行链接                     | Base64 编码的纯文本   |
| SSR           | `ssr://` URI       | 每行一个 `ssr://` 链接         | Base64 编码的纯文本   |
| Snell         | `snell://` URI     | 每行一个 `snell://` 链接       | Base64 编码的纯文本   |
| Clash         | YAML 对象          | 完整的 YAML 配置文件           | YAML                  |

## 11. Reality/XTLS 参数说明（VLESS）

VLESS + Reality 多用于 TCP 直连场景，常见参数：

- `security=reality`: 启用 Reality。
- `flow=xtls-rprx-vision`: XTLS Vision 流控（常用）。
- `sni`: 指定 Server Name。
- `fp`: 指纹（如 `chrome`）。
- `pbk`: Reality 公钥（public key）。
- `sid`: Reality 短 ID（short id）。
- `spx`: SpiderX（如 `/` 或 `/path`）。

示例（VLESS Reality）：

```
vless://uuid@host:443?encryption=none&security=reality&flow=xtls-rprx-vision&fp=chrome&pbk=<pbk>&sid=<sid>&sni=example.com&spx=%2F#vless-reality
```

提示：Reality/XTLS 通常不与 WebSocket 同用；更适用于 TCP 直连。

## 12. 客户端参数映射速查表（示例）

说明：不同客户端版本字段名可能略有差异，请以实际版本文档为准。以下示例覆盖常见映射与写法。

### 12.1 Clash / Clash.Meta

- VLESS + Reality (XTLS Vision)（Clash.Meta 支持）

```yaml
proxies:
  - name: vless-reality
    type: vless
    server: host
    port: 443
    uuid: <uuid>
    udp: true
    tls: true
    servername: example.com     # 对应 sni
    flow: xtls-rprx-vision
    client-fingerprint: chrome  # 对应 fp
    reality-opts:
      public-key: <pbk>         # 对应 pbk
      short-id: <sid>           # 对应 sid
      spider-x: "/"            # 对应 spx
```

- VLESS + WS + TLS

```yaml
  - name: vless-ws
    type: vless
    server: host
    port: 443
    uuid: <uuid>
    udp: true
    tls: true
    servername: example.com
    network: ws
    ws-opts:
      path: "/path"
      headers:
        Host: example.com
```

- Trojan（TCP/WS + TLS）

```yaml
  - name: trojan-ws
    type: trojan
    server: host
    port: 443
    password: <password>
    sni: example.com            # 或 servername
    network: ws
    ws-opts:
      path: "/ws"
      headers:
        Host: example.com
```

- Hysteria2

```yaml
  - name: hy2
    type: hysteria2
    server: host
    port: 443
    password: <token>
    sni: example.com
    insecure: false
```

提示：标准 Clash 不支持 VLESS；上述 VLESS/Reality 需要 Clash.Meta。

### 12.2 Surge（iOS/macOS）

- Trojan（TCP/WS + TLS）

```
Proxy = trojan-ws, trojan, host, 443, password=<password>, sni=example.com, ws=true, ws-path=/ws, ws-headers=Host:example.com, tfo=true, udp-relay=true
```

- VMess（WS + TLS）

```
Proxy = vmess-ws, vmess, host, 443, username=<uuid>, tls=true, sni=example.com, ws=true, ws-path=/path, ws-headers=Host:example.com, skip-cert-verify=false
```

说明：Surge 目前不支持 VLESS/Reality，建议使用 Trojan/VMess/Hysteria2。

### 12.3 Quantumult X（iOS）

- Trojan（TCP/WS + TLS）

```
trojan=host:443, password=<password>, sni=example.com, ws=true, ws-path=/ws, ws-headers=Host:example.com, fast-open=true, udp-relay=true
```

- VMess（WS + TLS）

```
vmess=host:443, method=none, password=<uuid>, tls=true, tls-host=example.com, obfs=ws, obfs-host=example.com, obfs-uri=/path, fast-open=true, udp-relay=true
```

- VLESS + Reality（新版本支持，字段以实际版本为准）

```
vless=host:443, encryption=none, password=<uuid>, flow=xtls-rprx-vision, reality=true, sni=example.com, pbk=<pbk>, sid=<sid>, fp=chrome, fast-open=true
```

提示：Quantumult X 的键名可能随版本更新（如 `tls-host`/`sni`），以客户端内置提示为准。

### 12.4 Surge 的 Hysteria2（Hy2）示例

```
Proxy = hy2, hysteria2, host, 443, password=<token>, sni=example.com, alpn=h3, fast-open=true, udp-relay=true, skip-cert-verify=false
```

说明：`password` 对应 Hy2 的认证令牌；如需自签证书测试可将 `skip-cert-verify=true`。

## 13. URI → 客户端字段映射对照（常见项）

| URI 参数/字段            | 语义/协议               | Clash(.Meta)                | Surge                               | Quantumult X                         |
| ------------------------ | ----------------------- | --------------------------- | ----------------------------------- | ------------------------------------ |
| `hostname` / `add`       | 服务器地址              | `server`                    | 第2字段（主机名）                   | `host`（如 `vmess=host:...`）        |
| `port`                   | 端口                    | `port`                      | 第3字段（端口）                     | `:port`（如 `host:443`）             |
| `id`/`uuid` (VMess/VLESS)| 用户ID/UUID             | `uuid` (vless/vmess)        | `username=<uuid>`(vmess)            | `password=<uuid>`(vmess)、`encryption=none`(vless) |
| `password` (Trojan/Hy2)  | 密码/令牌               | `password`                  | `password=<...>`                    | `password=<...>`/`token=<...>`       |
| `security=tls`           | 启用 TLS                | `tls: true`                 | `tls=true`（默认）                  | `tls=true`/`tls-host=...`            |
| `sni` / `host`           | TLS SNI/WS Host         | `servername`/`ws-opts.headers.Host` | `sni=example.com`             | `tls-host=...`/`obfs-host=...`       |
| `network=ws`             | WebSocket 传输          | `network: ws` + `ws-opts`   | `ws=true, ws-path=..., ws-headers=...` | `obfs=ws, obfs-uri=..., obfs-host=...` |
| `path`                   | WS 路径                 | `ws-opts.path`              | `ws-path=/path`                     | `obfs-uri=/path`                     |
| `flow=xtls-rprx-vision`  | XTLS Vision（Reality）  | `flow: xtls-rprx-vision`    | 不支持（使用 Trojan/VMess/Hy2）     | `flow=xtls-rprx-vision`(vless 新版)  |
| `security=reality`       | Reality                 | `reality-opts.*` + `tls`    | 不支持                              | `reality=true, pbk, sid, fp`         |
| `pbk`/`sid`/`spx`/`fp`   | Reality 公钥/短ID/SpiderX/指纹 | `reality-opts.public-key/short-id/spider-x`, `client-fingerprint` | 不支持 | `pbk`/`sid`/`fp`（字段名随版本） |
| `insecure=1`/`skip-cert-verify=true` | 跳过证书校验 | `skip-cert-verify: true`    | `skip-cert-verify=true`             | `tls-verification=false`             |
| Hy2 专用 `alpn=h3`       | 指定 ALPN               | `alpn: [h3]`（若支持）       | `alpn=h3`                           | `alpn=h3`（若支持）                  |

提示：表格仅列常见关键字段。个别客户端对参数名、可选值和兼容性存在版本差异，实际以客户端文档为准。

像 `subconverter-go` 这样的工具，就是为了在这些不同的格式之间进行转换，以适应不同客户端的需求。

## 14. 逐字段示例转换（URI → Clash.Meta YAML）

以下示例展示如何从标准 URI 拆解字段并填入 Clash.Meta 配置。

### 14.1 VMess（WS + TLS）

- 示例 URI：

```
vmess://base64({"add":"example.com","port":"443","id":"<uuid>","net":"ws","host":"example.com","path":"/path","tls":"tls"})
```

- 字段映射：`add→server`，`port→port`，`id→uuid`，`net=ws→network/ws-opts`，`host→ws-opts.headers.Host`，`path→ws-opts.path`，`tls→tls:true`。

- 目标 YAML：

```yaml
proxies:
  - name: vmess-ws
    type: vmess
    server: example.com
    port: 443
    uuid: <uuid>
    cipher: auto
    tls: true
    network: ws
    ws-opts:
      path: "/path"
      headers:
        Host: example.com
```

### 14.2 VLESS（Reality + XTLS Vision）

- 示例 URI：

```
vless://<uuid>@example.com:443?encryption=none&security=reality&flow=xtls-rprx-vision&fp=chrome&pbk=<pbk>&sid=<sid>&sni=example.com&spx=%2F
```

- 字段映射：`uuid→uuid`，`security=reality→tls:true+reality-opts`，`flow→flow`，`fp→client-fingerprint`，`pbk/sid/spx→reality-opts.*`，`sni→servername`。

- 目标 YAML（Clash.Meta）：

```yaml
proxies:
  - name: vless-reality
    type: vless
    server: example.com
    port: 443
    uuid: <uuid>
    tls: true
    servername: example.com
    flow: xtls-rprx-vision
    client-fingerprint: chrome
    reality-opts:
      public-key: <pbk>
      short-id: <sid>
      spider-x: "/"
```

### 14.3 Trojan（WS + TLS）

- 示例 URI：

```
trojan://<password>@example.com:443?security=tls&type=ws&host=example.com&path=%2Fws
```

- 字段映射：`password→password`，`host→sni/ws-headers.Host`，`type=ws→network: ws`，`path→ws-opts.path`，`security=tls→tls`（Clash 使用 `sni` 字段）。

- 目标 YAML：

```yaml
proxies:
  - name: trojan-ws
    type: trojan
    server: example.com
    port: 443
    password: <password>
    sni: example.com
    network: ws
    ws-opts:
      path: "/ws"
      headers:
        Host: example.com
```

### 14.4 Hysteria2（Hy2）

- 示例 URI：

```
hysteria2://example.com:443?auth=<token>&sni=example.com&alpn=h3
```

- 字段映射：`auth→password`，`sni→sni`，`alpn=h3→alpn: [h3]`（如支持）。

- 目标 YAML：

```yaml
proxies:
  - name: hy2
    type: hysteria2
    server: example.com
    port: 443
    password: <token>
    sni: example.com
    alpn:
      - h3
```

## 15. Loon / Shadowrocket 示例

### 15.1 Loon（多数与 Surge 语法兼容）

- Trojan（WS + TLS）

```
# [Proxy] 节中
trojan = trojan-ws, example.com, 443, password=<password>, sni=example.com, ws=true, ws-path=/ws, ws-headers=Host:example.com, udp=true, tfo=true, skip-cert-verify=false
```

- VMess（WS + TLS）

```
vmess = vmess-ws, example.com, 443, username=<uuid>, tls=true, sni=example.com, ws=true, ws-path=/path, ws-headers=Host:example.com, skip-cert-verify=false
```

提示：不同版本的 Loon 可能对参数名有微调，基本与 Surge 一致。

### 15.2 Shadowrocket（iOS）

- Trojan（WS + TLS）

```
trojan=example.com:443, password=<password>, sni=example.com, ws=true, ws-path=/ws, ws-headers=Host:example.com, udp-relay=true, fast-open=true
```

- VMess（WS + TLS）

```
vmess=example.com:443, method=none, password=<uuid>, tls=true, tls-host=example.com, obfs=ws, obfs-host=example.com, obfs-uri=/path
```

- VLESS + Reality（新版本支持）

```
vless=example.com:443, encryption=none, password=<uuid>, reality=true, flow=xtls-rprx-vision, sni=example.com, pbk=<pbk>, sid=<sid>, fp=chrome
```

- Hysteria2（Hy2）

```
hysteria2=example.com:443, password=<token>, sni=example.com, alpn=h3, skip-cert-verify=false
```

提示：Shadowrocket 的键名可能随版本变化（如 `tls-host`/`sni`），以客户端内置文档为准。

## 16. 常见问题与排查清单

- TLS/SNI/Host 不一致：握手失败或证书域名不匹配。确保 TLS 的 `sni/servername` 与证书一致；WS 的 `Host` 头可与 SNI 不同，但需与服务端反代配置匹配。
- 路径/编码错误：WS `path` 需以 `/` 开头；URI 中路径需 URL 编码（如 `/ws` → `%2Fws`）。
- Base64 兼容性：VMess/VLESS 链接常见去掉 `=` 填充或使用 URL‑safe base64；解码前可按需补齐 `=` 或替换 `-_`。
- 备注与空白：`#remark` 需 URL 编码；去除换行与空格，避免多余 CRLF 影响解析。
- 证书校验：仅在测试时使用 `skip-cert-verify/insecure`；生产建议正确配置信任链与中间证书。
- Reality 参数：`pbk/sid/spx/flow(fp)` 需与服务端对应；VLESS/Reality 仅 Clash.Meta 等支持，标准 Clash/Surge 不支持。
- 传输不匹配：客户端 `network/alpn` 与服务端不一致（如 Hy2 `alpn=h3`），或端口被阻断。可用 `openssl s_client -connect host:443 -servername sni`/`curl -vk --resolve` 排查。
- 时间偏差：系统时间不准会导致 TLS 失败；确保 NTP 同步。

### 16.1 故障对照表（症状 → 原因 → 修复）

| 症状 | 可能原因 | 修复建议 |
| --- | --- | --- |
| TLS handshake failed | `sni` 与证书不匹配、证书链缺失 | 设置正确 `servername/sni`；补全中间证书或更新证书 |
| 400/404 on WS | `ws-opts.path` 错误、`Host` 头不匹配反代 | 路径以 `/` 开头；设置 `Host` 与反代一致 |
| Hy2 无法连通 | `alpn` 不匹配、UDP 被阻断 | 统一 `alpn=h3` 或移除；检查防火墙与端口映射 |
| 订阅解析失败 | Base64 填充缺失/URL-safe 变体 | 补齐 `=`，替换 `-_` 为 `+/` 再解码 |
| Reality 不生效 | `pbk/sid/spx/flow(fp)` 不一致 | 与服务端配置逐一核对，避免混用 WS |
| 速率异常 | 走错策略/直连 | 校验 `proxy-groups` 与 `rules`，使用基准测试定位 |
| DNS 异常 | 污染或解析错误 | 客户端设置可信 DNS/Bootstrap，或使用 DoH/DoQ |

## 17. 批量转换脚本示例（URI → Clash.Meta）

方案 A（推荐）：使用本仓库 CLI

```
# 安装 CLI
go install ./cmd/subctl

# 将一批订阅链接写入 urls.txt（每行一个 URL/URI）
# 使用内置转换为 Clash（可配合自定义模板/config）
subctl convert -u file://$(pwd)/urls.txt -t clash -o clash.yaml -c configs/config.yaml
```

方案 B（增强示例 Go 脚本：覆盖 VMess(WS+TLS) / VLESS(Reality/WS) / Trojan(WS+TLS) / Hysteria2 / SSR，输出完整 Clash.Meta 配置骨架）

```
// 保存为 tools/uri2clashmeta.go
// 用法：
//   go run tools/uri2clashmeta.go urls.txt > clash.yaml
package main

import (
  "bufio"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "io"
  "net/url"
  "os"
  "strings"
)

type vmess struct{ Add, Port, ID, Net, Host, Path, Tls string `json:"add","port","id","net","host","path","tls"` }

func main() {
  if len(os.Args) < 2 { fmt.Fprintln(os.Stderr, "usage: uri2clashmeta <file>"); os.Exit(1) }
  f, _ := os.Open(os.Args[1]); defer f.Close()
  var proxies []string
  var names []string
  s := bufio.NewScanner(f)
  s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
  for s.Scan() {
    line := strings.TrimSpace(s.Text())
    if line == "" || strings.HasPrefix(line, "#") { continue }
    switch {
    case strings.HasPrefix(line, "vmess://"):
      b64 := strings.TrimPrefix(line, "vmess://")
      // 补齐可能缺失的 '='
      if m := len(b64) % 4; m != 0 { b64 += strings.Repeat("=", 4-m) }
      raw, _ := base64.StdEncoding.DecodeString(b64)
      var v vmess; _ = json.Unmarshal(raw, &v)
      tls := strings.EqualFold(v.Tls, "tls")
      b := &strings.Builder{}
      fmt.Fprintf(b, "  - name: vmess-%s\n    type: vmess\n    server: %s\n    port: %s\n    uuid: %s\n    cipher: auto\n    tls: %v\n", v.Add, v.Add, v.Port, v.ID, tls)
      if strings.EqualFold(v.Net, "ws") {
        fmt.Fprintf(b, "    network: ws\n    ws-opts:\n      path: \"%s\"\n      headers:\n        Host: %s\n", v.Path, v.Host)
      }
      proxies = append(proxies, b.String())
      names = append(names, fmt.Sprintf("vmess-%s", v.Add))
    case strings.HasPrefix(line, "trojan://"):
      u, _ := url.Parse(line)
      q := u.Query()
      host := u.Hostname(); port := u.Port(); if port == "" { port = "443" }
      pwd := ""
      if u.User != nil { _, pwd = u.User.Username(), func(){p,_:=u.User.Password(); pwd=p}() }
      b := &strings.Builder{}
      fmt.Fprintf(b, "  - name: trojan-%s\n    type: trojan\n    server: %s\n    port: %s\n    password: %s\n", host, host, port, pwd)
      if sni := q.Get("sni"); sni != "" { fmt.Fprintf(b, "    sni: %s\n", sni) }
      if strings.EqualFold(q.Get("type"), "ws") {
        p := q.Get("path"); if p == "" { p = "/" } else if !strings.HasPrefix(p, "/") { p = "/" + strings.TrimPrefix(p, "%2F") }
        h := q.Get("host")
        fmt.Fprintf(b, "    network: ws\n    ws-opts:\n      path: \"%s\"\n      headers:\n        Host: %s\n", p, h)
      }
      proxies = append(proxies, b.String())
      names = append(names, fmt.Sprintf("trojan-%s", host))
    case strings.HasPrefix(line, "vless://"):
      u, _ := url.Parse(line); q := u.Query()
      host := u.Hostname(); port := u.Port(); if port == "" { port = "443" }
      uuid := u.User.Username()
      b := &strings.Builder{}
      fmt.Fprintf(b, "  - name: vless-%s\n    type: vless\n    server: %s\n    port: %s\n    uuid: %s\n", host, host, port, uuid)
      if strings.EqualFold(q.Get("security"), "reality") {
        fmt.Fprintf(b, "    tls: true\n")
        if sni := q.Get("sni"); sni != "" { fmt.Fprintf(b, "    servername: %s\n", sni) }
        if flow := q.Get("flow"); flow != "" { fmt.Fprintf(b, "    flow: %s\n", flow) }
        if fp := q.Get("fp"); fp != "" { fmt.Fprintf(b, "    client-fingerprint: %s\n", fp) }
        fmt.Fprintf(b, "    reality-opts:\n      public-key: %s\n      short-id: %s\n", q.Get("pbk"), q.Get("sid"))
        spx := q.Get("spx"); if spx == "" { spx = "/" }
        fmt.Fprintf(b, "      spider-x: \"%s\"\n", strings.TrimPrefix(spx, "/"))
      } else if strings.EqualFold(q.Get("type"), "ws") || strings.EqualFold(q.Get("network"), "ws") {
        fmt.Fprintf(b, "    tls: %v\n", strings.EqualFold(q.Get("security"), "tls"))
        if sni := q.Get("sni"); sni != "" { fmt.Fprintf(b, "    servername: %s\n", sni) }
        p := q.Get("path"); if p == "" { p = "/" } else if !strings.HasPrefix(p, "/") { p = "/" + strings.TrimPrefix(p, "%2F") }
        h := q.Get("host")
        fmt.Fprintf(b, "    network: ws\n    ws-opts:\n      path: \"%s\"\n      headers:\n        Host: %s\n", p, h)
      }
      proxies = append(proxies, b.String())
      names = append(names, fmt.Sprintf("vless-%s", host))
    case strings.HasPrefix(line, "hysteria2://"):
      u, _ := url.Parse(line); q := u.Query()
      host := u.Hostname(); port := u.Port(); if port == "" { port = "443" }
      token := q.Get("auth"); if token == "" { token = q.Get("password") }
      b := &strings.Builder{}
      fmt.Fprintf(b, "  - name: hy2-%s\n    type: hysteria2\n    server: %s\n    port: %s\n    password: %s\n", host, host, port, token)
      if sni := q.Get("sni"); sni != "" { fmt.Fprintf(b, "    sni: %s\n", sni) }
      if strings.EqualFold(q.Get("alpn"), "h3") { fmt.Fprintf(b, "    alpn:\n      - h3\n") }
      if q.Get("insecure") == "1" { fmt.Fprintf(b, "    insecure: true\n") }
      proxies = append(proxies, b.String())
      names = append(names, fmt.Sprintf("hy2-%s", host))
    case strings.HasPrefix(line, "ssr://"):
      b64 := strings.TrimPrefix(line, "ssr://")
      // URL-safe 兼容
      b64 = strings.ReplaceAll(b64, "-", "+")
      b64 = strings.ReplaceAll(b64, "_", "/")
      if m := len(b64) % 4; m != 0 { b64 += strings.Repeat("=", 4-m) }
      raw, err := base64.StdEncoding.DecodeString(b64); if err != nil { continue }
      parts := strings.SplitN(string(raw), "/?", 2)
      main := parts[0]
      qstr := ""; if len(parts) == 2 { qstr = parts[1] }
      seg := strings.Split(main, ":")
      if len(seg) < 6 { continue }
      host, port, protocol, method, obfs, passb64 := seg[0], seg[1], seg[2], seg[3], seg[4], seg[5]
      if m := len(passb64) % 4; m != 0 { passb64 += strings.Repeat("=", 4-m) }
      passb64 = strings.ReplaceAll(passb64, "-", "+")
      passb64 = strings.ReplaceAll(passb64, "_", "/")
      pwdBytes, _ := base64.StdEncoding.DecodeString(passb64)
      pwd := string(pwdBytes)
      v, _ := url.ParseQuery(qstr)
      obfsParam := v.Get("obfsparam")
      protoParam := v.Get("protoparam")
      b := &strings.Builder{}
      fmt.Fprintf(b, "  - name: ssr-%s\n    type: ssr\n    server: %s\n    port: %s\n    cipher: %s\n    password: %s\n    protocol: %s\n    obfs: %s\n", host, host, port, method, pwd, protocol, obfs)
      if obfsParam != "" { fmt.Fprintf(b, "    obfs-param: %s\n", obfsParam) }
      if protoParam != "" { fmt.Fprintf(b, "    protocol-param: %s\n", protoParam) }
      proxies = append(proxies, b.String())
      names = append(names, fmt.Sprintf("ssr-%s", host))
    }
  }
  // 输出完整 Clash 配置骨架
  write := func(w io.Writer, s string) { _, _ = io.WriteString(w, s) }
  write(os.Stdout, "port: 7890\n")
  write(os.Stdout, "socks-port: 7891\n")
  write(os.Stdout, "allow-lan: true\nmode: Rule\nlog-level: info\n\n")
  write(os.Stdout, "proxies:\n")
  for _, p := range proxies { write(os.Stdout, p) }
  write(os.Stdout, "\nproxy-groups:\n")
  write(os.Stdout, "  - name: Proxy\n    type: select\n    proxies:\n")
  for _, n := range names { write(os.Stdout, fmt.Sprintf("      - %s\n", n)) }
  write(os.Stdout, "  - name: Auto\n    type: fallback\n    url: https://www.google.com/generate_204\n    interval: 300\n    proxies:\n")
  for _, n := range names { write(os.Stdout, fmt.Sprintf("      - %s\n", n)) }
  write(os.Stdout, "\nrules:\n  - GEOIP,CN,DIRECT\n  - MATCH,Proxy\n")
}
```

注意：示例脚本仍未覆盖全部边界（如复杂传输、完整证书校验、更多可选字段）。生产使用请优先选择 `subctl` 或服务端 API 进行转换；SSR/Reality 等需 Clash.Meta 或兼容分支支持。
