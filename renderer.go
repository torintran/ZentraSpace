package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

const maxRenderedPage = 16 * 1024 * 1024

type frontendRenderer struct {
	endpoint *url.URL
	client   *http.Client
}

func newFrontendRenderer(value string) (*frontendRenderer, error) {
	base, err := url.Parse(value)
	if err != nil || base.Scheme != "http" || base.Host == "" || (base.Path != "" && base.Path != "/") || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("ZS_FRONTEND_URL must name the private HTTP frontend service")
	}
	return &frontendRenderer{endpoint: base, client: &http.Client{
		Timeout:       12 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

type invoiceView struct {
	Invoice
	Revision                                      string
	Remaining, Credit, NetCollected, NetReduction int64
}

func invoicePresentation(i *Invoice, shared bool) *invoiceView {
	if i == nil {
		return nil
	}
	view := &invoiceView{Invoice: *i, Remaining: i.Remaining(), Credit: i.Credit(), NetCollected: i.NetCollected(), NetReduction: i.NetReduction()}
	if shared {
		view.Entries = nil
		view.LeaseTenant = ""
		view.Reading.Adjustment.Reason = ""
	} else {
		view.Revision = i.Revision()
	}
	return view
}

func invoiceList(invoices []Invoice) []invoiceView {
	result := make([]invoiceView, 0, len(invoices))
	for j := range invoices {
		result = append(result, *invoicePresentation(&invoices[j], false))
	}
	return result
}

type roomPresentation struct {
	Room
	Revision string
}
type transferReviewView struct {
	TransferReview
	Invoices  []invoiceView
	Selected  *invoiceView
	Suggested *invoiceView
}
type pageView struct {
	Page
	Invoice    *invoiceView
	Invoices   []invoiceView
	RoomDetail *roomPresentation
	BankReview *transferReviewView
}

func frontendView(p Page) pageView {
	if len(p.Invites) > 0 {
		p.Invites = append([]Invite(nil), p.Invites...)
		for j := range p.Invites {
			p.Invites[j].CodeHash = ""
		}
	}
	view := pageView{Page: p, Invoice: invoicePresentation(p.Invoice, p.Screen == "shared"), Invoices: invoiceList(p.Invoices)}
	if p.RoomDetail != nil {
		view.RoomDetail = &roomPresentation{*p.RoomDetail, p.RoomDetail.Revision()}
	}
	if p.BankReview != nil {
		review := p.BankReview
		view.BankReview = &transferReviewView{TransferReview: *review, Invoices: invoiceList(review.Invoices),
			Selected: invoicePresentation(review.Selected, false), Suggested: invoicePresentation(review.Suggested, false)}
	}
	return view
}

func (a *App) page(w http.ResponseWriter, r *http.Request, p Page) {
	payload, err := json.Marshal(struct {
		Path string   `json:"path"`
		View pageView `json:"view"`
	}{r.URL.EscapedPath(), frontendView(p)})
	if err != nil {
		http.Error(w, "Không thể chuẩn bị giao diện.", 500)
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, a.frontend.endpoint.ResolveReference(&url.URL{Path: "/__render"}).String(), bytes.NewReader(payload))
	if err != nil {
		http.Error(w, "Không thể chuẩn bị giao diện.", 500)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.frontend.client.Do(req)
	if err != nil {
		log.Printf("frontend render failed (%T)", err)
		http.Error(w, "Giao diện tạm thời chưa sẵn sàng.", 502)
		return
	}
	defer resp.Body.Close()
	policy := resp.Header.Get("Content-Security-Policy")
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") || !strings.Contains(policy, "'nonce-") {
		log.Printf("frontend rejected render: HTTP %d", resp.StatusCode)
		http.Error(w, "Giao diện tạm thời chưa sẵn sàng.", 502)
		return
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRenderedPage+1))
	if err != nil || len(body) > maxRenderedPage {
		http.Error(w, "Không thể tải giao diện.", 502)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", policy)
	if p.Status != 0 {
		w.WriteHeader(p.Status)
	}
	w.Write(body)
}

func (f *frontendRenderer) assets() http.Handler {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(f.endpoint)
			// Static assets need range and cache negotiation, not browser credentials or framework headers.
			req.Out.Header = make(http.Header)
			for _, name := range []string{"Accept-Encoding", "Range", "If-Range", "If-None-Match", "If-Modified-Since"} {
				for _, value := range req.In.Header.Values(name) {
					req.Out.Header.Add(name, value)
				}
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusNotModified {
				resp.Header.Set("Cache-Control", "no-store")
			}
			return nil
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clean := path.Clean(r.URL.Path); clean != r.URL.Path || !strings.HasPrefix(clean, "/_next/static/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Del("Cache-Control")
		proxy.ServeHTTP(w, r)
	})
}
