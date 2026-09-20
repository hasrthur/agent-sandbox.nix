package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// readRequest parses a raw HTTP/1.1 request the way the proxy does, so tests
// exercise the same framing the wire produces rather than a hand-built struct.
func readRequest(t *testing.T, raw string) *http.Request {
	t.Helper()
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	return req
}

var (
	getOnlyPolicy  = Config{"example.com": {Methods: map[string]bool{"GET": true, "HEAD": true}}}
	wildcardPolicy = Config{"example.com": {AllowAll: true}}
)

func TestIsBlockedAddr(t *testing.T) {
	cases := []struct {
		addr    string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.1.2.3", true},
		{"::1", true},
		{"::ffff:127.0.0.1", true},
		{"0.0.0.0", true},
		{"::", true},
		{"::ffff:0.0.0.0", true},
		{"169.254.169.254", true},
		{"::ffff:169.254.169.254", true},
		{"fe80::1", true},
		// Private ranges stay dialable: allowlisting an internal company
		// server is legitimate, and allowedHostPorts cannot express it.
		{"10.0.0.5", false},
		{"172.16.0.1", false},
		{"192.168.1.1", false},
		{"fc00::1", false},
		{"93.184.216.34", false},
		{"2606:4700:4700::1111", false},
	}
	for _, c := range cases {
		ip := net.ParseIP(c.addr)
		if ip == nil {
			t.Fatalf("test bug: %q is not an IP", c.addr)
		}
		if got := isBlockedAddr(ip); got != c.blocked {
			t.Errorf("isBlockedAddr(%s) = %v, want %v", c.addr, got, c.blocked)
		}
	}
}

// A redirect skips resolveVetted, so an entry the caller never wrote is
// worth as much as one it did. "a=b=c" is what reaches the proxy when a key
// of "a=b" is written: taking the first "=" would silently redirect "a".
func TestParseRedirectEnvRejectsExtraEquals(t *testing.T) {
	if _, err := parseRedirectEnv("a=b=c"); err == nil {
		t.Error("parseRedirectEnv(\"a=b=c\") succeeded, want an error")
	}
}

func TestParseRedirectEnv(t *testing.T) {
	got, err := parseRedirectEnv("Example.com=127.0.0.1:8080, other.test=[::1]:9090")
	if err != nil {
		t.Fatalf("parseRedirectEnv errored: %v", err)
	}
	want := Redirects{"example.com": "127.0.0.1:8080", "other.test": "[::1]:9090"}
	if len(got) != len(want) {
		t.Fatalf("parseRedirectEnv = %v, want %v", got, want)
	}
	for host, addr := range want {
		if got[host] != addr {
			t.Errorf("parseRedirectEnv[%q] = %q, want %q", host, got[host], addr)
		}
	}
}

// An allowlisted name whose address is loopback must be refused: the proxy
// runs on the host, so dialing it would reach the host services that
// allowedHostPorts exists to gate.
func TestResolveVettedRefusesLoopback(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "localhost", "169.254.169.254"} {
		addr, err := resolveVetted(host, "443")
		if !errors.Is(err, errBlockedAddress) {
			t.Errorf("resolveVetted(%q) = (%q, %v), want errBlockedAddress", host, addr, err)
		}
	}
}

func TestResolveVettedAllowsPublicAndPrivate(t *testing.T) {
	cases := []struct {
		host string
		want string
	}{
		{host: "93.184.216.34", want: "93.184.216.34:443"},
		{host: "10.0.0.5", want: "10.0.0.5:443"},
		{host: "192.168.1.1", want: "192.168.1.1:443"},
		{host: "2606:4700:4700::1111", want: "[2606:4700:4700::1111]:443"},
	}
	for _, c := range cases {
		got, err := resolveVetted(c.host, "443")
		if err != nil {
			t.Errorf("resolveVetted(%q) errored: %v", c.host, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveVetted(%q) = %q, want %q", c.host, got, c.want)
		}
	}
}

func TestDialFailureStatus(t *testing.T) {
	if got := dialFailureStatus(errBlockedAddress); got != http.StatusForbidden {
		t.Errorf("blocked address status = %d, want 403", got)
	}
	if got := dialFailureStatus(errors.New("connection refused")); got != http.StatusBadGateway {
		t.Errorf("dial failure status = %d, want 502", got)
	}
}

func TestApplyFilters(t *testing.T) {
	longQuery := strings.Repeat("x", maxURLBytes+1)

	cases := []struct {
		name   string
		cfg    Config
		host   string
		raw    string
		status int
	}{
		{
			name:   "plain GET allowed",
			cfg:    getOnlyPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\n\r\n",
			status: 0,
		},
		{
			name:   "GET with an explicitly empty body allowed",
			cfg:    getOnlyPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 0\r\n\r\n",
			status: 0,
		},
		{
			name:   "GET carrying a Content-Length body refused",
			cfg:    getOnlyPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: http.StatusForbidden,
		},
		{
			name:   "GET carrying a chunked body refused",
			cfg:    getOnlyPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n0\r\n\r\n",
			status: http.StatusForbidden,
		},
		{
			name:   "HEAD carrying a body refused",
			cfg:    getOnlyPolicy,
			raw:    "HEAD /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: http.StatusForbidden,
		},
		{
			// curl -X get: the policy check uppercases, so the method passes.
			// The body check must see the same normalised value or the
			// GET-only policy stops being read-only.
			name:   "lowercase get carrying a body refused",
			cfg:    getOnlyPolicy,
			raw:    "get /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: http.StatusForbidden,
		},
		{
			name:   "lowercase get subject to the URL cap",
			cfg:    getOnlyPolicy,
			raw:    "get /thing?q=" + longQuery + " HTTP/1.1\r\nHost: example.com\r\n\r\n",
			status: http.StatusRequestURITooLong,
		},
		{
			name:   "URL cap applies to POST",
			cfg:    wildcardPolicy,
			raw:    "POST /thing?q=" + longQuery + " HTTP/1.1\r\nHost: example.com\r\nContent-Length: 0\r\n\r\n",
			status: http.StatusRequestURITooLong,
		},
		{
			name:   "POST with a body allowed under a wildcard policy",
			cfg:    wildcardPolicy,
			raw:    "POST /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: 0,
		},
		{
			name:   "POST refused under a GET-only policy",
			cfg:    getOnlyPolicy,
			raw:    "POST /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: http.StatusForbidden,
		},
		{
			name:   "lowercase post refused under a GET-only policy",
			cfg:    getOnlyPolicy,
			raw:    "post /thing HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello",
			status: http.StatusForbidden,
		},
		{
			name:   "WebSocket upgrade refused under a method policy",
			cfg:    getOnlyPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n",
			status: http.StatusForbidden,
		},
		{
			name:   "WebSocket upgrade allowed under a wildcard policy",
			cfg:    wildcardPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n",
			status: 0,
		},
		{
			// The wildcard makes every name allowlisted, so only the
			// comparison against the CONNECT host can refuse this.
			name:   "Host naming another host refused",
			cfg:    Config{"*": {AllowAll: true}},
			raw:    "GET /thing HTTP/1.1\r\nHost: other.example\r\n\r\n",
			status: http.StatusForbidden,
		},
		{
			name:   "Host bracketing the CONNECT host refused",
			cfg:    Config{"*": {AllowAll: true}},
			raw:    "GET /thing HTTP/1.1\r\nHost: [example.com]\r\n\r\n",
			status: http.StatusForbidden,
		},
		{
			name:   "Host carrying the port allowed",
			cfg:    wildcardPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com:443\r\n\r\n",
			status: 0,
		},
		{
			name:   "Host differing in case allowed",
			cfg:    wildcardPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: EXAMPLE.CoM\r\n\r\n",
			status: 0,
		},
		{
			name:   "Host carrying a trailing dot allowed",
			cfg:    wildcardPolicy,
			raw:    "GET /thing HTTP/1.1\r\nHost: example.com.\r\n\r\n",
			status: 0,
		},
		{
			name:   "IPv6 CONNECT host matched through its brackets",
			cfg:    Config{"2606:4700:4700::1111": {AllowAll: true}},
			host:   "2606:4700:4700::1111",
			raw:    "GET /thing HTTP/1.1\r\nHost: [2606:4700:4700::1111]:443\r\n\r\n",
			status: 0,
		},
		{
			name:   "request without a Host refused",
			cfg:    Config{"*": {AllowAll: true}},
			raw:    "GET /thing HTTP/1.0\r\n\r\n",
			status: http.StatusBadRequest,
		},
		{
			// handle refuses a non-ASCII authority, so the CONNECT host is
			// ASCII and a Host that only folds onto it under Unicode is a
			// different name: it reaches the origin punycoded as a third.
			name:   "Host folding onto the CONNECT host under Unicode refused",
			cfg:    Config{"*": {AllowAll: true}},
			host:   "k.example",
			raw:    "GET /thing HTTP/1.1\r\nHost: \u212A.example\r\n\r\n",
			status: http.StatusForbidden,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			host := c.host
			if host == "" {
				host = "example.com"
			}
			req := readRequest(t, c.raw)
			status, reason := applyFilters(req, host, c.cfg)
			if status != c.status {
				t.Errorf("applyFilters = (%d, %q), want status %d", status, reason, c.status)
			}
		})
	}
}

