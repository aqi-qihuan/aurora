package util

import (
	"net"
	"strings"
)

// GetClientIP 从请求上下文中提取真实客户端 IP
// 支持: X-Forwarded-For / X-Real-Ip / RemoteAddr
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 && parts[0] != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	if xRealIP != "" {
		return xRealIP
	}
	if remoteAddr != "" {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
		return remoteAddr
	}
	return "127.0.0.1"
}

// GetIPRegion 获取IP归属地信息（简化版，不依赖ip2region库）
// 返回格式: 国家|区域|省份|城市|ISP
func GetIPRegion(ip string) string {
	ip = strings.TrimPrefix(ip, "::ffff:")
	if ip == "127.0.0.1" || ip == "localhost" || ip == "::1" {
		return "本机地址"
	}
	if ip == "0.0.0.0" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
		return "内网IP"
	}
	return "未知" // TODO: 集成ip2region后实现真实查询
}

// GetCity 从 IPRegion 字符串中提取城市名
func GetCity(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) >= 4 && parts[3] != "0" && parts[3] != "" {
		city := parts[3]
		city = strings.TrimSuffix(city, "市")
		city = strings.TrimSuffix(city, "区")
		return city
	}
	if len(parts) >= 3 && parts[2] != "0" && parts[2] != "" {
		province := parts[2]
		province = strings.TrimSuffix(province, "省")
		return province
	}
	return "未知"
}

// GetProvince 从 IPRegion 中提取省份
func GetProvince(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) >= 3 && parts[2] != "0" && parts[2] != "" {
		province := strings.TrimSuffix(parts[2], "省")
		return province
	}
	return "未知"
}

// GetISP 从 IPRegion 中提取运营商
func GetISP(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) >= 5 && parts[4] != "0" && parts[4] != "" {
		return parts[4]
	}
	return ""
}

// ResolveIP 解析IP归属地(简化版)
func ResolveIP(ip string) string {
	return GetIPRegion(ip)
}

// IsPrivateIP 判断是否为内网IP
func IsPrivateIP(ip string) bool {
	ip = strings.TrimPrefix(ip, "::ffff:")
	return ip == "127.0.0.1" || ip == "localhost" || ip == "::1" ||
		ip == "0.0.0.0" || strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.")
}
