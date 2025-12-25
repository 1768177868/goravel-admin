package unit

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

// IPMatcherTestSuite IP 匹配器测试套件
type IPMatcherTestSuite struct {
	suite.Suite
}

func TestIPMatcherTestSuite(t *testing.T) {
	suite.Run(t, new(IPMatcherTestSuite))
}

func (s *IPMatcherTestSuite) SetupTest() {
}

func (s *IPMatcherTestSuite) TearDownTest() {
}

// isIPInRange 检查 IP 是否在指定范围内
func isIPInRange(ip, startIP, endIP net.IP) bool {
	ipBytes := ip.To4()
	startBytes := startIP.To4()
	endBytes := endIP.To4()

	if ipBytes == nil || startBytes == nil || endBytes == nil {
		return false
	}

	ipInt := uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3])
	startInt := uint32(startBytes[0])<<24 | uint32(startBytes[1])<<16 | uint32(startBytes[2])<<8 | uint32(startBytes[3])
	endInt := uint32(endBytes[0])<<24 | uint32(endBytes[1])<<16 | uint32(endBytes[2])<<8 | uint32(endBytes[3])

	return ipInt >= startInt && ipInt <= endInt
}

// TestIsIPInRange_InRange 测试 IP 在范围内
func (s *IPMatcherTestSuite) TestIsIPInRange_InRange() {
	ip := net.ParseIP("192.168.1.50")
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.True(isIPInRange(ip, startIP, endIP))
}

// TestIsIPInRange_AtStart 测试 IP 在范围起点
func (s *IPMatcherTestSuite) TestIsIPInRange_AtStart() {
	ip := net.ParseIP("192.168.1.1")
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.True(isIPInRange(ip, startIP, endIP))
}

// TestIsIPInRange_AtEnd 测试 IP 在范围终点
func (s *IPMatcherTestSuite) TestIsIPInRange_AtEnd() {
	ip := net.ParseIP("192.168.1.100")
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.True(isIPInRange(ip, startIP, endIP))
}

// TestIsIPInRange_OutOfRange 测试 IP 超出范围
func (s *IPMatcherTestSuite) TestIsIPInRange_OutOfRange() {
	ip := net.ParseIP("192.168.1.101")
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.False(isIPInRange(ip, startIP, endIP))
}

// TestIsIPInRange_BeforeRange 测试 IP 在范围之前
func (s *IPMatcherTestSuite) TestIsIPInRange_BeforeRange() {
	ip := net.ParseIP("192.168.0.255")
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.False(isIPInRange(ip, startIP, endIP))
}

// TestIsIPInRange_InvalidIP 测试无效 IP
func (s *IPMatcherTestSuite) TestIsIPInRange_InvalidIP() {
	startIP := net.ParseIP("192.168.1.1")
	endIP := net.ParseIP("192.168.1.100")

	s.False(isIPInRange(nil, startIP, endIP))
}

// TestCIDRMatch 测试 CIDR 匹配
func (s *IPMatcherTestSuite) TestCIDRMatch() {
	tests := []struct {
		name  string
		ip    string
		cidr  string
		match bool
	}{
		{"匹配 /24", "192.168.1.100", "192.168.1.0/24", true},
		{"不匹配 /24", "192.168.2.1", "192.168.1.0/24", false},
		{"/32 精确匹配", "192.168.1.1", "192.168.1.1/32", true},
		{"/32 不匹配", "192.168.1.2", "192.168.1.1/32", false},
		{"/16 匹配", "192.168.100.50", "192.168.0.0/16", true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ip := net.ParseIP(tt.ip)
			_, ipNet, _ := net.ParseCIDR(tt.cidr)

			if tt.match {
				s.True(ipNet.Contains(ip))
			} else {
				s.False(ipNet.Contains(ip))
			}
		})
	}
}

// TestIPParsing 测试 IP 解析
func (s *IPMatcherTestSuite) TestIPParsing() {
	tests := []struct {
		name  string
		ip    string
		valid bool
	}{
		{"有效 IPv4", "192.168.1.1", true},
		{"有效 IPv4 边界", "255.255.255.255", true},
		{"有效 IPv4 零", "0.0.0.0", true},
		{"无效 IP", "invalid-ip", false},
		{"空字符串", "", false},
		{"超出范围", "256.1.1.1", false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ip := net.ParseIP(tt.ip)
			if tt.valid {
				s.NotNil(ip)
			} else {
				s.Nil(ip)
			}
		})
	}
}