// The Host header selects the origin on an upstream that serves several names
// from one address, while the allowlist was applied to the CONNECT host.
func TestApplyFiltersReportsHostMismatch(t *testing.T) {
	req := readRequest(t, "GET /thing HTTP/1.1\r\nHost: other.example\r\n\r\n")
	status, reason := applyFilters(req, "example.com", Config{"*": {AllowAll: true}})
	if status != http.StatusForbidden {
		t.Fatalf("applyFilters = (%d, %q), want 403", status, reason)
	}
	if !strings.Contains(reason, "CONNECT host") {
		t.Errorf("reason = %q, want it to name the CONNECT host", reason)
	}
}

func TestHostOnly(t *testing.T) {
	cases := []struct {
		addr string
		want string
		ok   bool
	}{
		{addr: "example.com:443", want: "example.com", ok: true},
		{addr: "example.com", want: "example.com", ok: true},
		{addr: "[::1]:443", want: "::1", ok: true},
		{addr: "[2606:4700:4700::1111]", want: "2606:4700:4700::1111", ok: true},
		// Brackets wrap an IPv6 literal and nothing else. SplitHostPort strips
		// them without reading what is inside, so a name written this way
		// would otherwise pass as the bare name it contains.
		{addr: "[example.com]:443", ok: false},
		{addr: "[example.com]", ok: false},
		{addr: "[127.0.0.1]:443", ok: false},
		{addr: "[::1", ok: false},
		{addr: "[::1]x", ok: false},
	}
	for _, c := range cases {
		got, ok := hostOnly(c.addr)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("hostOnly(%q) = (%q, %v), want (%q, %v)", c.addr, got, ok, c.want, c.ok)
		}
	}
}

