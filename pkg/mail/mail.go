package mail

import (
	"fmt"
	"net/smtp"

	"github.com/lvtao/go-gin-api-admin/internal/config"
)

// MailService 邮件服务
type MailService struct {
	enabled   bool
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string
}

// NewMailService 创建邮件服务
func NewMailService() *MailService {
	cfg := config.GetMailSettings()

	enabled, _ := cfg["enabled"].(bool)
	host, _ := cfg["host"].(string)
	port, _ := cfg["port"].(int)
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)
	fromEmail, _ := cfg["fromEmail"].(string)
	fromName, _ := cfg["fromName"].(string)

	return &MailService{
		enabled:   enabled,
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendEmail 发送邮件
func (s *MailService) SendEmail(to, subject, body string) error {
	if !s.enabled {
		return fmt.Errorf("mail service is not enabled")
	}

	// 构建邮件内容
	msg := fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromEmail)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-version: 1.0;\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += "\r\n"
	msg += body

	// 连接 SMTP 服务器并发送
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// GetTemplate 获取邮件模板
// 注意：这里使用默认模板，在实际实现中应该从数据库读取
func (s *MailService) GetTemplate(templateType string) (string, string, error) {
	return s.getDefaultTemplate(templateType)
}

// getDefaultTemplate 获取默认模板
func (s *MailService) getDefaultTemplate(templateType string) (string, string, error) {
	switch templateType {
	case "passwordReset":
		return "密码重置", `<p>您好，</p>
<p>您正在请求重置密码。请点击以下链接重置密码：</p>
<p><a href="{{resetLink}}">重置密码</a></p>
<p>如果无法点击，请复制以下链接到浏览器：</p>
<p>{{resetLink}}</p>
<p>该链接将在24小时后失效。</p>
<p>如果您没有请求重置密码，请忽略此邮件。</p>
<p>祝好</p>`, nil
	case "verification":
		return "邮箱验证", `<p>您好，</p>
<p>感谢您注册！请点击以下链接验证您的邮箱：</p>
<p><a href="{{verifyLink}}">验证邮箱</a></p>
<p>如果无法点击，请复制以下链接到浏览器：</p>
<p>{{verifyLink}}</p>
<p>祝好</p>`, nil
	default:
		return "", "", fmt.Errorf("unknown template type")
	}
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *MailService) SendPasswordResetEmail(to, token string) error {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset</h2>
			<p>You requested a password reset. Use the following token to reset your password:</p>
			<p><strong>%s</strong></p>
			<p>This token will expire in 1 hour.</p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>
	`, token)
	return s.SendEmail(to, subject, body)
}

// SendVerificationEmail 发送邮箱验证邮件
func (s *MailService) SendVerificationEmail(to, token string) error {
	subject := "Email Verification"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Verify Your Email</h2>
			<p>Please use the following token to verify your email address:</p>
			<p><strong>%s</strong></p>
			<p>This token will expire in 24 hours.</p>
		</body>
		</html>
	`, token)
	return s.SendEmail(to, subject, body)
}

