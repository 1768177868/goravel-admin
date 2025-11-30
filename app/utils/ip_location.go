package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IPLocation IP 地理位置信息
type IPLocation struct {
	Country     string `json:"country"`     // 国家
	Region      string `json:"region"`      // 省份/州
	City        string `json:"city"`        // 城市
	ISP         string `json:"isp"`         // ISP
	CountryCode string `json:"countryCode"` // 国家代码
}

// GetIPLocation 根据 IP 地址获取地理位置信息
// 使用 ip-api.com 免费 API（无需 API Key，有速率限制）
// 如果查询失败，返回空字符串，不影响主流程
func GetIPLocation(ip string) string {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
		return "内网IP"
	}

	// 使用 ip-api.com 免费 API
	// 格式：http://ip-api.com/json/{ip}?fields=status,message,country,regionName,city,isp,countryCode
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,regionName,city,isp,countryCode&lang=zh-CN", ip)

	client := &http.Client{
		Timeout: 3 * time.Second, // 3秒超时，避免阻塞
	}

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var result struct {
		Status      string `json:"status"`
		Message     string `json:"message"`
		Country     string `json:"country"`
		RegionName  string `json:"regionName"`
		City        string `json:"city"`
		ISP         string `json:"isp"`
		CountryCode string `json:"countryCode"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	if result.Status != "success" {
		return ""
	}

	// 构建位置字符串：国家 省份 城市
	var locationParts []string
	if result.Country != "" {
		locationParts = append(locationParts, result.Country)
	}
	if result.RegionName != "" {
		locationParts = append(locationParts, result.RegionName)
	}
	if result.City != "" {
		locationParts = append(locationParts, result.City)
	}

	location := strings.Join(locationParts, " ")
	if location == "" {
		return ""
	}

	// 如果位置信息太长，截断
	if len(location) > 100 {
		location = location[:100]
	}

	return location
}

// GetIPLocationAsync 异步获取 IP 地理位置信息
// 用于不阻塞主流程的场景
func GetIPLocationAsync(ip string, callback func(location string)) {
	go func() {
		location := GetIPLocation(ip)
		if callback != nil {
			callback(location)
		}
	}()
}
