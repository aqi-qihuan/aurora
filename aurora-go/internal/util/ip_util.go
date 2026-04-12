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
// 对标 Java IpUtil.getIpProvince()
// ipSource 格式: 国家|区域|省份|城市|运营商 (ip2region 原始输出)
// Java 的 getIpSource() 会先去掉 "|0" 和 "0|"，所以格式可能变为:
//   - 中国|四川|成都市|电信 (区域=省名)
//   - 中国|0|四川|南阳市|电信 (区域=0, 省份=省名)
//   - 中国|南阳市|电信 (无省份信息，区域=城市名)
func GetProvince(region string) string {
	if region == "" {
		return "未知"
	}
	parts := strings.Split(region, "|")
	
	// 策略1: 如果索引1以"省"结尾，取之（对标 Java strings[1].endsWith("省")）
	if len(parts) > 1 && parts[1] != "" && parts[1] != "0" && strings.HasSuffix(parts[1], "省") {
		return strings.TrimSuffix(parts[1], "省")
	}
	
	// 策略2: 检查索引2（省份字段），排除以"市"结尾的
	if len(parts) > 2 && parts[2] != "" && parts[2] != "0" && !strings.HasSuffix(parts[2], "市") {
		return strings.TrimSuffix(parts[2], "省")
	}
	
	// 策略3: 回退到索引1（区域），但排除城市名和国家名
	if len(parts) > 1 && parts[1] != "" && parts[1] != "0" && !strings.HasSuffix(parts[1], "市") && parts[1] != "中国" {
		return parts[1]
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
