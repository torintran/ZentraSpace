package app

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const platformSettingsScope = "!zentraspace-platform-settings"

type smtpConfiguration struct {
	Host, Port, Sender, Mode, Username, Password, Revision string
	Disabled                                               bool
}

var smtpHost = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{0,252}$`)

func (c smtpConfiguration) validate() error {
	_, ipErr := netip.ParseAddr(c.Host)
	port, portErr := strconv.Atoi(c.Port)
	if (ipErr != nil && !smtpHost.MatchString(c.Host)) || portErr != nil || port < 1 || port > 65535 ||
		!validEmail(c.Sender) || (c.Mode != "tls" && c.Mode != "starttls") {
		return errors.New("Kiểm tra máy chủ, cổng, địa chỉ gửi và kiểu mã hóa SMTP.")
	}
	if len(c.Username) > 254 || len(c.Password) > 1024 || (c.Username == "") != (c.Password == "") {
		return errors.New("Điền cả tên đăng nhập và mật khẩu SMTP hoặc để trống cả hai.")
	}
	return nil
}

func (a *App) readSMTP(ctx context.Context) (*smtpConfiguration, error) {
	var saved *smtpConfiguration
	err := a.store.book(ctx, platformSettingsScope, false, func(b *Book) error {
		saved = b.PlatformSMTP
		return nil
	})
	return saved, err
}

func (a *App) smtpMailer(c *smtpConfiguration) *verificationSMTP {
	if c == nil {
		return a.smtpFallback
	}
	if c.Disabled {
		return nil
	}
	address := net.JoinHostPort(c.Host, c.Port)
	trusted := a.smtpFallback != nil && strings.EqualFold(a.smtpFallback.address, address)
	return &verificationSMTP{address: address, sender: c.Sender, username: c.Username, password: c.Password,
		tls: &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}, startTLS: c.Mode == "starttls", publicOnly: !trusted}
}

func (a *App) smtpFormValues(saved *smtpConfiguration) smtpConfiguration {
	if saved != nil {
		return *saved
	}
	if fallback := a.smtpFallback; fallback != nil {
		host, port, _ := net.SplitHostPort(fallback.address)
		return smtpConfiguration{Host: host, Port: port, Sender: fallback.sender, Mode: "tls", Username: fallback.username, Password: fallback.password}
	}
	return smtpConfiguration{Port: "465", Mode: "tls"}
}

func (a *App) smtpView(ctx context.Context, admin bool) (SMTPSettings, error) {
	saved, err := a.readSMTP(ctx)
	if err != nil {
		return SMTPSettings{}, err
	}
	view := SMTPSettings{Configured: a.smtpMailer(saved) != nil, CanEdit: admin}
	if admin {
		c := a.smtpFormValues(saved)
		view.Host, view.Port, view.Sender, view.Mode, view.Username, view.Revision = c.Host, c.Port, c.Sender, c.Mode, c.Username, c.Revision
		view.Authenticated = c.Password != ""
		if c.Host != "" {
			view.Address = net.JoinHostPort(c.Host, c.Port)
		}
	}
	return view, nil
}

func (a *App) updateSMTP(w http.ResponseWriter, r *http.Request, action string) {
	ctx := r.Context()
	status := "smtp"
	if action == "smtp-test" {
		if !a.allow("smtp-test", 6, time.Minute) {
			a.problem(w, r, inputError{errors.New("Hãy đợi một phút trước khi thử kết nối lại.")})
			return
		}
		saved, err := a.readSMTP(ctx)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		mailer := a.smtpMailer(saved)
		if mailer == nil {
			a.problem(w, r, inputError{errors.New("Hãy lưu cấu hình SMTP trước khi kiểm tra.")})
			return
		}
		testCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
		defer cancel()
		client, err := mailer.connect(testCtx)
		if err == nil {
			defer client.Close()
			err = client.Noop()
		}
		if err != nil {
			a.problem(w, r, inputError{errors.New("Kết nối SMTP thất bại. Kiểm tra mạng, chứng chỉ TLS và thông tin đăng nhập.")})
			return
		}
		status = "smtp-tested"
	} else {
		err := a.store.book(ctx, platformSettingsScope, true, func(b *Book) error {
			prior := a.smtpFormValues(b.PlatformSMTP)
			if prior.Revision != r.Form.Get("revision") {
				return inputError{errors.New("Cấu hình đã thay đổi. Hãy tải lại trang trước khi lưu.")}
			}
			if action == "smtp-disable" {
				prior.Disabled, prior.Revision = true, token()
				b.PlatformSMTP = &prior
				status = "smtp-disabled"
				return nil
			}
			c := smtpConfiguration{Host: strings.ToLower(strings.TrimSpace(r.Form.Get("host"))), Port: strings.TrimSpace(r.Form.Get("port")),
				Sender: strings.TrimSpace(r.Form.Get("sender")), Mode: r.Form.Get("mode"), Username: strings.TrimSpace(r.Form.Get("username")),
				Password: r.Form.Get("password"), Revision: token()}
			if c.Password == "" && c.Username != "" && c.Username == prior.Username && c.Host == strings.ToLower(prior.Host) && c.Port == prior.Port && c.Mode == prior.Mode {
				c.Password = prior.Password
			}
			if err := c.validate(); err != nil {
				return inputError{err}
			}
			b.PlatformSMTP = &c
			return nil
		})
		if err != nil {
			a.problem(w, r, err)
			return
		}
	}
	http.Redirect(w, r, "/?tab=settings&saved="+status+"#smtp", http.StatusSeeOther)
}

var smtpNonPublic = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("64:ff9b:1::/48"), netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("2001::/32"),
}

func smtpPublicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	for _, reserved := range smtpNonPublic {
		if reserved.Contains(ip) {
			return false
		}
	}
	return true
}

func (m *verificationSMTP) destination(ctx context.Context) (string, error) {
	if !m.publicOnly {
		return m.address, nil
	}
	host, port, err := net.SplitHostPort(m.address)
	if err != nil {
		return "", err
	}
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return "", err
	}
	for _, ip := range addresses {
		if smtpPublicIP(ip) {
			return net.JoinHostPort(ip.String(), port), nil
		}
	}
	return "", errors.New("SMTP host must resolve to a public address")
}