func TestLowerASCII(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "EXAMPLE.CoM", want: "example.com"},
		// U+212A KELVIN SIGN and U+0130 lowercase to "k" and "i" under a
		// Unicode fold. They are their own names to the resolver and to the
		// origin, and must stay their own names here.
		{in: "\u212A.example", want: "\u212A.example"},
		{in: "\u0130.example", want: "\u0130.example"},
	}
	for _, c := range cases {
		if got := lowerASCII(c.in); got != c.want {
			t.Errorf("lowerASCII(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHasNonASCII(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{in: "example.com:443", want: false},
		{in: "[2606:4700:4700::1111]:443", want: false},
		{in: "\u212A.example:443", want: true},
		{in: "sub.\u0130.example", want: true},
	}
	for _, c := range cases {
		if got := hasNonASCII(c.in); got != c.want {
			t.Errorf("hasNonASCII(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLookupPolicyFoldsOnlyASCII(t *testing.T) {
	cfg := Config{"k.example": {AllowAll: true}, "i.example": {AllowAll: true}}
	for _, host := range []string{"\u212A.example", "\u0130.example", "sub.\u212A.example"} {
		if isDomainAllowed(host, cfg) {
			t.Errorf("isDomainAllowed(%q) = true, want false: it is not the allowlisted name", host)
		}
	}
	for _, host := range []string{"K.EXAMPLE", "sub.K.example"} {
		if !isDomainAllowed(host, cfg) {
			t.Errorf("isDomainAllowed(%q) = false, want true: ASCII case must still fold", host)
		}
	}
}

// An allowlist key folded to ASCII would admit a host the operator never
// wrote: a key spelled with U+0130 would be stored, and allowed, as
// "i.example".
func TestLoadConfigKeepsNonASCIIDomainUnmatched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "allowlist.json")
	if err := os.WriteFile(path, []byte(`{"\u0130.example": "*"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stderr := captureStderr(t)
	cfg, err := loadConfig(path)
	log := stderr()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if isDomainAllowed("i.example", cfg) {
		t.Error("i.example allowed: the key folded onto an ASCII name the operator never wrote")
	}
	if !strings.Contains(log, "punycode") {
		t.Errorf("stderr = %q, want a warning telling the operator to write punycode", log)
	}
}

func TestLoadConfigWarnsOnCatchAll(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		want  string
		quiet bool
	}{
		{name: "all methods", raw: `{"*": "*"}`, want: "all methods"},
		{name: "method list", raw: `{"*": ["post","GET"]}`, want: "GET, POST"},
		{name: "no catch-all", raw: `{"example.com": "*"}`, quiet: true},
		{name: "empty method list", raw: `{"*": []}`, quiet: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "allowlist.json")
			if err := os.WriteFile(path, []byte(c.raw), 0o600); err != nil {
				t.Fatalf("write config: %v", err)
			}

			stderr := captureStderr(t)
			_, err := loadConfig(path)
			log := stderr()
			if err != nil {
				t.Fatalf("loadConfig: %v", err)
			}
			if c.quiet {
				if strings.Contains(log, "WARNING") {
					t.Errorf("stderr = %q, want no warning: nothing here reaches an unlisted domain", log)
				}
				return
			}
			if !strings.Contains(log, c.want) {
				t.Errorf("stderr = %q, want it to name the methods %q permits", log, c.want)
			}
			if !strings.Contains(log, "every domain not listed") {
				t.Errorf("stderr = %q, want it to name what the entry reaches", log)
			}
		})
	}
}

func TestHasRequestBody(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"no body", "GET /x HTTP/1.1\r\nHost: example.com\r\n\r\n", false},
		{"zero length", "GET /x HTTP/1.1\r\nHost: example.com\r\nContent-Length: 0\r\n\r\n", false},
		{"content length", "GET /x HTTP/1.1\r\nHost: example.com\r\nContent-Length: 5\r\n\r\nhello", true},
		{"chunked", "GET /x HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n0\r\n\r\n", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasRequestBody(readRequest(t, c.raw)); got != c.want {
				t.Errorf("hasRequestBody = %v, want %v", got, c.want)
			}
		})
	}
}

func TestWebSocketTunnel(t *testing.T) {
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen upstream: %v", err)
	}
	defer upstream.Close()
	go func() {
		conn, err := upstream.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		br := bufio.NewReader(conn)
		req, err := http.ReadRequest(br)
		if err != nil {
			return
		}
		if !isWebSocketUpgrade(req) {
			fmt.Fprintf(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
			return
		}
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		io.Copy(conn, br)
	}()

	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}
	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen proxy: %v", err)
	}
	defer proxyLn.Close()
	redirects := Redirects{"example.com": upstream.Addr().String()}
	go func() {
		conn, err := proxyLn.Accept()
		if err != nil {
			return
		}
		handle(conn, wildcardPolicy, ca, redirects)
	}()

	conn, err := net.Dial("tcp", proxyLn.Addr().String())
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil || !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("CONNECT response = %q, err = %v", status, err)
	}
	if _, err := br.ReadString('\n'); err != nil {
		t.Fatalf("read CONNECT header terminator: %v", err)
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "example.com", InsecureSkipVerify: true})
	if err := tlsConn.Handshake(); err != nil {
		t.Fatalf("TLS handshake: %v", err)
	}

	fmt.Fprintf(tlsConn, "GET /ws HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
	tlsBr := bufio.NewReader(tlsConn)
	status, err = tlsBr.ReadString('\n')
	if err != nil || !strings.HasPrefix(status, "HTTP/1.1 101") {
		t.Fatalf("upgrade response = %q, err = %v", status, err)
	}
	for {
		line, err := tlsBr.ReadString('\n')
		if err != nil {
			t.Fatalf("read upgrade headers: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}

	if _, err := tlsConn.Write([]byte("ping")); err != nil {
		t.Fatalf("write through tunnel: %v", err)
	}
	echo := make([]byte, 4)
	if _, err := io.ReadFull(tlsBr, echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(echo) != "ping" {
		t.Errorf("echo = %q, want %q", echo, "ping")
	}
}

func TestFirstContact(t *testing.T) {
	host := "first-contact.test"
	if !firstContact(host) {
		t.Errorf("firstContact(%q) = false on first call, want true", host)
	}
	if firstContact(host) {
		t.Errorf("firstContact(%q) = true on repeat call, want false", host)
	}
	if !firstContact("other." + host) {
		t.Errorf("firstContact of an unseen host = false, want true")
	}
}

type countingWriter struct {
	writes int
	buf    bytes.Buffer
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.writes++
	return c.buf.Write(p)
}

func TestWriteSwitchingProtocolsSingleWrite(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusSwitchingProtocols,
		Header: http.Header{
			"Upgrade":                {"websocket"},
			"Connection":             {"Upgrade"},
			"Sec-Websocket-Accept":   {"s3pPLMBiTxaQ9kYGzzhZRbK+xOo="},
			"Sec-Websocket-Protocol": {"chat"},
			"Date":                   {"Fri, 05 Sep 2026 12:46:33 GMT"},
			"Server":                 {"cloudflare"},
			"Cf-Ray":                 {"deadbeefcafe-LHR"},
		},
	}
	var w countingWriter
	if err := writeSwitchingProtocols(&w, resp); err != nil {
		t.Fatalf("writeSwitchingProtocols: %v", err)
	}
	if w.writes != 1 {
		t.Errorf("writes = %d, want 1: a TLS conn emits a record per write, and clients reject a drip-fed handshake", w.writes)
	}
	got := w.buf.String()
	if !strings.HasPrefix(got, "HTTP/1.1 101 Switching Protocols\r\n") {
		t.Errorf("status line = %q", got)
	}
	if !strings.HasSuffix(got, "\r\n\r\n") {
		t.Errorf("header block not terminated: %q", got)
	}
	for _, want := range []string{"Upgrade: websocket", "Sec-Websocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo="} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}

func TestHeaderLimitReader(t *testing.T) {
	src := strings.Repeat("a", maxHeaderBytes*2)

	h := &headerLimitReader{r: strings.NewReader(src)}
	h.arm()
	n, err := io.Copy(io.Discard, h)
	if !errors.Is(err, errHeaderTooLarge) {
		t.Errorf("armed read error = %v, want errHeaderTooLarge", err)
	}
	if n != maxHeaderBytes {
		t.Errorf("armed read passed %d bytes, want %d", n, maxHeaderBytes)
	}

	h = &headerLimitReader{r: strings.NewReader(src)}
	h.arm()
	h.disarm()
	n, err = io.Copy(io.Discard, h)
	if err != nil {
		t.Errorf("disarmed read error = %v, want nil", err)
	}
	if n != int64(len(src)) {
		t.Errorf("disarmed read passed %d bytes, want %d", n, len(src))
	}
}

func TestHandleRejectsOversizedHeader(t *testing.T) {
	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}
	client, server := net.Pipe()
	defer client.Close()
	go handle(server, wildcardPolicy, ca, Redirects{})

	client.SetDeadline(time.Now().Add(10 * time.Second))
	go func() {
		fmt.Fprintf(client, "GET / HTTP/1.1\r\nHost: example.com\r\nX-Big: %s\r\n\r\n", strings.Repeat("a", maxHeaderBytes*2))
	}()

	// The header never completes, so the proxy closes without answering
	// rather than forwarding it upstream.
	got, err := io.ReadAll(client)
	if err != nil && !errors.Is(err, io.ErrClosedPipe) && !errors.Is(err, io.EOF) {
		t.Fatalf("read after oversized header: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("proxy answered %q, want the connection closed with nothing written", got)
	}
}

// The allowlist entry here is the ASCII name the requested authority folds
// onto under Unicode, so nothing after this point would refuse it: CONNECT
// answers 200 and runs the client handshake before any per-request check.
func TestHandleRefusesNonASCIIHost(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "CONNECT", raw: "CONNECT \u212A.example:443 HTTP/1.1\r\nHost: \u212A.example:443\r\n\r\n"},
		{name: "plaintext", raw: "GET /thing HTTP/1.1\r\nHost: \u212A.example\r\n\r\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ca, err := newCertAuthority()
			if err != nil {
				t.Fatalf("new cert authority: %v", err)
			}
			client, server := net.Pipe()
			defer client.Close()

			stderr := captureStderr(t)
			done := make(chan struct{})
			go func() {
				defer close(done)
				handle(server, Config{"k.example": {AllowAll: true}}, ca, Redirects{})
			}()

			client.SetDeadline(time.Now().Add(10 * time.Second))
			go fmt.Fprint(client, c.raw)
			status, err := bufio.NewReader(client).ReadString('\n')
			<-done
			log := stderr()
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if !strings.HasPrefix(status, "HTTP/1.1 403") {
				t.Errorf("response = %q, want 403: the authority is not the allowlisted name", status)
			}
			if !strings.Contains(log, "non-ASCII host") {
				t.Errorf("log = %q, want it to name the refusal", log)
			}
		})
	}
}

func TestHandleMITMDeadlinesClientHandshake(t *testing.T) {
	old := clientHandshakeTimeout
	clientHandshakeTimeout = 50 * time.Millisecond
	defer func() { clientHandshakeTimeout = old }()

	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}
	client, server := net.Pipe()
	defer client.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handle(server, wildcardPolicy, ca, Redirects{})
	}()

	client.SetDeadline(time.Now().Add(10 * time.Second))
	go fmt.Fprintf(client, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")

	// The 200 is written before the handshake, so a client that stops here
	// holds a connection slot until the deadline gives it back.
	br := bufio.NewReader(client)
	status, err := br.ReadString('\n')
	if err != nil || !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("CONNECT response = %q, err = %v", status, err)
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("handle did not return: a client that never sends a ClientHello parks the goroutine")
	}
}

// fakeListener drives serve from a closure, so a test can script accept
// failures the network will not produce on demand.
type fakeListener struct {
	accept func() (net.Conn, error)
}

func (f *fakeListener) Accept() (net.Conn, error) { return f.accept() }
func (f *fakeListener) Close() error              { return nil }
func (f *fakeListener) Addr() net.Addr            { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }

func captureStderr(t *testing.T) func() string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()
	return func() string {
		os.Stderr = old
		w.Close()
		<-done
		r.Close()
		return buf.String()
	}
}

func TestServeLogsAcceptBackoffOnce(t *testing.T) {
	stderr := captureStderr(t)
	stop := make(chan struct{})
	var calls atomic.Int32
	reached := make(chan struct{})
	var once sync.Once

	ln := &fakeListener{accept: func() (net.Conn, error) {
		if calls.Add(1) <= 3 {
			return nil, errors.New("accept tcp: too many open files")
		}
		once.Do(func() { close(reached) })
		<-stop
		return nil, errors.New("listener stopped")
	}}
	go serve(ln, wildcardPolicy, nil, Redirects{})

	select {
	case <-reached:
	case <-time.After(10 * time.Second):
		close(stop)
		stderr()
		t.Fatal("serve did not retry past the scripted accept failures")
	}
	close(stop)
	got := stderr()

	// One line for the storm, not one per failure: the point of the log is to
	// show up beside the traffic it explains, not to bury it.
	if n := strings.Count(got, "accept error, backing off"); n != 1 {
		t.Errorf("backoff logged %d times, want 1:\n%s", n, got)
	}
	if !strings.Contains(got, "too many open files") {
		t.Errorf("log does not carry the underlying error:\n%s", got)
	}
}

func TestServeRefusesBeyondMaxConns(t *testing.T) {
	stderr := captureStderr(t)
	defer func() {
		if t.Failed() {
			t.Log(stderr())
		}
	}()

	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}

	// Built up front rather than inside the closure, which runs on serve's
	// goroutine while the cleanup below reads the slice from this one.
	held := make([]net.Conn, maxConns)
	ready := make([]net.Conn, maxConns)
	for i := range held {
		held[i], ready[i] = net.Pipe()
	}
	defer func() {
		for _, c := range held {
			c.Close()
		}
	}()
	refusedClient, refusedServer := net.Pipe()
	defer refusedClient.Close()

	stop := make(chan struct{})
	defer close(stop)
	handed := 0
	// Accept is serial and the slot is claimed before the next call, so by the
	// time the extra connection is handed over every slot is genuinely taken.
	ln := &fakeListener{accept: func() (net.Conn, error) {
		switch {
		case handed < maxConns:
			handed++
			return ready[handed-1], nil
		case handed == maxConns:
			handed++
			return refusedServer, nil
		default:
			<-stop
			return nil, errors.New("listener stopped")
		}
	}}
	go serve(ln, wildcardPolicy, ca, Redirects{})

	refusedClient.SetDeadline(time.Now().Add(10 * time.Second))
	status, err := bufio.NewReader(refusedClient).ReadString('\n')
	if err != nil {
		t.Fatalf("read refusal: %v", err)
	}
	if !strings.HasPrefix(status, "HTTP/1.1 503") {
		t.Errorf("connection %d got %q, want 503 once the cap is reached", maxConns+1, status)
	}
}

const (
	// Equal byte lengths, as the caller minting the phantom is required to
	// guarantee: base64 output length is a function of input length, so the
	// encoded forms below match in length only because these do.
	testPhantom = "PHANTOM-0123456789abcdef"
	testReal    = "ghp_realrealrealrealreal"
)

// encodedPair is the payload the header form carries: the base64 of a fixed
// username joined to the token, in which the token appears nowhere literally.
// The caller registers this form, the proxy does no encoding of its own.
func encodedPair(token string) string {
	return base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))
}

// basicAuth builds the header git actually sends.
func basicAuth(token string) string {
	return "Basic " + encodedPair(token)
}

func credentialFor(phantom, real string, hosts ...string) Credentials {
	set := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		set[h] = true
	}
	return Credentials{{Phantom: phantom, Real: real, Hosts: set}}
}

func testCredentials(hosts ...string) Credentials {
	return credentialFor(testPhantom, testReal, hosts...)
}

func encodedCredentials(hosts ...string) Credentials {
	return credentialFor(encodedPair(testPhantom), encodedPair(testReal), hosts...)
}

func TestParseCredentialEnv(t *testing.T) {
	cases := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{
			name: "unset declares nothing",
			env:  "",
		},
		{
			name: "a credential with its hosts",
			env:  `[{"phantom":"p","real":"r","hosts":["github.com"]}]`,
		},
		{
			// Substitutable nowhere rather than everywhere, and so not an error.
			name: "an absent host list",
			env:  `[{"phantom":"p","real":"r"}]`,
		},
		{
			name: "an empty host list",
			env:  `[{"phantom":"p","real":"r","hosts":[]}]`,
		},
		{
			name:    "malformed JSON refuses to start",
			env:     `[{"phantom":`,
			wantErr: true,
		},
		{
			name:    "an object rather than an array refuses to start",
			env:     `{"phantom":"p","real":"r"}`,
			wantErr: true,
		},
		{
			name:    "an empty phantom refuses to start",
			env:     `[{"phantom":"","real":"r","hosts":["github.com"]}]`,
			wantErr: true,
		},
		{
			name:    "an empty real value refuses to start",
			env:     `[{"phantom":"p","real":"","hosts":["github.com"]}]`,
			wantErr: true,
		},
		{
			name:    "an empty host refuses to start",
			env:     `[{"phantom":"p","real":"r","hosts":[""]}]`,
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseCredentialEnv(c.env)
			if (err != nil) != c.wantErr {
				t.Errorf("parseCredentialEnv errored = %v, want %v", err, c.wantErr)
			}
		})
	}
}

func TestParseCredentialEnvNormalisesHosts(t *testing.T) {
	got, err := parseCredentialEnv(`[{"phantom":"p","real":"r","hosts":["GitHub.com.", " api.github.com "]}]`)
	if err != nil {
		t.Fatalf("parseCredentialEnv errored: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("parseCredentialEnv returned %d credentials, want 1", len(got))
	}
	for _, host := range []string{"github.com", "api.github.com"} {
		if !got[0].Hosts[host] {
			t.Errorf("host %q not declared, got %v", host, got[0].Hosts)
		}
	}
}

// The proxy's stderr is its log, so an error naming the entry it rejected —
// which is what parseRedirectEnv does — would write the credential to it.
func TestParseCredentialEnvErrorOmitsTheValue(t *testing.T) {
	cases := []struct {
		name string
		env  string
	}{
		{name: "rejected entry", env: `[{"phantom":"","real":"` + testReal + `","hosts":["github.com"]}]`},
		{name: "malformed JSON", env: `[{"real":"` + testReal + `"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parseCredentialEnv(c.env)
			if err == nil {
				t.Fatal("parseCredentialEnv succeeded, want an error")
			}
			if strings.Contains(err.Error(), testReal) {
				t.Errorf("error names the credential: %q", err)
			}
		})
	}
}

func TestSubstituteCredentials(t *testing.T) {
	cases := []struct {
		name   string
		creds  Credentials
		host   string
		header http.Header
		want   http.Header
	}{
		{
			name:   "a declared host receives the real value",
			creds:  testCredentials("example.com"),
			header: http.Header{"Authorization": {"Bearer " + testPhantom}},
			want:   http.Header{"Authorization": {"Bearer " + testReal}},
		},
		{
			// Registered in the form that travels, and matched inside the
			// larger header value that carries it.
			name:   "the encoded header form is substituted",
			creds:  encodedCredentials("example.com"),
			header: http.Header{"Authorization": {basicAuth(testPhantom)}},
			want:   http.Header{"Authorization": {basicAuth(testReal)}},
		},
		{
			// Base64 hides the token from a scan for its bare form, so the
			// bare registration reaches it by decoding rather than by having
			// the encoded pair registered alongside.
			name:   "a bare registration covers the encoded form",
			creds:  testCredentials("example.com"),
			header: http.Header{"Authorization": {basicAuth(testPhantom)}},
			want:   http.Header{"Authorization": {basicAuth(testReal)}},
		},
		{
			name:   "an undeclared host keeps the phantom",
			creds:  encodedCredentials("github.com"),
			header: http.Header{"Authorization": {basicAuth(testPhantom)}},
			want:   http.Header{"Authorization": {basicAuth(testPhantom)}},
		},
		{
			name:   "an empty host list is substitutable nowhere",
			creds:  encodedCredentials(),
			header: http.Header{"Authorization": {basicAuth(testPhantom)}},
			want:   http.Header{"Authorization": {basicAuth(testPhantom)}},
		},
		{
			// The tunnel was established to example.com and the filter was
			// applied to that name. Keying on the Host header inside would let
			// a client name a declared host there and collect its credential.
			name:   "a credential declared for another host is not applied to this tunnel",
			creds:  encodedCredentials("other.example"),
			header: http.Header{"Authorization": {basicAuth(testPhantom)}, "Host": {"other.example"}},
			want:   http.Header{"Authorization": {basicAuth(testPhantom)}, "Host": {"other.example"}},
		},
		{
			// Nothing decides between the two forms: the plain scan runs over
			// every value, and the decode runs after it on the values that
			// declare themselves as Basic. A request carrying both is covered
			// in one pass.
			name:  "a plain and an encoded credential in the same request",
			creds: testCredentials("example.com"),
			header: http.Header{
				"Authorization": {basicAuth(testPhantom)},
				"X-Api-Key":     {testPhantom},
			},
			want: http.Header{
				"Authorization": {basicAuth(testReal)},
				"X-Api-Key":     {testReal},
			},
		},
		{
			// The plain pass rewrites the value the decode then reads, so a
			// Basic value whose payload is not base64 of anything must still
			// come back exactly as it arrived.
			name:   "a Basic value whose payload is the plain phantom",
			creds:  testCredentials("example.com"),
			header: http.Header{"Authorization": {"Basic " + testPhantom}},
			want:   http.Header{"Authorization": {"Basic " + testReal}},
		},
		{
			name:   "every header value is scanned, not a fixed list of names",
			creds:  testCredentials("example.com"),
			header: http.Header{"X-Whatever": {"token=" + testPhantom + "; scope=repo"}},
			want:   http.Header{"X-Whatever": {"token=" + testReal + "; scope=repo"}},
		},
		{
			name:   "every value of a repeated header is scanned",
			creds:  testCredentials("example.com"),
			header: http.Header{"X-Whatever": {"first " + testPhantom, "second " + testPhantom}},
			want:   http.Header{"X-Whatever": {"first " + testReal, "second " + testReal}},
		},
		{
			// Nothing is ever added. A request that carries no phantom cannot
			// gain a credential, which is what makes a missed match inert
			// rather than a leak.
			name:   "a request carrying no phantom receives no credential",
			creds:  encodedCredentials("example.com"),
			header: http.Header{"Authorization": {basicAuth("some-other-token")}},
			want:   http.Header{"Authorization": {basicAuth("some-other-token")}},
		},
		{
			name:   "the host matches through case and a trailing dot",
			creds:  testCredentials("example.com"),
			host:   "EXAMPLE.com.",
			header: http.Header{"Authorization": {"Bearer " + testPhantom}},
			want:   http.Header{"Authorization": {"Bearer " + testReal}},
		},
		{
			name:   "no credentials declared leaves every header alone",
			creds:  nil,
			header: http.Header{"Authorization": {basicAuth(testPhantom)}},
			want:   http.Header{"Authorization": {basicAuth(testPhantom)}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			host := c.host
			if host == "" {
				host = "example.com"
			}
			lengths := make(map[string][]int, len(c.header))
			for name, values := range c.header {
				for _, v := range values {
					lengths[name] = append(lengths[name], len(v))
				}
			}

			substituteCredentials(c.header, host, c.creds)

			for name, want := range c.want {
				got := c.header[name]
				if len(got) != len(want) {
					t.Fatalf("header %q = %v, want %v", name, got, want)
				}
				for i := range want {
					if got[i] != want[i] {
						t.Errorf("header %q[%d] = %q, want %q", name, i, got[i], want[i])
					}
					// Framing survives substitution only while the phantom and
					// the real value are the same length, so the property is
					// asserted rather than left to the caller's padding.
					if len(got[i]) != lengths[name][i] {
						t.Errorf("header %q[%d] changed length %d -> %d", name, i, lengths[name][i], len(got[i]))
					}
				}
			}
			if len(c.header) != len(c.want) {
				t.Errorf("header set = %v, want %v", c.header, c.want)
			}
		})
	}
}

// basicAuthFor builds an HTTP Basic value the way a provider's client would,
// with the token in the password slot behind the username that provider uses.
func basicAuthFor(user, token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+token))
}

// Base64 packs three bytes into four characters, so whether a token survives a
// scan for its own encoded form depends on whether the username in front of it
// divides by three. Decoding removes that coincidence from the design: one
// registration of the bare pair covers every provider, whatever it calls its
// username.
func TestSubstituteCredentialsInsideBasicAuth(t *testing.T) {
	cases := []struct {
		name  string
		user  string
		creds Credentials
		value string
		want  string
	}{
		{
			// 15 bytes before the token, which divides by three.
			name: "github, whose username length divides by three",
			user: "x-access-token",
		},
		{
			// 7 bytes, which does not. A pre-encoded registration of the bare
			// phantom would never match this one.
			name: "gitlab, whose username length does not",
			user: "oauth2",
		},
		{
			name: "bitbucket",
			user: "x-token-auth",
		},
		{
			// "Basic base64(":<token>")": no username, token as the password.
			name: "an empty username",
			user: "",
		},
		{
			// "Basic base64("<token>:")": the token IS the username and the
			// password is empty, which is how Stripe, Jira and the GitHub API
			// take one. The token sits at offset zero here rather than behind
			// a prefix, so it is the case a prefix-aware scheme would miss.
			name:  "the token in the username slot, with an empty password",
			value: "Basic " + base64.StdEncoding.EncodeToString([]byte(testPhantom+":")),
			want:  "Basic " + base64.StdEncoding.EncodeToString([]byte(testReal+":")),
		},
		{
			// Malformed per RFC 7617, which requires the colon, but a server
			// that accepts it would still be handed the phantom otherwise.
			name:  "a payload carrying no colon at all",
			value: "Basic " + base64.StdEncoding.EncodeToString([]byte(testPhantom)),
			want:  "Basic " + base64.StdEncoding.EncodeToString([]byte(testReal)),
		},
		{
			// The scheme is case-insensitive per RFC 7235, and Go's own
			// client writes it capitalised, but a hand-rolled one may not.
			name:  "a lowercase scheme token",
			value: "basic " + base64.StdEncoding.EncodeToString([]byte(testPhantom+":")),
			want:  "basic " + base64.StdEncoding.EncodeToString([]byte(testReal+":")),
		},
		{
			name:  "an undeclared host is left alone",
			user:  "x-access-token",
			creds: testCredentials("github.com"),
			value: basicAuthFor("x-access-token", testPhantom),
			want:  basicAuthFor("x-access-token", testPhantom),
		},
		{
			name:  "a Basic value carrying no phantom gains nothing",
			creds: testCredentials("example.com"),
			value: basicAuthFor("x-access-token", "some-other-token"),
			want:  basicAuthFor("x-access-token", "some-other-token"),
		},
		{
			// Rewriting a value this cannot re-encode identically would
			// change bytes it did not understand.
			name:  "a non-canonical encoding is left alone",
			creds: testCredentials("example.com"),
			value: "Basic " + strings.TrimRight(base64.StdEncoding.EncodeToString([]byte("x:"+testPhantom)), "="),
			want:  "Basic " + strings.TrimRight(base64.StdEncoding.EncodeToString([]byte("x:"+testPhantom)), "="),
		},
		{
			name:  "a value that is not base64 at all is left alone",
			creds: testCredentials("example.com"),
			value: "Basic not-base64-!!",
			want:  "Basic not-base64-!!",
		},
		{
			name:  "another scheme is left alone",
			creds: testCredentials("example.com"),
			value: "Bearer " + base64.StdEncoding.EncodeToString([]byte("x:"+testPhantom)),
			want:  "Bearer " + base64.StdEncoding.EncodeToString([]byte("x:"+testPhantom)),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			creds, value, want := c.creds, c.value, c.want
			if creds == nil {
				creds = testCredentials("example.com")
			}
			if value == "" {
				value = basicAuthFor(c.user, testPhantom)
				want = basicAuthFor(c.user, testReal)
			}

			header := http.Header{"Authorization": {value}}
			substituteCredentials(header, "example.com", creds)

			if got := header.Get("Authorization"); got != want {
				t.Errorf("Authorization = %q, want %q", got, want)
			}
			if got := header.Get("Authorization"); len(got) != len(value) {
				t.Errorf("length changed %d -> %d", len(value), len(got))
			}
		})
	}
}

// Substitution sits between the filter and the write to upstream, so what the
// origin received is the only honest witness to both the ordering and the
// hostname it keys on.
func TestHandleMITMSubstitutesAfterTheFilter(t *testing.T) {
	cases := []struct {
		name         string
		cfg          Config
		creds        Credentials
		request      string
		wantStatus   string
		wantUpstream string
	}{
		{
			name:         "a declared host receives the real credential",
			cfg:          wildcardPolicy,
			creds:        encodedCredentials("example.com"),
			request:      "GET /thing HTTP/1.1\r\nHost: example.com\r\nAuthorization: " + basicAuth(testPhantom) + "\r\n\r\n",
			wantStatus:   "HTTP/1.1 200",
			wantUpstream: basicAuth(testReal),
		},
		{
			name:         "an undeclared host receives the phantom",
			cfg:          wildcardPolicy,
			creds:        encodedCredentials("github.com"),
			request:      "GET /thing HTTP/1.1\r\nHost: example.com\r\nAuthorization: " + basicAuth(testPhantom) + "\r\n\r\n",
			wantStatus:   "HTTP/1.1 200",
			wantUpstream: basicAuth(testPhantom),
		},
		{
			// The filter lets the inner Host carry the port, so this is a
			// request whose two hostnames differ in spelling and still pass.
			// It substitutes only while the key is the CONNECT name; the Host
			// header would be read as an undeclared "example.com:443".
			name:         "the tunnel host is the key, not the Host header",
			cfg:          wildcardPolicy,
			creds:        encodedCredentials("example.com"),
			request:      "GET /thing HTTP/1.1\r\nHost: example.com:443\r\nAuthorization: " + basicAuth(testPhantom) + "\r\n\r\n",
			wantStatus:   "HTTP/1.1 200",
			wantUpstream: basicAuth(testReal),
		},
		{
			// The filter refuses POST here, and a refused request must never be
			// dialed, let alone credentialed.
			name:       "a request the filter blocked receives no credential",
			cfg:        getOnlyPolicy,
			creds:      encodedCredentials("example.com"),
			request:    "POST /thing HTTP/1.1\r\nHost: example.com\r\nAuthorization: " + basicAuth(testPhantom) + "\r\nContent-Length: 0\r\n\r\n",
			wantStatus: "HTTP/1.1 403",
		},
		{
			// Naming a declared host inside a tunnel approved for another is
			// the exfiltration path. It is the filter that refuses this one,
			// before substitution is ever reached — the keying itself is what
			// the ":443" case above pins, since that request passes the filter
			// with its two hostnames spelled differently.
			name:       "a request naming another host inside the tunnel receives no credential",
			cfg:        Config{"*": {AllowAll: true}},
			creds:      encodedCredentials("other.example"),
			request:    "GET /thing HTTP/1.1\r\nHost: other.example\r\nAuthorization: " + basicAuth(testPhantom) + "\r\n\r\n",
			wantStatus: "HTTP/1.1 403",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			received := make(chan string, 1)
			upstream, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen upstream: %v", err)
			}
			defer upstream.Close()
			go func() {
				conn, err := upstream.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}
				received <- req.Header.Get("Authorization")
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
			}()

			ca, err := newCertAuthority()
			if err != nil {
				t.Fatalf("new cert authority: %v", err)
			}
			proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen proxy: %v", err)
			}
			defer proxyLn.Close()

			old := credentials
			credentials = c.creds
			defer func() { credentials = old }()

			redirects := Redirects{"example.com": upstream.Addr().String()}
			go func() {
				conn, err := proxyLn.Accept()
				if err != nil {
					return
				}
				handle(conn, c.cfg, ca, redirects)
			}()

			conn, err := net.Dial("tcp", proxyLn.Addr().String())
			if err != nil {
				t.Fatalf("dial proxy: %v", err)
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(10 * time.Second))

			fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")
			br := bufio.NewReader(conn)
			status, err := br.ReadString('\n')
			if err != nil || !strings.HasPrefix(status, "HTTP/1.1 200") {
				t.Fatalf("CONNECT response = %q, err = %v", status, err)
			}
			if _, err := br.ReadString('\n'); err != nil {
				t.Fatalf("read CONNECT header terminator: %v", err)
			}

			tlsConn := tls.Client(conn, &tls.Config{ServerName: "example.com", InsecureSkipVerify: true})
			if err := tlsConn.Handshake(); err != nil {
				t.Fatalf("TLS handshake: %v", err)
			}
			if _, err := io.WriteString(tlsConn, c.request); err != nil {
				t.Fatalf("write request: %v", err)
			}
			status, err = bufio.NewReader(tlsConn).ReadString('\n')
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if !strings.HasPrefix(status, c.wantStatus) {
				t.Fatalf("response = %q, want %q", status, c.wantStatus)
			}

			if c.wantUpstream == "" {
				// The upstream is dialed lazily on the first approved request,
				// so a blocked one leaves this empty by never connecting.
				select {
				case got := <-received:
					t.Errorf("upstream was reached and received %q", got)
				default:
				}
				return
			}
			select {
			case got := <-received:
				if got != c.wantUpstream {
					t.Errorf("upstream received %q, want %q", got, c.wantUpstream)
				}
				if len(got) != len(basicAuth(testPhantom)) {
					t.Errorf("upstream received %d bytes, want the %d the client sent", len(got), len(basicAuth(testPhantom)))
				}
			case <-time.After(10 * time.Second):
				t.Fatal("upstream received nothing")
			}
		})
	}
}

