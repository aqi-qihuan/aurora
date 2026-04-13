package util

import (
	"net"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// GetClientIP 从Gin Context中提取真实客户端 IP
// 支持: X-Forwarded-For / X-Real-Ip / RemoteAddr
func GetClientIP(c interface{}) string {
	// 使用类型断言获取gin.Context
	type GinContext interface {
		GetHeader(key string) string
		RemoteIP() net.IP
	}
	
	if ctx, ok := c.(GinContext); ok {
		// 优先从X-Forwarded-For获取
		if xff := ctx.GetHeader("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 && parts[0] != "" {
				return strings.TrimSpace(parts[0])
			}
		}
		
		// 其次从X-Real-Ip获取
		if xri := ctx.GetHeader("X-Real-Ip"); xri != "" {
			return strings.TrimSpace(xri)
		}
		
		// 最后从RemoteIP获取
		if remoteIP := ctx.RemoteIP(); remoteIP != nil {
			return remoteIP.String()
		}
	}
	
	return "127.0.0.1"
}

// ip2region 全局查询器 (v3版本使用 .xdb 格式)
var xdbSearcher *xdb.Searcher

// InitIP2Region 初始化 ip2region 查询器 (应用启动时调用)
func InitIP2Region(dbFile string) error {
	// 加载 .xdb 文件到内存
	data, err := xdb.LoadContentFromFile(dbFile)
	if err != nil {
		return err
	}
	
	// 创建基于内存的查询器 (速度快，适合生产环境)
	// v3版本需要指定IP版本（IPv4）
	xdbSearcher, err = xdb.NewWithBuffer(xdb.IPv4, data)
	if err != nil {
		return err
	}
	
	return nil
}

// GetIPRegion 获取IP归属地信息 (集成 ip2region)
// 返回格式: 国家|区域|省份|城市|ISP (对标Java版)
func GetIPRegion(ip string) string {
	ip = strings.TrimPrefix(ip, "::ffff:")
	if ip == "127.0.0.1" || ip == "localhost" || ip == "::1" {
		return "本机地址"
	}
	if ip == "0.0.0.0" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
		return "内网IP"
	}
	
	// 使用 ip2region 查询
	if xdbSearcher != nil {
		region, err := xdbSearcher.Search(ip)
		if err == nil && region != "" {
			return region
		}
	}
	
	return "未知"
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

// ParseBrowser 从 User-Agent 字符串中解析浏览器名称（对标Java UserAgentUtils）
func ParseBrowser(userAgent string) string {
	if userAgent == "" {
		return "未知"
	}
	
	ua := strings.ToLower(userAgent)
	
	// 检测顺序很重要，避免误判
	if strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/") {
		return "Edge"
	} else if strings.Contains(ua, "chrome/") && !strings.Contains(ua, "chromium/") {
		return "Chrome"
	} else if strings.Contains(ua, "firefox/") {
		return "Firefox"
	} else if strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/") {
		return "Safari"
	} else if strings.Contains(ua, "opera/") || strings.Contains(ua, "opr/") {
		return "Opera"
	} else if strings.Contains(ua, "msie ") || strings.Contains(ua, "trident/") {
		return "IE"
	}
	
	return "其他"
}

// ParseOS 从 User-Agent 字符串中解析操作系统（对标Java UserAgentUtils）
func ParseOS(userAgent string) string {
	if userAgent == "" {
		return "未知"
	}
	
	ua := strings.ToLower(userAgent)
	
	// Windows 系列
	if strings.Contains(ua, "windows nt 10.0") {
		return "Windows 10"
	} else if strings.Contains(ua, "windows nt 6.3") {
		return "Windows 8.1"
	} else if strings.Contains(ua, "windows nt 6.2") {
		return "Windows 8"
	} else if strings.Contains(ua, "windows nt 6.1") {
		return "Windows 7"
	} else if strings.Contains(ua, "windows nt 6.0") {
		return "Windows Vista"
	} else if strings.Contains(ua, "windows nt 5.1") || strings.Contains(ua, "windows xp") {
		return "Windows XP"
	} else if strings.Contains(ua, "windows") {
		return "Windows"
	}
	
	// macOS / Mac OS X
	if strings.Contains(ua, "mac os x") {
		return "macOS"
	}
	
	// Linux 系列
	if strings.Contains(ua, "linux") && strings.Contains(ua, "android") {
		return "Android"
	} else if strings.Contains(ua, "linux") {
		return "Linux"
	}
	
	// iOS / iPadOS
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return "iOS"
	}
	
	return "其他"
}
