package banner

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceProbe 外部服务探针信息
type ServiceProbe struct {
	Name     string
	Check    func() bool
	Timeout  time.Duration
}

// PrintStartupBanner 打印启动面板
func PrintStartupBanner(mode string, port int, startTime time.Time, probes []ServiceProbe, swaggerEnabled bool) {
	hostname, _ := os.Hostname()
	elapsed := time.Since(startTime).Round(10 * time.Millisecond)

	// 获取本机 IP
	localIP := getLocalIP()
	addr := fmt.Sprintf("http://%s:%d", localIP, port)

	// 计算面板宽度
	const width = 62

	// === 顶部标题面板 ===
	printBox("Aurora Go 启动完成", []row{
		{"hostname", hostname},
		{"mode", mode},
		{"port", fmt.Sprintf("%d", port)},
		{"address", addr},
		{"elapsed", fmt.Sprintf("%v", elapsed)},
	}, width)

	fmt.Println()

	// === API & Swagger ===
	infoRows := []row{
		{"API Base", fmt.Sprintf("http://%s:%d/api", localIP, port)},
		{"Health", fmt.Sprintf("http://%s:%d/health", localIP, port)},
	}
	if swaggerEnabled {
		infoRows = append([]row{{"Swagger", fmt.Sprintf("http://%s:%d/swagger/index.html", localIP, port)}}, infoRows...)
	}
	printBox("", infoRows, width)

	fmt.Println()

	// === 外部服务诊断 ===
	rows := make([]row, 0, len(probes))
	for _, p := range probes {
		icon := "❌ FAIL"
		if p.Check() {
			icon = "✅ OK"
		}
		rows = append(rows, row{p.Name, icon})
	}
	printBox("外部服务诊断", rows, width)
	fmt.Println()
}

type row struct {
	key   string
	value string
}

func printBox(title string, rows []row, width int) {
	border := strings.Repeat("─", width)

	// 标题行
	if title == "" {
		fmt.Printf("┌%s┐\n", border)
	} else {
		fmt.Printf("┌─ %s %s┐\n", title, strings.Repeat("─", width-3-len(title)))
	}

	// 内容行
	for _, r := range rows {
		left := fmt.Sprintf("│  %-10s", r.key)
		padding := width - len(left) - len(r.value) - 2 // -2 for "│ " and "│"
		if padding < 1 {
			padding = 1
		}
		fmt.Printf("%s%s%s│\n", left, strings.Repeat(" ", padding), r.value)
	}

	fmt.Printf("└%s┘\n", border)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// MySQLProbe 创建 MySQL 探针
func MySQLProbe(db *gorm.DB) ServiceProbe {
	return ServiceProbe{
		Name:    "MySQL",
		Timeout: 3 * time.Second,
		Check: func() bool {
			if db == nil {
				return false
			}
			sqlDB, err := db.DB()
			if err != nil {
				return false
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			return sqlDB.PingContext(ctx) == nil
		},
	}
}

// RedisProbe 创建 Redis 探针
func RedisProbe(rdb *redis.Client) ServiceProbe {
	return ServiceProbe{
		Name:    "Redis",
		Timeout: 2 * time.Second,
		Check: func() bool {
			if rdb == nil {
				return false
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			return rdb.Ping(ctx).Err() == nil
		},
	}
}
