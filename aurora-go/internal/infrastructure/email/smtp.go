package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"time"

	"github.com/aurora-go/aurora/internal/config"
)

// EmailService 邮件发送服务（对标 Java 版 EmailUtils + 邮件模板）
type EmailService struct {
	auth    smtp.Auth
	addr    string
	from    string
	cfg     *config.EmailConfig
}

var emailService *EmailService

// InitEmailService 初始化邮件服务（延迟初始化，按需连接）
func InitEmailService(cfg *config.EmailConfig) {
	if !cfg.Enabled {
		slog.Info("Email service disabled, skipping initialization")
		return
	}

	emailService = &EmailService{
		auth: smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
		addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		from:  fmt.Sprintf("%s<%s>", cfg.FromName, cfg.Username),
		cfg:   cfg,
	}

	slog.Info("Email service initialized",
		"host", cfg.Host,
		"from", emailService.from,
		"ssl", cfg.IsSSL,
	)
}

// SendCommentNotification 发送评论通知邮件
// 对标 Java 版 EmailConsumer: 评论邮件异步通知
func SendCommentNotification(toEmail string, subject string, data CommentNotifyData) error {
	if emailService == nil || !emailService.cfg.Enabled {
		return fmt.Errorf("email service not enabled")
	}

	body, err := renderTemplate(commentNotifyTmpl, data)
	if err != nil {
		return err
	}

	return send(toEmail, "[Aurora] "+subject, body)
}

// SendSubscribeNotification 发送文章订阅通知邮件
// 对标 Java 版 SubscribeConsumer: 新文章订阅推送
func SendSubscribeNotification(toEmail string, subject string, data SubscribeNotifyData) error {
	if emailService == nil || !emailService.cfg.Enabled {
		return fmt.Errorf("email service not enabled")
	}

	body, err := renderTemplate(subscribeNotifyTmpl, data)
	if err != nil {
		return err
	}

	return send(toEmail, "[Aurora] "+subject, body)
}

// SendVerificationCode 发送验证码邮件（注册/OAuth绑定邮箱）
func SendVerificationCode(toEmail string, code string) error {
	if emailService == nil || !emailService.cfg.Enabled {
		return fmt.Errorf("email service not enabled")
	}

	data := VerificationCodeData{
		Code:      code,
		ExpireMin: time.Duration(emailService.cfg.VerifyCodeExpire) * time.Minute,
	}
	body, err := renderTemplate(verificationCodeTmpl, data)
	if err != nil {
		return err
	}

	return send(toEmail, "[Aurora] 邮箱验证码 - 请勿泄露给他人", body)
}

// send 底层SMTP发送方法
func send(toEmail string, subject string, htmlBody string) error {
	headers := make(map[string]string)
	headers["From"] = emailService.from
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = `text/html; charset="UTF-8"`
	headers["Date"] = time.Now().Format(time.RFC1123)

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(k)
		msg.WriteString(": ")
		msg.WriteString(v)
		msg.WriteString("\r\n")
	}
	msg.WriteString("\r\n") // header/body分隔符
	msg.WriteString(htmlBody)

	err := smtp.SendMail(emailService.addr, emailService.auth, emailService.cfg.Username, []string{toEmail}, msg.Bytes())
	if err != nil {
		slog.Error("Failed to send email",
			"to", toEmail,
			"subject", subject,
			"error", err,
		)
		return err
	}

	slog.Debug("Email sent successfully", "to", toEmail, "subject", subject)
	return nil
}

// renderTemplate 渲染HTML邮件模板
func renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ==================== 数据结构定义 ====================

// CommentNotifyData 评论通知数据
type CommentNotifyData struct {
	Nickname       string // 评论者昵称
	ArticleTitle   string // 被评论的文章标题
	Content        string // 评论内容摘要
	ArticleURL     string // 文章链接
	SiteName       string // 站点名称
	CurrentYear    int    // 年份
}

// SubscribeNotifyData 订阅通知数据
type SubscribeNotifyData struct {
	ArticleTitle   string // 新文章标题
	ArticleSummary string // 文章摘要
	ArticleURL     string // 文章链接
	SiteName       string // 站点名称
	UnsubscribeURL string // 取消订阅链接
	CurrentYear    int
}