// chunkedReader hands out at most size bytes per Read, so a test can place a
// phantom across a read boundary deliberately rather than hoping one lands
// there. The hold-back buffer is the only reason a split phantom still matches,
// and nothing smaller than a forced split exercises it.
type chunkedReader struct {
	data []byte
	size int
}

func (c *chunkedReader) Read(p []byte) (int, error) {
	if len(c.data) == 0 {
		return 0, io.EOF
	}
	n := min(min(c.size, len(p)), len(c.data))
	copy(p, c.data[:n])
	c.data = c.data[n:]
	return n, nil
}

func (c *chunkedReader) Close() error { return nil }

func readBodyThrough(t *testing.T, body string, chunk int, creds Credentials) string {
	t.Helper()
	src := &chunkedReader{data: []byte(body), size: chunk}
	out, err := io.ReadAll(newSubstitutingBody(src, creds))
	if err != nil {
		t.Fatalf("read substituted body: %v", err)
	}
	return string(out)
}

func TestSubstitutingBody(t *testing.T) {
	cases := []struct {
		name  string
		creds Credentials
		body  string
		want  string
	}{
		{
			name:  "a phantom in the body is replaced",
			creds: testCredentials("example.com"),
			body:  "pack" + testPhantom + "file",
			want:  "pack" + testReal + "file",
		},
		{
			name:  "a body carrying no phantom is untouched",
			creds: testCredentials("example.com"),
			body:  "just some packfile bytes",
			want:  "just some packfile bytes",
		},
		{
			name:  "an empty body stays empty",
			creds: testCredentials("example.com"),
			body:  "",
			want:  "",
		},
		{
			name:  "a phantom at the very start is replaced",
			creds: testCredentials("example.com"),
			body:  testPhantom + "trailing",
			want:  testReal + "trailing",
		},
		{
			// The tail the reader holds back is flushed only at EOF, so a
			// phantom ending exactly at the last byte is the case that a
			// missing flush would silently drop.
			name:  "a phantom at the very end is replaced",
			creds: testCredentials("example.com"),
			body:  "leading" + testPhantom,
			want:  "leading" + testReal,
		},
		{
			name:  "every occurrence is replaced",
			creds: testCredentials("example.com"),
			body:  testPhantom + "|" + testPhantom + "|" + testPhantom,
			want:  testReal + "|" + testReal + "|" + testReal,
		},
		{
			name:  "no credentials leaves the body untouched",
			creds: nil,
			body:  "pack" + testPhantom + "file",
			want:  "pack" + testPhantom + "file",
		},
		{
			name: "each credential is replaced with its own value",
			creds: append(
				credentialFor(testPhantom, testReal, "example.com"),
				credentialFor("SECOND-PHANTOM", "second-realval", "example.com")...,
			),
			body: testPhantom + " and SECOND-PHANTOM",
			want: testReal + " and second-realval",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Read sizes from one byte up past the phantom, so the same body is
			// split at every offset a phantom can straddle.
			for _, chunk := range []int{1, 2, 3, 7, len(testPhantom) - 1, len(testPhantom), len(testPhantom) + 1, 512} {
				got := readBodyThrough(t, c.body, chunk, c.creds)
				if got != c.want {
					t.Errorf("chunk %d: body = %q, want %q", chunk, got, c.want)
				}
			}
		})
	}
}

