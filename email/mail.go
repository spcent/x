package email

import (
	"net/mail"
)

// Driver 邮件发送驱动接口定义
type Driver interface {
	// Send 发送邮件
	Send(to, subject, body string) error
	// Close 关闭链接
	Close()
}

// Send 发送邮件
func Send(to, subject, body string) error {
	Lock.RLock()
	defer Lock.RUnlock()

	if Client == nil {
		return nil
	}

	return Client.Send(to, subject, body)
}

// ValidateEmail validates whether the given string is a correctly formed e-mail address.
func ValidateEmail(email string) bool {
	// Since `mail.ParseAddress` parses an email address which can also contain an optional name component,
	// here we check if incoming email string is same as the parsed email.Address. So this eliminates
	// any valid email address with name and also valid address with empty name like `<abc@example.com>`.
	em, err := mail.ParseAddress(email)
	if err != nil || em.Address != email {
		return false
	}

	return true
}
