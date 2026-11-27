package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func validEmail(value string) bool {
	parsed, err := mail.ParseAddress(value)
	return err == nil && len(value) <= 254 && parsed.Address == value && !strings.ContainsAny(value, "\r\n")
}

func (b *Book) updateLeaseEmail(roomID, effective, email string) error {
	room := b.room(roomID)
	if room == nil {
		return errMissing
	}
	if email != "" && !validEmail(email) {
		return errors.New("Địa chỉ email chưa hợp lệ.")
	}
	for _, lease := range b.leaseHistory(*room) {
		if lease.Effective == effective && lease.Tenant != "" {
			if b.LeaseEmails == nil {
				b.LeaseEmails = make(map[string]string)
			}
			key := roomID + "/" + effective
			if b.LeaseEmails[key] == email {
				return nil
			}
			if email == "" {
				delete(b.LeaseEmails, key)
			} else {
				b.LeaseEmails[key] = email
			}
			for j := range b.Invites {
				v := &b.Invites[j]
				i := b.invoice(v.InvoiceID)
				if i != nil && i.RoomID == roomID && i.LeaseEffective == effective {
					v.Revoked = true
				}
			}
			return nil
		}
	}
	return errMissing
}

type verificationSMTP struct {
	address, sender, username, password string
	tls                                 *tls.Config
	startTLS, publicOnly                bool
}

type SMTPSettings struct {
	Configured, Authenticated, CanEdit                    bool
	Address, Host, Port, Sender, Mode, Username, Revision string
}

func configuredVerificationSMTP() (*verificationSMTP, error) {
	address, sender := os.Getenv("ZS_SMTP_ADDRESS"), os.Getenv("ZS_SMTP_FROM")
	if address == "" && sender == "" {
		return nil, nil
	}
	host, port, err := net.SplitHostPort(address)
	username, password := os.Getenv("ZS_SMTP_USERNAME"), os.Getenv("ZS_SMTP_PASSWORD")
	if err != nil || host == "" || port == "" || !validEmail(sender) || (username == "") != (password == "") {
		return nil, errors.New("SMTP verification requires a TLS host:port, sender email and matching credentials if used")
	}
	return &verificationSMTP{address: address, sender: sender, username: username, password: password, tls: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}, nil
}

func (m *verificationSMTP) send(ctx context.Context, to, code string) error {
	if !validEmail(to) {
		return errors.New("invalid mail recipient")
	}
	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()
	c, err := m.connect(ctx)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.Mail(m.sender); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	data, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(data, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", m.sender, to, mime.QEncoding.Encode("utf-8", "Mã xác minh ZentraSpace"), time.Now().Format(time.RFC1123Z))
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(data, "Ma xac minh dung mot lan: %s\r\nNhap ma vao trang bang ke tu lien ket chu tro da gui.\r\nMa het han sau 24 gio.\r\n", code); err != nil {
		return err
	}
	return data.Close()
}

func (m *verificationSMTP) connect(ctx context.Context) (*smtp.Client, error) {
	address, err := m.destination(ctx)
	if err != nil {
		return nil, err
	}
	var conn net.Conn
	if m.startTLS {
		conn, err = (&net.Dialer{}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = (&tls.Dialer{Config: m.tls}).DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return nil, err
	}
	ready := false
	defer func() {
		if !ready {
			conn.Close()
		}
	}()
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, err
		}
	}
	c, err := smtp.NewClient(conn, m.tls.ServerName)
	if err != nil {
		return nil, err
	}
	if m.startTLS {
		if err := c.StartTLS(m.tls); err != nil {
			return nil, err
		}
	}
	if m.username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.username, m.password, m.tls.ServerName)); err != nil {
			return nil, err
		}
	}
	ready = true
	return c, nil
}

func (a *App) sendInviteCode(w http.ResponseWriter, r *http.Request, org, csrf string) {
	if !a.allow("invite-email:"+org, 12, time.Minute) {
		a.problem(w, r, inputError{errors.New("Gửi mã email chưa sẵn sàng. Hãy đợi hoặc dùng cách gửi mã thủ công.")})
		return
	}
	configuration, err := a.readSMTP(r.Context())
	if err != nil {
		a.problem(w, r, err)
		return
	}
	mailer := a.smtpMailer(configuration)
	if mailer == nil {
		a.problem(w, r, inputError{errors.New("Hệ thống chưa bật SMTP. Bạn có thể gửi mã thủ công.")})
		return
	}
	now := time.Now()
	invoiceID, code := r.Form.Get("id"), accessCode()
	v := Invite{ID: token(), InvoiceID: invoiceID, CodeHash: hash(code), Expires: now.Add(24 * time.Hour), PendingDelivery: true}
	p := Page{Screen: "invite", Role: "owner", CSRF: csrf}
	var to string
	err = a.store.book(r.Context(), org, true, func(b *Book) error {
		i := b.invoice(invoiceID)
		if i == nil || !i.Issued {
			return errMissing
		}
		to = b.LeaseEmails[i.RoomID+"/"+i.LeaseEffective]
		if !validEmail(to) {
			return inputError{errors.New("Hãy đăng ký email người nhận trong hồ sơ hợp đồng trước.")}
		}
		if len(b.Invites) >= 5000 {
			return inputError{errors.New("Đã đạt giới hạn lời mời của bản thử.")}
		}
		for j := range b.Invites {
			prior := &b.Invites[j]
			if prior.InvoiceID != invoiceID || !prior.PendingDelivery || prior.Revoked {
				continue
			}
			if prior.Expires.Add(-24 * time.Hour).After(now.Add(-time.Minute)) {
				return inputError{errors.New("Mã đang được gửi. Hãy chờ một phút trước khi thử lại.")}
			}
			prior.Revoked = true
		}
		route, err := a.store.shareToken(org, v.ID)
		if err != nil {
			return err
		}
		p.ShareURL, p.Invoice = a.origin+"/p/"+route, i
		b.Invites = append(b.Invites, v)
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	deliveryErr := mailer.send(r.Context(), to, code)
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 3*time.Second)
	defer cancel()
	err = a.store.book(finishCtx, org, true, func(b *Book) error {
		current := b.invite(v.ID)
		if current == nil || current.Revoked {
			return inputError{errors.New("Lời mời đã thay đổi. Hãy tạo lại link.")}
		}
		current.PendingDelivery = false
		if deliveryErr != nil {
			current.Revoked = true
			return nil
		}
		for j := range b.Invites {
			prior := &b.Invites[j]
			if prior.InvoiceID == invoiceID && prior.ID != v.ID {
				prior.Revoked = true
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if deliveryErr != nil {
		a.problem(w, r, inputError{errors.New("Không gửi được mã qua email. Hãy thử lại hoặc tạo link với mã thủ công.")})
		return
	}
	p.CodeDelivery = to
	a.page(w, r, p)
}
