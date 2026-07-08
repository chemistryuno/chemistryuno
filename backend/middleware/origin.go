package middleware

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// 允许的跨源来源白名单。
//
// 生产环境应通过 CORS_ALLOWED_ORIGINS（逗号分隔）显式配置允许的来源。
// 未配置时，会回退到已知的应用来源环境变量，并额外放行本地开发地址
// （localhost / 127.0.0.1，任意端口），以保证本地开发体验不受影响。
var (
	originsOnce      sync.Once
	allowedOriginSet map[string]struct{}
)

func loadAllowedOrigins() {
	allowedOriginSet = make(map[string]struct{})

	add := func(raw string) {
		for _, part := range strings.Split(raw, ",") {
			o := strings.TrimSpace(part)
			if o == "" {
				continue
			}
			allowedOriginSet[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
		}
	}

	// 显式白名单优先
	add(os.Getenv("CORS_ALLOWED_ORIGINS"))

	// 回退到已知的应用来源
	add(os.Getenv("WEBAUTHN_RP_ORIGIN"))
	add(os.Getenv("WEBAUTHN_ORIGIN"))
	add(os.Getenv("VITE_SERVER_ORIGIN"))
	add(os.Getenv("CHEM_SERVER_ORIGIN"))
}

// isLocalhostOrigin 判断 origin 是否为本地开发地址（任意端口）。
func isLocalhostOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// IsAllowedOrigin 判断给定 Origin 是否被允许发起跨源请求 / WebSocket 连接。
// 空 Origin（同源或非浏览器客户端）视为允许。
func IsAllowedOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return true
	}
	originsOnce.Do(loadAllowedOrigins)

	normalized := strings.ToLower(strings.TrimRight(origin, "/"))
	if _, ok := allowedOriginSet[normalized]; ok {
		return true
	}

	// 未显式配置白名单时，放行本地开发地址，避免破坏本地开发
	if len(allowedOriginSet) == 0 && isLocalhostOrigin(origin) {
		return true
	}
	return false
}

// IsAllowedOriginForRequest 在 IsAllowedOrigin 的基础上额外放行“同源”请求。
//
// 生产环境常见为单体部署（前端静态资源嵌入后端，同域提供），此时浏览器
// 发来的 Origin 与请求自身的 Host 相同。这类同源请求不存在跨站攻击风险，
// 必须放行，否则未配置 CORS_ALLOWED_ORIGINS 白名单时会误拒真实用户
// （表现为 WebSocket 升级 403 / CheckOrigin 拒绝）。
//
// 判定顺序：空 Origin → 放行；同源（Origin host == 请求 Host）→ 放行；
// 否则回落到白名单校验（防跨站 WebSocket 劫持 / 跨源携带凭证）。
func IsAllowedOriginForRequest(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	if isSameOrigin(origin, r.Host) {
		return true
	}
	return IsAllowedOrigin(origin)
}

// isSameOrigin 判断 Origin 的 host（含端口）是否与请求 Host 一致。
func isSameOrigin(origin, host string) bool {
	if host == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	// url.Host 含端口（如 example.com:8080），与 r.Host 语义一致
	return strings.EqualFold(u.Host, host)
}
