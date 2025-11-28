package helpers

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
)

// ParseUserAgent 解析User-Agent字符串，返回浏览器和操作系统信息
func ParseUserAgent(userAgent string) (browser, os string) {
	if userAgent == "" {
		return "Unknown", "Unknown"
	}

	ua := strings.ToLower(userAgent)

	// 解析浏览器
	browser = parseBrowser(ua)

	// 解析操作系统
	os = parseOS(ua)

	return browser, os
}

// parseBrowser 解析浏览器类型
func parseBrowser(ua string) string {
	// Chrome
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") && !strings.Contains(ua, "opr") {
		// 提取Chrome版本
		if idx := strings.Index(ua, "chrome/"); idx != -1 {
			version := extractVersion(ua, idx+7)
			return "Chrome " + version
		}
		return "Chrome"
	}

	// Edge
	if strings.Contains(ua, "edg") {
		if idx := strings.Index(ua, "edg/"); idx != -1 {
			version := extractVersion(ua, idx+4)
			return "Edge " + version
		}
		return "Edge"
	}

	// Firefox
	if strings.Contains(ua, "firefox") {
		if idx := strings.Index(ua, "firefox/"); idx != -1 {
			version := extractVersion(ua, idx+8)
			return "Firefox " + version
		}
		return "Firefox"
	}

	// Safari
	if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		if idx := strings.Index(ua, "version/"); idx != -1 {
			version := extractVersion(ua, idx+8)
			return "Safari " + version
		}
		return "Safari"
	}

	// Opera
	if strings.Contains(ua, "opr") || strings.Contains(ua, "opera") {
		if idx := strings.Index(ua, "opr/"); idx != -1 {
			version := extractVersion(ua, idx+4)
			return "Opera " + version
		}
		if idx := strings.Index(ua, "version/"); idx != -1 {
			version := extractVersion(ua, idx+8)
			return "Opera " + version
		}
		return "Opera"
	}

	// IE
	if strings.Contains(ua, "msie") || strings.Contains(ua, "trident") {
		if idx := strings.Index(ua, "msie "); idx != -1 {
			version := extractVersion(ua, idx+5)
			return "IE " + version
		}
		return "IE"
	}

	return "Unknown"
}

// parseOS 解析操作系统
func parseOS(ua string) string {
	// Windows
	if strings.Contains(ua, "windows") {
		if strings.Contains(ua, "windows nt 10.0") || strings.Contains(ua, "windows nt 6.3") {
			return "Windows 10/11"
		}
		if strings.Contains(ua, "windows nt 6.2") {
			return "Windows 8"
		}
		if strings.Contains(ua, "windows nt 6.1") {
			return "Windows 7"
		}
		if strings.Contains(ua, "windows nt 6.0") {
			return "Windows Vista"
		}
		if strings.Contains(ua, "windows nt 5.1") {
			return "Windows XP"
		}
		return "Windows"
	}

	// macOS
	if strings.Contains(ua, "mac os x") || strings.Contains(ua, "macintosh") {
		if idx := strings.Index(ua, "mac os x "); idx != -1 {
			version := extractVersion(ua, idx+9)
			return "macOS " + version
		}
		return "macOS"
	}

	// iOS
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		if idx := strings.Index(ua, "os "); idx != -1 {
			version := extractVersion(ua, idx+3)
			version = strings.ReplaceAll(version, "_", ".")
			return "iOS " + version
		}
		return "iOS"
	}

	// Android
	if strings.Contains(ua, "android") {
		if idx := strings.Index(ua, "android "); idx != -1 {
			version := extractVersion(ua, idx+8)
			return "Android " + version
		}
		return "Android"
	}

	// Linux
	if strings.Contains(ua, "linux") {
		return "Linux"
	}

	return "Unknown"
}

// extractVersion 从User-Agent字符串中提取版本号
func extractVersion(ua string, startIdx int) string {
	if startIdx >= len(ua) {
		return ""
	}

	var version strings.Builder
	for i := startIdx; i < len(ua); i++ {
		c := ua[i]
		if (c >= '0' && c <= '9') || c == '.' || c == '_' {
			version.WriteByte(c)
		} else {
			break
		}
	}

	result := version.String()
	if len(result) > 10 {
		return result[:10]
	}
	return result
}

// GetBrowserAndOS 从HTTP上下文获取浏览器和操作系统信息
func GetBrowserAndOS(ctx http.Context) (browser, os string) {
	userAgent := ctx.Request().Header("User-Agent", "")
	return ParseUserAgent(userAgent)
}