// The hold-back buffer exists for exactly this: a phantom whose bytes arrive in
// two reads must still match across the join. Every split offset is exercised,
// since an off-by-one in the retained tail fails only at one of them.
func TestSubstitutingBodyMatchesAcrossEveryChunkBoundary(t *testing.T) {
	creds := testCredentials("example.com")
	for split := 1; split < len(testPhantom); split++ {
		t.Run(fmt.Sprintf("split after %d bytes", split), func(t *testing.T) {
			src := &boundaryReader{parts: []string{
				"leading" + testPhantom[:split],
				testPhantom[split:] + "trailing",
			}}
			out, err := io.ReadAll(newSubstitutingBody(src, creds))
			if err != nil {
				t.Fatalf("read substituted body: %v", err)
			}
			if want := "leading" + testReal + "trailing"; string(out) != want {
				t.Errorf("body = %q, want %q", out, want)
			}
		})
	}
}

// boundaryReader returns exactly one part per Read, so the split between two
// parts is guaranteed to be a read boundary.
type boundaryReader struct {
	parts []string
}

func (b *boundaryReader) Read(p []byte) (int, error) {
	if len(b.parts) == 0 {
		return 0, io.EOF
	}
	n := copy(p, b.parts[0])
	b.parts[0] = b.parts[0][n:]
	if b.parts[0] == "" {
		b.parts = b.parts[1:]
	}
	return n, nil
}