// VerificationCodeData 验证码数据
type VerificationCodeData struct {
	Code      string
	ExpireMin time.Duration
}

// ==================== HTML 邮件模板 ====================
// 对标 Java 版 resources/email/ 目录下的 .ftl 模板

const commentNotifyTmpl = `
<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f5f7fa;color:#333;margin:0;padding:20px}.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,.08)}.header{background:linear-gradient(135deg,#667eea,#764ba2);padding:30px;text-align:center}.header h1{color:#fff;margin:0;font-size:22px}.content{padding:30px}.comment-card{background:#f8f9fc;border-left:4px solid #667eea;padding:16px;margin:20px 0;border-radius:0 4px 4px 0}.comment-card .author{font-weight:600;color:#667eea;font-size:14px}.comment-card .text{margin-top:8px;line-height:1.6;color:#555}.btn{display:inline-block;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;padding:12px 28px;border-radius:25px;text-decoration:none;font-weight:500;margin-top:20px}.footer{text-align:center;padding:20px;color:#999;font-size:12px;border-top:1px solid #eee}
</style></head><body><div class="container"><div class="header"><h1>{{.SiteName}} - 新评论通知</h1></div><div class="content"><p>您好，您的文章收到了新评论：</p><div class="comment-card"><div class="author">{{.Nickname}}</div><div class="text">{{.Content}}</div></div><p>文章标题：<strong>{{.ArticleTitle}}</strong></p><a href="{{.ArticleURL}}" class="btn">查看完整评论</a></div><div class="footer"><p>&copy; {{.CurrentYear}} {{.SiteName}} · 此邮件由系统自动发送</p></div></div></body></html>`

const subscribeNotifyTmpl = `
<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f5f7fa;color:#333;margin:0;padding:20px}.container{max-width:600px;margin:0 auto;background:#fff;border-radius:8px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,.08)}.header{background:linear-gradient(135deg,#11998e,#38ef7d);padding:30px;text-align:center}.header h1{color:#fff;margin:0;font-size:22px}.content{padding:30px}.article-title{font-size:18px;font-weight:600;color:#333;margin-bottom:12px}.article-summary{line-height:1.6;color:#666;background:#f8f9fc;padding:16px;border-radius:6px;margin:16px 0}.btn{display:inline-block;background:linear-gradient(135deg,#11998e,#38ef7d);color:#fff;padding:12px 28px;border-radius:25px;text-decoration:none;font-weight:500;margin-top:12px}.footer{text-align:center;padding:20px;color:#999;font-size:12px;border-top:1px solid #eee}
</style></head><body><div class="container"><div class="header"><h1>{{.SiteName}} - 新文章发布</h1></div><div class="content"><p>您好！您订阅的博客发布了新文章：</p><div class="article-title">{{.ArticleTitle}}</div><div class="article-summary">{{.ArticleSummary}}</div><a href="{{.ArticleURL}}" class="btn">阅读全文</a><br><a href="{{.UnsubscribeURL}}" style="font-size:12px;color:#999;margin-top:20px;display:inline-block">取消订阅</a></div><div class="footer"><p>&copy; {{.CurrentYear}} {{.SiteName}} · 此邮件由系统自动发送</p></div></div></body></html>`

const verificationCodeTmpl = `
<!DOCTYPE html><html><head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f5f7fa;color:#333;margin:0;padding:20px;display:flex;align-items:center;justify-content:center;min-height:100vh}.container{max-width:480px;background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.1)}.header{background:linear-gradient(135deg,#f093fb,#f5576c);padding:40px;text-align:center}.code{font-size:36px;font-weight:700;letter-spacing:8px;color:#fff;text-shadow:0 2px 8px rgba(0,0,0,.15)}.content{padding:40px;text-align:center}.expire-notice{color:#999;font-size:13px;margin-top:16px}
</style></head><body><div class="container"><div class="header"><div class="code">{{.Code}}</div></div><div class="content"><p style="font-size:16px;color:#333">您的验证码如上所示</p><p style="color:#666">验证码将在 <strong>{{printf "%.0f" .ExpireMin}} 分钟</strong>后过期</p><p class="expire-notice">如果这不是您的操作，请忽略此邮件</p></div></div></body></html>`
