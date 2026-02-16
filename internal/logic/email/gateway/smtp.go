package gateway

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"time"
	"unibee/internal/logic/email/sender"
	"unibee/utility"

	"github.com/gogf/gf/v2/frame/g"
)

// SuccessResponse is the synthetic response returned by SMTP send functions.
// SaveHistory checks strings.Contains(response, "202") to determine success,
// which also matches SendGrid's HTTP 202 responses.
const SuccessResponse = "202 OK"

type SmtpConfig struct {
	SmtpHost      string `json:"smtpHost"`
	SmtpPort      int    `json:"smtpPort"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	UseTLS        bool   `json:"useTLS"`
	SkipTLSVerify bool   `json:"skipTLSVerify"`
	AuthType      string `json:"authType"`
	OAuthToken    string `json:"oauthToken"`
}

type xoauth2Auth struct {
	username string
	token    string
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	resp := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.token)
	return "XOAUTH2", []byte(resp), nil
}

func (a *xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}

type loginAuth struct {
	username string
	password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch string(fromServer) {
	case "Username:", "username:":
		return []byte(a.username), nil
	case "Password:", "password:":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected LOGIN challenge: %s", fromServer)
	}
}

func SendSmtpEmailToUser(ctx context.Context, f *sender.Sender, smtpConfig *SmtpConfig, mailTo string, subject string, body string) (string, error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}

	htmlContent := "<div>" + body + " </div>"
	msg := buildMimeMessage(f, mailTo, subject, htmlContent, nil, "")

	err := sendSmtp(smtpConfig, f.Address, mailTo, msg)
	if err != nil {
		g.Log().Errorf(ctx, "SendSmtpEmailToUser error:%s", err.Error())
		return "", err
	}
	return SuccessResponse, nil
}

func SendSmtpPdfAttachEmailToUser(ctx context.Context, f *sender.Sender, smtpConfig *SmtpConfig, mailTo string, subject string, body string, pdfFilePath string, pdfFileName string) (string, error) {
	if f == nil {
		f = sender.GetDefaultSender()
	}

	dat, err := os.ReadFile(pdfFilePath)
	if err != nil {
		g.Log().Errorf(ctx, "SendSmtpPdfAttachEmailToUser read file error:%s", err.Error())
		return "", err
	}

	htmlContent := "<div>" + body + " </div>"
	msg := buildMimeMessage(f, mailTo, subject, htmlContent, dat, pdfFileName)

	err = sendSmtp(smtpConfig, f.Address, mailTo, msg)
	if err != nil {
		g.Log().Errorf(ctx, "SendSmtpPdfAttachEmailToUser error:%s", err.Error())
		return "", err
	}
	return SuccessResponse, nil
}

func buildMimeMessage(f *sender.Sender, mailTo string, subject string, htmlContent string, attachData []byte, attachName string) []byte {
	var buf bytes.Buffer
	sanitizer := strings.NewReplacer("\r", "", "\n", "")
	mailTo = sanitizer.Replace(mailTo)
	fromAddr := sanitizer.Replace(f.Address)
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	safeName := sanitizer.Replace(f.Name)

	if len(attachData) == 0 {
		buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", safeName, fromAddr))
		buf.WriteString(fmt.Sprintf("To: %s\r\n", mailTo))
		buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(htmlContent)
		return buf.Bytes()
	}

	boundary := fmt.Sprintf("===============%d==", time.Now().UnixNano())
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", safeName, fromAddr))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", mailTo))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlContent)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", "application/pdf")
	header.Set("Content-Transfer-Encoding", "base64")
	safeAttachName := strings.NewReplacer("\r", "", "\n", "", "\"", "").Replace(attachName)
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeAttachName))
	for k, v := range header {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	buf.WriteString("\r\n")

	encoded := base64.StdEncoding.EncodeToString(attachData)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[i:end] + "\r\n")
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buf.Bytes()
}

func sendSmtp(config *SmtpConfig, from string, to string, msg []byte) error {
	// Re-validate host at send time to prevent DNS rebinding attacks.
	// The host was validated at setup time, but DNS records may have changed
	// to point to internal addresses since then.
	if err := utility.ValidateExternalHost(config.SmtpHost); err != nil {
		return fmt.Errorf("smtp host validation failed: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)

	tlsConfig := &tls.Config{
		ServerName:         config.SmtpHost,
		InsecureSkipVerify: config.SkipTLSVerify,
	}

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}

	client, err := smtp.NewClient(conn, config.SmtpHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp new client error: %w", err)
	}
	// quitted tracks whether Quit was called; Close is the fallback for error paths.
	quitted := false
	defer func() {
		if !quitted {
			client.Close()
		}
	}()

	if config.UseTLS {
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls error: %w", err)
		}
	}

	var auth smtp.Auth
	switch config.AuthType {
	case "plain", "":
		auth = smtp.PlainAuth("", config.Username, config.Password, config.SmtpHost)
	case "login":
		auth = &loginAuth{username: config.Username, password: config.Password}
	case "cram-md5":
		auth = smtp.CRAMMD5Auth(config.Username, config.Password)
	case "xoauth2":
		auth = &xoauth2Auth{username: config.Username, token: config.OAuthToken}
	case "none":
		auth = nil
	default:
		return fmt.Errorf("unsupported smtp auth type: %s", config.AuthType)
	}
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth error: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail error: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data error: %w", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp write error: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("smtp close error: %w", err)
	}

	quitted = true
	return client.Quit()
}
