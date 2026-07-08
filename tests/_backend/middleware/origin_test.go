package middleware

import (
	"net/http"
	"testing"
)

func requestWithOrigin(host, origin string) *http.Request {
	r := &http.Request{Host: host, Header: http.Header{}}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

func TestIsAllowedOriginForRequest_SameOrigin(t *testing.T) {
	// 单体部署常态：Origin 与请求 Host 相同，必须放行（回归 WS 403 的根因）
	cases := []struct {
		host   string
		origin string
	}{
		{"game.example.com", "https://game.example.com"},
		{"game.example.com:8080", "http://game.example.com:8080"},
		{"127.0.0.1:5000", "http://127.0.0.1:5000"},
	}
	for _, c := range cases {
		r := requestWithOrigin(c.host, c.origin)
		if !IsAllowedOriginForRequest(r) {
			t.Errorf("same-origin request should be allowed: host=%s origin=%s", c.host, c.origin)
		}
	}
}

func TestIsAllowedOriginForRequest_EmptyOrigin(t *testing.T) {
	// 非浏览器客户端 / 同源无 Origin：放行
	if !IsAllowedOriginForRequest(requestWithOrigin("game.example.com", "")) {
		t.Error("empty origin should be allowed")
	}
}

func TestIsAllowedOriginForRequest_CrossSiteRejected(t *testing.T) {
	// 未配置白名单时，跨站来源应被拒绝（防跨站 WebSocket 劫持）
	r := requestWithOrigin("game.example.com", "https://evil.example.net")
	if IsAllowedOriginForRequest(r) {
		t.Error("cross-site origin should be rejected when not whitelisted")
	}
}
