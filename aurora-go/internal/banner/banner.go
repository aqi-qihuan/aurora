package banner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aurora-go/aurora/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceResult 单个服务检测结果
type ServiceResult struct {
	Name    string
	Status  string // "✅ OK" / "❌ FAIL" / "⏭ SKIP"
	Latency string
}

// PrintStartupBanner 打印启动面板
func PrintStartupBanner(cfg *config.Config, startTime time.Time, db *gorm.DB, rdb *redis.Client, swaggerEnabled bool) {
	hostname, _ := os.Hostname()
	elapsed := time.Since(startTime).Round(10 * time.Millisecond)
	localIP := getLocalIP()
	port := cfg.Server.Port

	width := 62

	// === 标题面板 ===
	printBox("Aurora Go 启动完成", []row{
		{"hostname", hostname},
		{"mode", cfg.Server.Mode},
		{"port", fmt.Sprintf("%d", port)},
		{"address", fmt.Sprintf("http://%s:%d", localIP, port)},
		{"elapsed", fmt.Sprintf("%v", elapsed)},
	}, width)

	fmt.Println()

	// === API 端点 ===
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
	results := detectAll(cfg, db, rdb)
	rows := make([]row, len(results))
	for i, r := range results {
		latencyStr := r.Latency
		if r.Status == "⏭ SKIP" {
			latencyStr = "未配置"
		}
		rows[i] = row{r.Name, fmt.Sprintf("%s  (%s)", r.Status, latencyStr)}
	}
	printBox("外部服务诊断", rows, width)
	fmt.Println()
}

type row struct {
	key   string
	value string
}

func printBox(title string, rows []row, width int) {
	if title == "" {
		fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
	} else {
		fmt.Printf("┌─ %s %s┐\n", title, strings.Repeat("─", width-3-len(title)))
	}

	for _, r := range rows {
		left := fmt.Sprintf("│  %-10s", r.key)
		padding := width - len(left) - len(r.value) - 2
		if padding < 1 {
			padding = 1
		}
		fmt.Printf("%s%s│\n", left, strings.Repeat(" ", padding)+r.value)
	}

	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}

func getLocalIP() string {
	conns, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53})
	if err == nil {
		defer conns.Close()
		localAddr := conns.LocalAddr().(*net.UDPAddr)
		ip := localAddr.IP.String()
		if !strings.HasPrefix(ip, "169.254.") && !strings.HasPrefix(ip, "127.") {
			return ip
		}
	}

	// 遍历所有接口，优先找 192.168.x.x
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				s := ip4.String()
				if strings.HasPrefix(s, "192.168.") || strings.HasPrefix(s, "10.") {
					return s
				}
			}
		}
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil && !strings.HasPrefix(ip4.String(), "169.254.") {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}

func detectAll(cfg *config.Config, db *gorm.DB, rdb *redis.Client) []ServiceResult {
	var results []ServiceResult

	// MySQL
	results = append(results, detectService("MySQL", cfg.MySQL.Host, func() (time.Duration, error) {
		if db == nil {
			return 0, fmt.Errorf("not initialized")
		}
		sqlDB, err := db.DB()
		if err != nil {
			return 0, err
		}
		start := time.Now()
		err = sqlDB.Ping()
		return time.Since(start), err
	}))

	// Redis
	results = append(results, detectService("Redis", cfg.Redis.Host, func() (time.Duration, error) {
		if rdb == nil {
			return 0, fmt.Errorf("not initialized")
		}
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := rdb.Ping(ctx).Err()
		return time.Since(start), err
	}))

	// RabbitMQ
	results = append(results, detectService("RabbitMQ", cfg.RabbitMQ.Host, func() (time.Duration, error) {
		if cfg.RabbitMQ.Host == "" {
			return 0, nil
		}
		addr := fmt.Sprintf("%s:%d", cfg.RabbitMQ.Host, cfg.RabbitMQ.Port)
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			return 0, err
		}
		defer conn.Close()
		return time.Since(start), nil
	}))

	// Elasticsearch
	results = append(results, detectService("Elasticsearch", cfg.ES.GetPrimaryURL(), func() (time.Duration, error) {
		if len(cfg.ES.URLs) == 0 || cfg.ES.URLs[0] == "" {
			return 0, nil
		}
		parsedURL, err := url.Parse(cfg.ES.GetPrimaryURL())
		if err != nil {
			return 0, err
		}
		esAddr := parsedURL.Host
		start := time.Now()
		resp, err := http.Get(fmt.Sprintf("http://%s/", esAddr))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("status %d", resp.StatusCode)
		}
		return time.Since(start), nil
	}))

	// MinIO
	results = append(results, detectService("MinIO", cfg.MinIO.Endpoint, func() (time.Duration, error) {
		if cfg.MinIO.Endpoint == "" {
			return 0, nil
		}
		start := time.Now()
		resp, err := http.Get(fmt.Sprintf("%s/minio/health/live", cfg.MinIO.Endpoint))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("status %d", resp.StatusCode)
		}
		return time.Since(start), nil
	}))

	// Email (SMTP)
	results = append(results, detectService("Email", cfg.Email.Host, func() (time.Duration, error) {
		if !cfg.Email.Enabled || cfg.Email.Host == "" {
			return 0, nil
		}
		addr := fmt.Sprintf("%s:%d", cfg.Email.Host, cfg.Email.Port)
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			return 0, err
		}
		defer conn.Close()
		return time.Since(start), nil
	}))

	return results
}

func detectService(name, host string, check func() (time.Duration, error)) ServiceResult {
	if host == "" {
		return ServiceResult{Name: name, Status: "⏭ SKIP", Latency: "未配置"}
	}

	latency, err := check()
	if err != nil {
		return ServiceResult{Name: name, Status: "❌ FAIL", Latency: err.Error()}
	}
	return ServiceResult{Name: name, Status: "✅ OK", Latency: latency.Round(time.Millisecond).String()}
}

// MySQLProbe 创建 MySQL 探针 (供外部单独使用)
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
			return sqlDB.Ping() == nil
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

// ServiceProbe 外部服务探针信息
type ServiceProbe struct {
	Name    string
	Check   func() bool
	Timeout time.Duration
}