func (b *boundaryReader) Close() error { return nil }

// A body is forwarded as it arrives, not gathered first. The source here never
// ends, so a read that only returns once the whole body is in hand would block
// here rather than fail — which is the failure this ticket exists to avoid.
func TestSubstitutingBodyStreamsBeforeTheSourceEnds(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()
	body := newSubstitutingBody(pr, testCredentials("example.com"))

	const sent = "a packfile chunk that has not finished arriving"
	go func() { pw.Write([]byte(sent)) }()

	// Everything but the withheld tail, which is the one part that cannot be
	// emitted until something follows it or the source ends.
	want := len(sent) - (len(testPhantom) - 1)
	got := make([]byte, len(sent))
	read := make(chan int, 1)
	go func() {
		n, _ := io.ReadAtLeast(body, got, want)
		read <- n
	}()

	select {
	case n := <-read:
		if n < want {
			t.Errorf("read %d bytes before the source ended, want at least %d", n, want)
		}
		if string(got[:n]) != sent[:n] {
			t.Errorf("read %q, want the prefix of %q", got[:n], sent)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the body blocked until the source ended: it is being buffered whole")
	}
}

// fillerReader produces n bytes without holding them, so the test does not
// buffer whole what it is asserting the proxy does not buffer whole.
type fillerReader struct {
	left    int
	phantom string
	at      int
	pos     int
}

func (f *fillerReader) Read(p []byte) (int, error) {
	if f.left == 0 {
		return 0, io.EOF
	}
	n := min(len(p), f.left)
	for i := 0; i < n; i++ {
		switch offset := f.pos + i - f.at; {
		case offset >= 0 && offset < len(f.phantom):
			p[i] = f.phantom[offset]
		default:
			p[i] = byte('a' + (f.pos+i)%26)
		}
	}
	f.pos += n
	f.left -= n
	return n, nil
}

func (f *fillerReader) Close() error { return nil }

// bodyScan is what a reader saw without keeping any of it, so a test may
// assert on a body far larger than it wants in memory.
type bodyScan struct {
	length     int
	sum        [32]byte
	sawPhantom bool
	sawReal    bool
	err        error
}

// scanStream reads r to its end, carrying the longest value it looks for minus
// one byte across each read so that one spanning two reads is still seen.
func scanStream(r io.Reader) bodyScan {
	var scan bodyScan
	sum := sha256.New()
	keep := max(len(testPhantom), len(testReal)) - 1
	buf := make([]byte, 32<<10)
	var window []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			sum.Write(buf[:n])
			window = append(window, buf[:n]...)
			scan.sawPhantom = scan.sawPhantom || bytes.Contains(window, []byte(testPhantom))
			scan.sawReal = scan.sawReal || bytes.Contains(window, []byte(testReal))
			if len(window) > keep {
				window = append(window[:0], window[len(window)-keep:]...)
			}
			scan.length += n
		}
		if err != nil {
			scan.err = err
			break
		}
	}
	copy(scan.sum[:], sum.Sum(nil))
	return scan
}

// No size ceiling: a body far larger than any buffer the proxy could hold
// streams through intact, with the phantom placed at an offset no read size
// divides, so it lands mid-chunk rather than on a boundary by luck.
func TestSubstitutingBodyHasNoSizeCeiling(t *testing.T) {
	const size = 64 << 20
	const at = size / 2 % 4093
	src := &fillerReader{left: size, phantom: testPhantom, at: at}

	scan := scanStream(newSubstitutingBody(src, testCredentials("example.com")))

	if scan.err != io.EOF {
		t.Fatalf("read substituted body: %v", scan.err)
	}
	if scan.length != size {
		t.Errorf("streamed %d bytes, want %d", scan.length, size)
	}
	if !scan.sawReal {
		t.Error("the real value never appeared: the phantom was not substituted")
	}
	if scan.sawPhantom {
		t.Error("the phantom survived into the forwarded body")
	}
}

// errAfterReader delivers its data and then fails, which is what a client that
// dies part-way through a push looks like from here.
type errAfterReader struct {
	data   string
	err    error
	closed bool
}

func (e *errAfterReader) Read(p []byte) (int, error) {
	if e.data == "" {
		return 0, e.err
	}
	n := copy(p, e.data)
	e.data = e.data[n:]
	return n, nil
}

func (e *errAfterReader) Close() error {
	e.closed = true
	return nil
}

// stalledReader makes no progress and never fails. Read's loop ends only on
// bytes or an error, so without a bound this spins rather than returning.
type stalledReader struct{}

func (stalledReader) Read(p []byte) (int, error) { return 0, nil }

func (stalledReader) Close() error { return nil }

