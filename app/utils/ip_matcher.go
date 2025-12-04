package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/goravel/framework/support/str"
)

// IsIPInBlacklist 检查IP是否在黑名单中
// ip: 要检查的IP地址
// blacklistIPs: 黑名单IP字符串，支持：
//   - 单个IP: 192.168.1.1
//   - 多个IP（逗号分隔）: 192.168.1.1,192.168.1.2
//   - CIDR格式: 192.168.0.0/24
//   - IP范围: 192.168.1.1-192.168.1.100
func IsIPInBlacklist(ip string, blacklistIPs string) bool {
	if blacklistIPs == "" {
		return false
	}

	// 解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 分割多个IP（支持逗号分隔）
	ipList := strings.SplitSeq(blacklistIPs, ",")
	for blacklistIP := range ipList {
		blacklistIP = str.Of(blacklistIP).Trim().String()
		if str.Of(blacklistIP).IsEmpty() {
			continue
		}

		// 检查单个IP
		if blacklistIP == ip {
			return true
		}

		// 检查CIDR格式 (192.168.0.0/24)
		if str.Of(blacklistIP).Contains("/") {
			_, ipNet, err := net.ParseCIDR(blacklistIP)
			if err == nil && ipNet.Contains(parsedIP) {
				return true
			}
		}

		// 检查IP范围格式 (192.168.1.1-192.168.1.100)
		if str.Of(blacklistIP).Contains("-") {
			parts := str.Of(blacklistIP).Split("-")
			if len(parts) == 2 {
				startIP := net.ParseIP(str.Of(parts[0]).Trim().String())
				endIP := net.ParseIP(str.Of(parts[1]).Trim().String())
				if startIP != nil && endIP != nil {
					if isIPInRange(parsedIP, startIP, endIP) {
						return true
					}
				}
			}
		}
	}

	return false
}

// isIPInRange 检查IP是否在指定范围内
func isIPInRange(ip, startIP, endIP net.IP) bool {
	ipBytes := ip.To4()
	startBytes := startIP.To4()
	endBytes := endIP.To4()

	if ipBytes == nil || startBytes == nil || endBytes == nil {
		return false
	}

	// 将IP地址转换为32位整数进行比较
	ipInt := uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3])
	startInt := uint32(startBytes[0])<<24 | uint32(startBytes[1])<<16 | uint32(startBytes[2])<<8 | uint32(startBytes[3])
	endInt := uint32(endBytes[0])<<24 | uint32(endBytes[1])<<16 | uint32(endBytes[2])<<8 | uint32(endBytes[3])

	return ipInt >= startInt && ipInt <= endInt
}

// ValidateBlacklistIP 验证黑名单IP格式
// 返回错误信息，如果格式正确返回空字符串
func ValidateBlacklistIP(ipStr string) string {
	if ipStr == "" {
		return "IP地址不能为空"
	}

	ipList := strings.SplitSeq(ipStr, ",")
	for ip := range ipList {
		ip = str.Of(ip).Trim().String()
		if str.Of(ip).IsEmpty() {
			continue
		}

		// 检查CIDR格式
		if str.Of(ip).Contains("/") {
			_, _, err := net.ParseCIDR(ip)
			if err != nil {
				return fmt.Sprintf("CIDR格式错误: %s", ip)
			}
			continue
		}

		// 检查IP范围格式
		if str.Of(ip).Contains("-") {
			parts := str.Of(ip).Split("-")
			if len(parts) != 2 {
				return fmt.Sprintf("IP范围格式错误: %s (格式应为: 192.168.1.1-192.168.1.100)", ip)
			}
			startIP := net.ParseIP(str.Of(parts[0]).Trim().String())
			endIP := net.ParseIP(str.Of(parts[1]).Trim().String())
			if startIP == nil {
				return fmt.Sprintf("起始IP格式错误: %s", parts[0])
			}
			if endIP == nil {
				return fmt.Sprintf("结束IP格式错误: %s", parts[1])
			}
			// 验证范围是否有效（结束IP应该大于等于起始IP）
			if !isIPInRange(endIP, startIP, endIP) && !endIP.Equal(startIP) {
				// 简单比较：检查结束IP是否大于起始IP
				startBytes := startIP.To4()
				endBytes := endIP.To4()
				if startBytes != nil && endBytes != nil {
					for i := range 4 {
						if endBytes[i] < startBytes[i] {
							return fmt.Sprintf("结束IP必须大于等于起始IP: %s", ip)
						}
						if endBytes[i] > startBytes[i] {
							break
						}
					}
				}
			}
			continue
		}

		// 检查单个IP格式
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return fmt.Sprintf("IP地址格式错误: %s", ip)
		}
	}

	return ""
}

// FormatIPRange 格式化IP范围显示
func FormatIPRange(ipStr string) string {
	if str.Of(ipStr).Contains("-") {
		parts := str.Of(ipStr).Split("-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s ~ %s", str.Of(parts[0]).Trim().String(), str.Of(parts[1]).Trim().String())
		}
	}
	return ipStr
}
