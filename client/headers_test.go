package client

import (
	"net/http"
	"testing"
)

// SPEC: MS-ASHTTP/headers.protocol-version
// SPEC: MS-ASHTTP/headers.content-type
// SPEC: MS-ASHTTP/headers.user-agent
func TestApplyMandatoryHeaders(t *testing.T) {
	h := http.Header{}
	ApplyMandatoryHeaders(h, HeaderOptions{
		ProtocolVersion: "14.1",
		UserAgent:       "go-activesync/0.0.1",
	})
	if got := h.Get("MS-ASProtocolVersion"); got != "14.1" {
		t.Errorf("MS-ASProtocolVersion = %q", got)
	}
	if got := h.Get("Content-Type"); got != "application/vnd.ms-sync.wbxml" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := h.Get("User-Agent"); got != "go-activesync/0.0.1" {
		t.Errorf("User-Agent = %q", got)
	}
	if got := h.Get("Accept"); got != "application/vnd.ms-sync.wbxml" {
		t.Errorf("Accept = %q", got)
	}
}

// SPEC: MS-ASHTTP/headers.policy-key
func TestApplyMandatoryHeaders_PolicyKey(t *testing.T) {
	h := http.Header{}
	ApplyMandatoryHeaders(h, HeaderOptions{
		ProtocolVersion: "14.1",
		UserAgent:       "go-activesync/0.0.1",
		PolicyKey:       "12345",
	})
	if got := h.Get("X-MS-PolicyKey"); got != "12345" {
		t.Errorf("X-MS-PolicyKey = %q", got)
	}
}

// SPEC: MS-ASHTTP/headers.accept-language
func TestApplyMandatoryHeaders_AcceptLanguage(t *testing.T) {
	h := http.Header{}
	ApplyMandatoryHeaders(h, HeaderOptions{
		ProtocolVersion: "14.1",
		UserAgent:       "go-activesync/0.0.1",
		AcceptLanguage:  "en-US",
	})
	if got := h.Get("Accept-Language"); got != "en-US" {
		t.Errorf("Accept-Language = %q", got)
	}
}