func TestSubstitutingBodyGivesUpOnAStalledSource(t *testing.T) {
	body := newSubstitutingBody(stalledReader{}, testCredentials("example.com"))

	done := make(chan error, 1)
	go func() {
		_, err := body.Read(make([]byte, 16))
		done <- err
	}()

	select {
	case err := <-done:
		if !errors.Is(err, io.ErrNoProgress) {
			t.Errorf("error = %v, want %v", err, io.ErrNoProgress)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Read spun on a source making no progress")
	}
}

func TestSubstitutingBodySurfacesASourceFailure(t *testing.T) {
	want := errors.New("client went away mid-body")
	src := &errAfterReader{data: "pack" + testPhantom + "file", err: want}
	body := newSubstitutingBody(src, testCredentials("example.com"))

	got, err := io.ReadAll(body)
	if !errors.Is(err, want) {
		t.Errorf("error = %v, want %v", err, want)
	}
	// The bytes that did arrive are still substituted and still handed on.
	// Dropping them would turn a read failure into a silent truncation, which
	// req.Write would then send under the length the client declared.
	if w := "pack" + testReal + "file"; string(got) != w {
		t.Errorf("body before the error = %q, want %q", got, w)
	}
	if err := body.Close(); err != nil {
		t.Errorf("Close errored: %v", err)
	}
	if !src.closed {
		t.Error("Close did not reach the source, so the client connection would leak")
	}
}

// writeRequest runs a request through the substituting path and returns the
// bytes that reach the upstream, so framing is asserted on the wire rather
// than on the struct fields that produce it.
func writeRequest(t *testing.T, raw, host string, creds Credentials) string {
	t.Helper()
	req := readRequest(t, raw)
	substituteRequestBody(req, host, creds)
	var buf bytes.Buffer
	if err := req.Write(&buf); err != nil {
		t.Fatalf("write request: %v", err)
	}
	return buf.String()
}

func postToHost(host, body string) string {
	return fmt.Sprintf("POST /push HTTP/1.1\r\nHost: %s\r\nContent-Length: %d\r\n\r\n%s", host, len(body), body)
}

func postWithBody(body string) string {
	return postToHost("example.com", body)
}

// encodedPost declares an encoding without applying one. What gates the skip
// is the header, not whether the bytes happen to be compressible: leaving the
// phantom literal means a proxy that ignored the header would substitute it,
// and the test would catch that.
func encodedPost(encoding, body string) string {
	return fmt.Sprintf("POST /push HTTP/1.1\r\nHost: example.com\r\nContent-Encoding: %s\r\nContent-Length: %d\r\n\r\n%s",
		encoding, len(body), body)
}

func chunkedPost(body string) string {
	return fmt.Sprintf("POST /push HTTP/1.1\r\nHost: example.com\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n%s\r\n0\r\n\r\n", len(body), body)
}

func TestSubstituteRequestBodyFraming(t *testing.T) {
	// Shorter than the phantom, so padding cannot reach it and the length the
	// client declared stops being true under substitution.
	shortReal := "ghp_short"

	cases := []struct {
		name        string
		raw         string
		host        string
		creds       Credentials
		wantBody    string
		wantChunked bool
		wantLength  string
	}{
		{
			name:       "an equal-length value keeps the length the client declared",
			raw:        postWithBody("pack" + testPhantom + "file"),
			host:       "example.com",
			creds:      testCredentials("example.com"),
			wantBody:   "pack" + testReal + "file",
			wantLength: fmt.Sprintf("Content-Length: %d", len("pack"+testPhantom+"file")),
		},
		{
			// Padding only works upwards, so a real value shorter than the
			// phantom cannot be length-matched. Re-framed rather than sent
			// under a Content-Length that no longer holds.
			name:        "a shorter value re-frames as chunked",
			raw:         postWithBody("pack" + testPhantom + "file"),
			host:        "example.com",
			creds:       credentialFor(testPhantom, shortReal, "example.com"),
			wantBody:    "pack" + shortReal + "file",
			wantChunked: true,
		},
		{
			name:        "a chunked body stays chunked",
			raw:         chunkedPost("pack" + testPhantom + "file"),
			host:        "example.com",
			creds:       testCredentials("example.com"),
			wantBody:    "pack" + testReal + "file",
			wantChunked: true,
		},
		{
			name:       "an undeclared host leaves the body and its framing alone",
			raw:        postWithBody("pack" + testPhantom + "file"),
			host:       "example.com",
			creds:      credentialFor(testPhantom, shortReal, "github.com"),
			wantBody:   "pack" + testPhantom + "file",
			wantLength: fmt.Sprintf("Content-Length: %d", len("pack"+testPhantom+"file")),
		},
		{
			name:       "an empty host list substitutes nowhere",
			raw:        postWithBody("pack" + testPhantom + "file"),
			host:       "example.com",
			creds:      credentialFor(testPhantom, shortReal),
			wantBody:   "pack" + testPhantom + "file",
			wantLength: fmt.Sprintf("Content-Length: %d", len("pack"+testPhantom+"file")),
		},
		{
			name:       "a request with no body is left alone",
			raw:        "GET /thing HTTP/1.1\r\nHost: example.com\r\n\r\n",
			host:       "example.com",
			creds:      credentialFor(testPhantom, shortReal, "example.com"),
			wantBody:   "",
			wantLength: "",
		},
		{
			// The scan cannot see into an encoded body, so it is forwarded as
			// it arrived and the phantom travels inert.
			name:       "an encoded body is not scanned",
			raw:        encodedPost("gzip", "pack"+testPhantom+"file"),
			host:       "example.com",
			creds:      testCredentials("example.com"),
			wantBody:   "pack" + testPhantom + "file",
			wantLength: fmt.Sprintf("Content-Length: %d", len("pack"+testPhantom+"file")),
		},
		{
			// The encoding check must run BEFORE the length decision. A body
			// nothing will rewrite must keep the framing the client gave it,
			// even when the declared credential could not be length-matched.
			name:       "an encoded body keeps its framing even when lengths differ",
			raw:        encodedPost("gzip", "pack"+testPhantom+"file"),
			host:       "example.com",
			creds:      credentialFor(testPhantom, shortReal, "example.com"),
			wantBody:   "pack" + testPhantom + "file",
			wantLength: fmt.Sprintf("Content-Length: %d", len("pack"+testPhantom+"file")),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wire := writeRequest(t, c.raw, c.host, c.creds)

			if got := strings.Contains(wire, "Transfer-Encoding: chunked"); got != c.wantChunked {
				t.Errorf("chunked = %v, want %v, in:\n%s", got, c.wantChunked, wire)
			}
			if c.wantChunked && strings.Contains(wire, "Content-Length:") {
				t.Errorf("a chunked request still carries a Content-Length:\n%s", wire)
			}
			if c.wantLength != "" && !strings.Contains(wire, c.wantLength) {
				t.Errorf("want %q, in:\n%s", c.wantLength, wire)
			}

			// Parsed back rather than string-matched, so a body that is chunked
			// but misframed fails here instead of passing on its contents.
			out, err := http.ReadRequest(bufio.NewReader(strings.NewReader(wire)))
			if err != nil {
				t.Fatalf("reparse forwarded request: %v", err)
			}
			body, err := io.ReadAll(out.Body)
			if err != nil {
				t.Fatalf("read forwarded body: %v", err)
			}
			if string(body) != c.wantBody {
				t.Errorf("forwarded body = %q, want %q", body, c.wantBody)
			}
		})
	}
}

// A skipped substitution is otherwise a 401 with nothing in the exchange to
// explain it, which is the unexplained-non-substitution failure the README's
// base64 note already warns about. It says which host and no more.
func TestSubstituteRequestBodyReportsASkippedEncodedBody(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		creds    Credentials
		wantWarn bool
	}{
		{
			name:     "an encoded body a credential was declared for",
			raw:      encodedPost("gzip", "pack"+testPhantom+"file"),
			creds:    testCredentials("example.com"),
			wantWarn: true,
		},
		{
			// Nothing would have been substituted here anyway, so a warning
			// would fire on every compressed request the sandbox ever makes.
			name:  "an encoded body for an undeclared host stays quiet",
			raw:   encodedPost("gzip", "pack"+testPhantom+"file"),
			creds: testCredentials("github.com"),
		},
		{
			name:  "an unencoded body stays quiet",
			raw:   postWithBody("pack" + testPhantom + "file"),
			creds: testCredentials("example.com"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stderr := captureStderr(t)
			writeRequest(t, c.raw, "example.com", c.creds)
			logged := stderr()

			if got := strings.Contains(logged, "encoded request body"); got != c.wantWarn {
				t.Errorf("warned = %v, want %v, log = %q", got, c.wantWarn, logged)
			}
			if !c.wantWarn {
				return
			}
			if !strings.Contains(logged, "example.com") {
				t.Errorf("the warning does not name the host: %q", logged)
			}
			// Ticket 03's rule: hostnames reach the log, header values do not.
			for _, secret := range []string{"gzip", testPhantom, testReal} {
				if strings.Contains(logged, secret) {
					t.Errorf("the warning carries %q: %q", secret, logged)
				}
			}
		})
	}
}

// The proxy logs hostnames. A body is the largest thing it now touches and the
// one most likely to carry a secret, so none of it may reach stderr.
func TestSubstituteRequestBodyLogsNothing(t *testing.T) {
	stderr := captureStderr(t)
	body := "pack" + testPhantom + "file"
	wire := writeRequest(t, postWithBody(body), "example.com", testCredentials("example.com"))
	logged := stderr()

	if logged != "" {
		t.Errorf("stderr = %q, want nothing", logged)
	}
	if !strings.Contains(wire, testReal) {
		t.Fatal("the test did not exercise a substitution")
	}
}

// receivedRequest is the upstream's view of what the proxy forwarded. The body
// is read straight off the wire, never back through substitutingBody: an
// oracle built from the code under test cannot catch that code dropping or
// duplicating bytes, because it would drop or duplicate them too.
type receivedRequest struct {
	authorization string
	contentLength int64
	chunked       bool
	body          bodyScan
}

// mitmUpstream accepts one request, scanning the body as it arrives rather
// than holding it, so a test may forward more than it can keep.
func mitmUpstream(t *testing.T, ln net.Listener, got chan<- receivedRequest) {
	t.Helper()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		req, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			return
		}
		seen := receivedRequest{
			authorization: req.Header.Get("Authorization"),
			contentLength: req.ContentLength,
		}
		for _, encoding := range req.TransferEncoding {
			if strings.EqualFold(encoding, "chunked") {
				seen.chunked = true
			}
		}
		seen.body = scanStream(req.Body)
		got <- seen
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	}()
}

