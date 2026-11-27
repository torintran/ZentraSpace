package app

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestVerificationEmailUsesTLSAndSendsOnlyCode(t *testing.T) {
	certServer := httptest.NewTLSServer(nil)
	defer certServer.Close()
	listener, err := tls.Listen("tcp4", "127.0.0.1:0", certServer.TLS)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	mailbox := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		fmt.Fprint(conn, "220 test SMTP\r\n")
		lines := bufio.NewScanner(conn)
		for lines.Scan() {
			line := lines.Text()
			switch {
			case strings.HasPrefix(line, "EHLO "), strings.HasPrefix(line, "MAIL FROM:<no-reply@example.test>"), line == "RCPT TO:<tenant@example.test>":
				fmt.Fprint(conn, "250 OK\r\n")
			case line == "DATA":
				fmt.Fprint(conn, "354 Start mail\r\n")
				var msg strings.Builder
				for lines.Scan() && lines.Text() != "." {
					fmt.Fprintln(&msg, lines.Text())
				}
				mailbox <- msg.String()
				fmt.Fprint(conn, "250 Accepted\r\n")
				return
			default:
				fmt.Fprint(conn, "550 Invalid envelope\r\n")
				return
			}
		}
	}()
	roots := x509.NewCertPool()
	roots.AddCert(certServer.Certificate())
	m := verificationSMTP{address: listener.Addr().String(), sender: "no-reply@example.test", tls: &tls.Config{RootCAs: roots, ServerName: "127.0.0.1", MinVersion: tls.VersionTLS12}}
	if err := m.send(context.Background(), "tenant@example.test", "123456"); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-mailbox:
		if !strings.Contains(msg, "To: tenant@example.test\n") || !strings.Contains(msg, "123456") || strings.Contains(msg, "https://") {
			t.Fatalf("unexpected message: %q", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("mailbox did not receive code")
	}
}

func TestEmailChangeRevokesPriorInvitationAndPendingCodeCannotVerify(t *testing.T) {
	now := time.Now()
	b := Book{Rooms: []Room{{ID: "r", Tenant: "Khách", Start: "2026-11"}}, Invoices: []Invoice{{ID: "i", RoomID: "r", LeaseEffective: "2026-11"}},
		Invites: []Invite{{InvoiceID: "i", PendingDelivery: true, CodeHash: hash("123456"), Expires: now.Add(time.Hour)}}}
	if b.Invites[0].verify("123456", now) {
		t.Fatal("pending code accepted")
	}
	if err := b.updateLeaseEmail("r", "2026-11", "tenant@example.test"); err != nil {
		t.Fatal(err)
	}
	if !b.Invites[0].Revoked || b.LeaseEmails["r/2026-11"] != "tenant@example.test" {
		t.Fatal("email change did not revoke the prior invitation")
	}
}
