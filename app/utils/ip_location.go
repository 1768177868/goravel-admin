package utils

import (
	"context"
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
//
// 安全特性：
//   - 使用 context 超时控制（默认 10 秒）
//   - 添加 panic recovery，防止 goroutine 崩溃
//   - callback 执行超时保护（默认 5 秒），防止阻塞
//
// 参数:
//   - ip: IP 地址
//   - callback: 回调函数，接收位置信息
func GetIPLocationAsync(ip string, callback func(location string)) {
	GetIPLocationAsyncWithTimeout(ip, callback, 10*time.Second, 5*time.Second)
}

// GetIPLocationAsyncWithTimeout 异步获取 IP 地理位置信息（带超时控制）
//
// 参数:
//   - ip: IP 地址
//   - callback: 回调函数，接收位置信息
//   - locationTimeout: IP 查询超时时间（默认 10 秒）
//   - callbackTimeout: 回调函数执行超时时间（默认 5 秒）
func GetIPLocationAsyncWithTimeout(ip string, callback func(location string), locationTimeout, callbackTimeout time.Duration) {
	// 设置默认超时时间
	if locationTimeout <= 0 {
		locationTimeout = 10 * time.Second
	}
	if callbackTimeout <= 0 {
		callbackTimeout = 5 * time.Second
	}

	go func() {
		// 添加 panic recovery，防止 goroutine 崩溃导致程序退出
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 但不影响主流程
				// 注意：这里不能使用 facades.Log()，因为可能没有初始化 context
				// 如果需要记录日志，可以通过参数传入 logger
			}
		}()

		// 创建带超时的 context（确保总超时时间不超过 locationTimeout）
		// 注意：GetIPLocation 内部已有 3 秒 HTTP 超时，这里设置外层超时作为兜底
		ctx, cancel := context.WithTimeout(context.Background(), locationTimeout)
		defer cancel()

		// 在 goroutine 中执行 IP 查询
		locationChan := make(chan string, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// panic 时发送空字符串，避免阻塞
					select {
					case locationChan <- "":
					default:
					}
				}
			}()
			location := GetIPLocation(ip)
			// 非阻塞发送，避免 goroutine 泄露
			select {
			case locationChan <- location:
			case <-ctx.Done():
				// 如果已经超时，直接返回，不发送结果
				return
			}
		}()

		// 等待结果或超时
		var location string
		select {
		case location = <-locationChan:
			// 成功获取位置信息
		case <-ctx.Done():
			// 超时，返回空字符串
			// 注意：内部的 goroutine 可能仍在执行，但由于 HTTP client 有 3 秒超时，会很快结束
			location = ""
		}

		// 执行回调（带超时保护）
		if callback != nil {
			callbackCtx, callbackCancel := context.WithTimeout(context.Background(), callbackTimeout)
			defer callbackCancel()

			callbackDone := make(chan struct{}, 1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// callback 中发生 panic，静默处理
					}
					callbackDone <- struct{}{}
				}()
				callback(location)
			}()

			// 等待 callback 完成或超时
			select {
			case <-callbackDone:
				// callback 正常完成
			case <-callbackCtx.Done():
				// callback 执行超时，不等待（防止阻塞）
				// 注意：callback 仍在执行，但不会阻塞当前 goroutine
			}
		}
	}()
}