// mitmExchange drives one request through the real CONNECT, TLS and filter
// path, so body substitution is asserted where it actually runs rather than on
// a directly-called function.
func mitmExchange(t *testing.T, cfg Config, creds Credentials, send func(io.Writer) error) (string, chan receivedRequest) {
	t.Helper()
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen upstream: %v", err)
	}
	t.Cleanup(func() { upstream.Close() })
	received := make(chan receivedRequest, 1)
	mitmUpstream(t, upstream, received)

	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}
	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen proxy: %v", err)
	}
	t.Cleanup(func() { proxyLn.Close() })

	old := credentials
	credentials = creds
	t.Cleanup(func() { credentials = old })

	redirects := Redirects{"example.com": upstream.Addr().String()}
	go func() {
		conn, err := proxyLn.Accept()
		if err != nil {
			return
		}
		handle(conn, cfg, ca, redirects)
	}()

	conn, err := net.Dial("tcp", proxyLn.Addr().String())
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	conn.SetDeadline(time.Now().Add(120 * time.Second))

	fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	if err != nil || !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("CONNECT response = %q, err = %v", status, err)
	}
	if _, err := br.ReadString('\n'); err != nil {
		t.Fatalf("read CONNECT header terminator: %v", err)
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "example.com", InsecureSkipVerify: true})
	if err := tlsConn.Handshake(); err != nil {
		t.Fatalf("TLS handshake: %v", err)
	}
	if err := send(tlsConn); err != nil {
		t.Fatalf("write request: %v", err)
	}
	status, err = bufio.NewReader(tlsConn).ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return status, received
}

func sendString(raw string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, raw)
		return err
	}
}

func TestHandleMITMSubstitutesTheRequestBody(t *testing.T) {
	body := "PACK" + strings.Repeat("x", 100) + testPhantom + strings.Repeat("y", 100)

	cases := []struct {
		name        string
		creds       Credentials
		raw         string
		wantReal    bool
		wantChunked bool
		wantLength  int64
	}{
		{
			name:       "a declared host receives the real value in the body",
			creds:      testCredentials("example.com"),
			raw:        postWithBody(body),
			wantReal:   true,
			wantLength: int64(len(body)),
		},
		{
			name:       "an undeclared host receives the phantom in the body",
			creds:      testCredentials("github.com"),
			raw:        postWithBody(body),
			wantLength: int64(len(body)),
		},
		{
			name:        "a chunked body is substituted and stays chunked",
			creds:       testCredentials("example.com"),
			raw:         chunkedPost(body),
			wantReal:    true,
			wantChunked: true,
			wantLength:  -1,
		},
		{
			// The filter lets the inner Host carry the port, so this request
			// passes with its two hostnames spelled differently. It is the
			// case that pins the keying: read req.Host here instead of the
			// CONNECT name and the body sees an undeclared "example.com:443".
			name:       "the tunnel host is the key, not the Host header",
			creds:      testCredentials("example.com"),
			raw:        postToHost("example.com:443", body),
			wantReal:   true,
			wantLength: int64(len(body)),
		},
		{
			// The scan cannot see into it, so the phantom travels inert and
			// the upstream receives the body byte for byte as the client sent
			// it — Content-Length included.
			name:       "an encoded body reaches the upstream untouched",
			creds:      testCredentials("example.com"),
			raw:        encodedPost("gzip", body),
			wantLength: int64(len(body)),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, received := mitmExchange(t, wildcardPolicy, c.creds, sendString(c.raw))
			if !strings.HasPrefix(status, "HTTP/1.1 200") {
				t.Fatalf("response = %q, want 200", status)
			}
			select {
			case seen := <-received:
				if seen.body.sawReal != c.wantReal {
					t.Errorf("upstream saw the real value = %v, want %v", seen.body.sawReal, c.wantReal)
				}
				if seen.body.sawPhantom == c.wantReal {
					t.Errorf("upstream saw the phantom = %v, want %v", seen.body.sawPhantom, !c.wantReal)
				}
				if seen.chunked != c.wantChunked {
					t.Errorf("upstream saw chunked = %v, want %v", seen.chunked, c.wantChunked)
				}
				if seen.contentLength != c.wantLength {
					t.Errorf("upstream saw Content-Length %d, want %d", seen.contentLength, c.wantLength)
				}
				if seen.body.length != len(body) {
					t.Errorf("upstream received %d body bytes, want the %d the client sent", seen.body.length, len(body))
				}
			case <-time.After(30 * time.Second):
				t.Fatal("upstream received nothing")
			}
		})
	}
}

// The push this feature exists for is a packfile of unbounded size. Eight
// mebibytes crosses several hundred TLS records and read boundaries, with the
// phantom at an offset none of them divides.
func TestHandleMITMSubstitutesALargeBody(t *testing.T) {
	const size = 8 << 20
	const at = 5_000_003 % (size - 64)

	send := func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "POST /push HTTP/1.1\r\nHost: example.com\r\nContent-Length: %d\r\n\r\n", size); err != nil {
			return err
		}
		_, err := io.Copy(w, &fillerReader{left: size, phantom: testPhantom, at: at})
		return err
	}

	status, received := mitmExchange(t, wildcardPolicy, testCredentials("example.com"), send)
	if !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("response = %q, want 200", status)
	}

	want := sha256.Sum256(expectedLargeBody(size, at))
	select {
	case seen := <-received:
		if seen.body.length != size {
			t.Fatalf("upstream received %d body bytes, want %d", seen.body.length, size)
		}
		if seen.contentLength != size {
			t.Errorf("upstream saw Content-Length %d, want %d: substitution changed the framing", seen.contentLength, size)
		}
		if !seen.body.sawReal {
			t.Error("upstream never saw the real value")
		}
		if seen.body.sawPhantom {
			t.Error("the phantom survived into the forwarded body")
		}
		if seen.body.sum != want {
			t.Error("the forwarded body differs from the substituted body byte for byte")
		}
	case <-time.After(120 * time.Second):
		t.Fatal("upstream received nothing")
	}
}

// The proxy logs hostnames. Neither the phantom nor the value it stands for
// may reach stderr from the body path, through the whole MITM loop rather
// than through substituteRequestBody alone.
func TestHandleMITMLogsNoBodyContent(t *testing.T) {
	stderr := captureStderr(t)
	body := "PACK" + testPhantom + "TAIL"
	status, received := mitmExchange(t, wildcardPolicy, testCredentials("example.com"), sendString(postWithBody(body)))

	var seen receivedRequest
	select {
	case seen = <-received:
	case <-time.After(30 * time.Second):
		stderr()
		t.Fatal("upstream received nothing")
	}
	logged := stderr()

	if !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("response = %q, want 200", status)
	}
	if !seen.body.sawReal {
		t.Fatal("the test did not exercise a substitution")
	}
	if strings.Contains(logged, testPhantom) || strings.Contains(logged, testReal) {
		t.Errorf("the log carries body content: %q", logged)
	}
}

func expectedLargeBody(size, at int) []byte {
	out := make([]byte, size)
	src := &fillerReader{left: size, phantom: testReal, at: at}
	io.ReadFull(src, out)
	return out
}

// The tunnel carries opaque frames the proxy has no business rewriting, and it
// is read straight off the client connection rather than through a body. A
// phantom in a frame must cross untouched.
func TestWebSocketTunnelIsNotSubstituted(t *testing.T) {
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen upstream: %v", err)
	}
	defer upstream.Close()
	go func() {
		conn, err := upstream.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		br := bufio.NewReader(conn)
		if _, err := http.ReadRequest(br); err != nil {
			return
		}
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		io.Copy(conn, br)
	}()

	ca, err := newCertAuthority()
	if err != nil {
		t.Fatalf("new cert authority: %v", err)
	}
	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen proxy: %v", err)
	}
	defer proxyLn.Close()

	old := credentials
	credentials = testCredentials("example.com")
	defer func() { credentials = old }()

	redirects := Redirects{"example.com": upstream.Addr().String()}
	go func() {
		conn, err := proxyLn.Accept()
		if err != nil {
			return
		}
		handle(conn, wildcardPolicy, ca, redirects)
	}()

	conn, err := net.Dial("tcp", proxyLn.Addr().String())
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	fmt.Fprintf(conn, "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")
	br := bufio.NewReader(conn)
	if status, err := br.ReadString('\n'); err != nil || !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("CONNECT response = %q, err = %v", status, err)
	}
	if _, err := br.ReadString('\n'); err != nil {
		t.Fatalf("read CONNECT header terminator: %v", err)
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "example.com", InsecureSkipVerify: true})
	if err := tlsConn.Handshake(); err != nil {
		t.Fatalf("TLS handshake: %v", err)
	}
	fmt.Fprintf(tlsConn, "GET /ws HTTP/1.1\r\nHost: example.com\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
	tlsBr := bufio.NewReader(tlsConn)
	if status, err := tlsBr.ReadString('\n'); err != nil || !strings.HasPrefix(status, "HTTP/1.1 101") {
		t.Fatalf("upgrade response = %q, err = %v", status, err)
	}
	for {
		line, err := tlsBr.ReadString('\n')
		if err != nil {
			t.Fatalf("read upgrade headers: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}

	if _, err := io.WriteString(tlsConn, testPhantom); err != nil {
		t.Fatalf("write through tunnel: %v", err)
	}
	echo := make([]byte, len(testPhantom))
	if _, err := io.ReadFull(tlsBr, echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(echo) != testPhantom {
		t.Errorf("the tunnel rewrote its frames: echo = %q, want %q", echo, testPhantom)
	}
}
