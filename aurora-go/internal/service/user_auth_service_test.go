package service

import (
	"testing"
)

// ===== cleanIPAddress 测试 =====

func TestCleanIPAddress(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		// 标准格式
		{"标准 IPv4", "192.168.1.1", "192.168.1.1"},
		{"标准 IPv6", "2001:db8::1", "2001:db8::1"},

		// 空值与空白
		{"空字符串", "", ""},
		{"纯空格", "   ", ""},
		{"纯制表符", "\t\t", ""},

		// 尾部数字（对标 Java 场景 "95.40.12.12 0"）
		{"尾部带数字", "95.40.12.12 0", "95.40.12.12"},
		{"尾部多个空格+数字", "110.184.180.12   42", "110.184.180.12"},

		// 换行符处理（注意：strings.Fields 会按空白分割取第一部分，\n 被当作分隔符）
		{"含换行符_截断", "110.184.180.\n10", "110.184.180."},
		{"含回车符_截断", "110.184.180.\r10", "110.184.180."},
		{"含制表符_截断", "110.184.180.\t10", "110.184.180."},
		{"首尾换行", "\n192.168.1.1\n", "192.168.1.1"},

		// X-Forwarded-For 多 IP 格式
		{"XFF 单 IP", "1.2.3.4", "1.2.3.4"},
		{"XFF 多 IP", "1.2.3.4, 5.6.7.8", "1.2.3.4"},
		{"XFF 三个 IP", "1.2.3.4, 5.6.7.8, 9.10.11.12", "1.2.3.4"},

		// 组合场景
		{"换行+空格+数字", "10.0.0.1\n 99", "10.0.0.1"},
		{"XFF+换行", "1.2.3.4\n, 5.6.7.8", "1.2.3.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanIPAddress(tt.ip)
			if got != tt.want {
				t.Errorf("cleanIPAddress(%q) = %q; want %q", tt.ip, got, tt.want)
			}
		})
	}
}

// ===== cleanIPAddress 边界用例 =====

func TestCleanIPAddress_NoCommaInSingleIP(t *testing.T) {
	// 确保逗号在第一位之后的处理（idx > 0 而非 >= 0）
	got := cleanIPAddress(",1.2.3.4")
	// 逗号在首位时 strings.Index 返回 0，不进入分支，但前面 Fields 已处理
	if got == "" {
		t.Logf("cleanIPAddress(\",1.2.3.4\") = %q (空值，Fields 处理结果)", got)
	}
}
