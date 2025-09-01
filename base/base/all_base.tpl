# SubConverter Base Template
# This template provides the base structure for Clash configurations

mixed-port: {{ .Port }}
socks-port: {{ .SocksPort }}
allow-lan: {{ .AllowLan }}
mode: {{ .Mode }}
log-level: {{ .LogLevel }}
external-controller: {{ .ExternalController }}
secret: {{ .Secret }}

proxies:
{{- range .Proxies }}
{{ . }}
{{- end }}

proxy-groups:
{{- range .ProxyGroups }}
{{ . }}
{{- end }}

rules:
{{- range .Rules }}
{{ . }}
{{- end }}