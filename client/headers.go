package client

import "net/http"

// ContentTypeWBXML is the MIME type EAS uses for WBXML payloads.
const ContentTypeWBXML = "application/vnd.ms-sync.wbxml"

// HeaderOptions describes the values used to populate the mandatory EAS
// headers on a request.
type HeaderOptions struct {
	ProtocolVersion string // e.g. "14.1"
	UserAgent       string // identifies the client
	PolicyKey       string // current policy key, "0" before initial Provision
	AcceptLanguage  string // optional, e.g. "en-US"
}

// ApplyMandatoryHeaders sets the mandatory MS-ASHTTP headers on h. Any
// existing values for managed headers are overwritten so the contract is
// deterministic for upstream code.
func ApplyMandatoryHeaders(h http.Header, opts HeaderOptions) {
	h.Set("MS-ASProtocolVersion", opts.ProtocolVersion)
	h.Set("Content-Type", ContentTypeWBXML)
	h.Set("Accept", ContentTypeWBXML)
	h.Set("User-Agent", opts.UserAgent)
	if opts.PolicyKey != "" {
		h.Set("X-MS-PolicyKey", opts.PolicyKey)
	}
	if opts.AcceptLanguage != "" {
		h.Set("Accept-Language", opts.AcceptLanguage)
	}
}
