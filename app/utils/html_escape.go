package utils

import "html"

// EscapeString 转义 HTML 特殊字符，防止 XSS 攻击
// 将 < > & " ' 等字符转换为 HTML 实体
//
// 示例:
//   EscapeString("<script>alert('XSS')</script>")
//   返回: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;"
func EscapeString(s string) string {
	return html.EscapeString(s)
}

// EscapeBytes 转义字节数组中的 HTML 特殊字符
func EscapeBytes(b []byte) []byte {
	return []byte(html.EscapeString(string(b)))
}
