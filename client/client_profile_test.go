package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/remdev/go-activesync/eas"
	"github.com/remdev/go-activesync/wbxml"
)

// SPEC: MS-ASHTTP/client.transport.force-http11
func TestNew_ForceHTTP11_TLSNextProtoEmptyMap(t *testing.T) {
	c, err := New(Config{
		BaseURL:     "https://example.invalid/Microsoft-Server-ActiveSync",
		DeviceID:    "d",
		DeviceType:  "t",
		ForceHTTP11: true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tr, ok := c.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport type got %T", c.HTTPClient.Transport)
	}
	if tr.TLSNextProto == nil {
		t.Fatal("TLSNextProto is nil; expected non-nil empty map to disable HTTP/2 ALPN")
	}
	if len(tr.TLSNextProto) != 0 {
		t.Fatalf("TLSNextProto len = %d; want 0", len(tr.TLSNextProto))
	}
}

// SPEC: MS-ASHTTP/client.transport.force-http11
func TestNew_ForceHTTP11_WithCustomHTTPClientIgnored(t *testing.T) {
	custom := &http.Client{Transport: http.DefaultTransport}
	c, err := New(Config{
		BaseURL:     "http://example.invalid/Microsoft-Server-ActiveSync",
		DeviceID:    "d",
		DeviceType:  "t",
		HTTPClient:  custom,
		ForceHTTP11: true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.HTTPClient != custom {
		t.Fatal("ForceHTTP11 must not replace a caller-supplied HTTPClient")
	}
}

// SPEC: MS-ASHTTP/client.extra-headers-merge
func TestProvision_OutboundExtraHeaders(t *testing.T) {
	var saw string
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.Header.Get("X-Integration-Probe")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var req eas.ProvisionRequest
		if err := wbxml.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		calls++
		var resp eas.ProvisionResponse
		switch calls {
		case 1:
			resp = eas.ProvisionResponse{
				Status: int32(eas.StatusSuccess),
				Policies: eas.PoliciesResponse{
					Policy: []eas.PolicyResponse{{
						PolicyType: eas.PolicyTypeWBXML,
						PolicyKey:  "temp-key",
						Status:     int32(eas.StatusSuccess),
						Data: &eas.EASProvisionDoc{
							DevicePasswordEnabled:              1,
							MinDevicePasswordLength:            4,
							MaxInactivityTimeDeviceLock:        900,
							MaxDevicePasswordFailedAttempts:    8,
							AllowSimpleDevicePassword:          1,
							AllowStorageCard:                   1,
							AllowCamera:                        1,
							RequireDeviceEncryption:            0,
							AlphanumericDevicePasswordRequired: 0,
						},
					}},
				},
			}
		case 2:
			resp = eas.ProvisionResponse{
				Status: int32(eas.StatusSuccess),
				Policies: eas.PoliciesResponse{
					Policy: []eas.PolicyResponse{{
						PolicyType: eas.PolicyTypeWBXML,
						PolicyKey:  "final-key",
						Status:     int32(eas.StatusSuccess),
					}},
				},
			}
		default:
			http.Error(w, "unexpected call", 500)
			return
		}
		data, err := wbxml.Marshal(&resp)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", ContentTypeWBXML)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	h := http.Header{}
	h.Set("X-Integration-Probe", "present")
	c, err := New(Config{
		BaseURL:      srv.URL + EndpointPath,
		HTTPClient:   srv.Client(),
		Auth:         &BasicAuth{Username: "u", Password: "p"},
		DeviceID:     "DEV",
		DeviceType:   "Outlook",
		UserAgent:    "ua-test/1",
		Locale:       0x0419,
		ExtraHeaders: h,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.Provision(context.Background(), "user@example.com"); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if saw != "present" {
		t.Fatalf("X-Integration-Probe = %q", saw)
	}
}

// SPEC: MS-ASCMD/status.165.device-information-required
// SPEC: MS-ASPROV/provision.device-information
func TestProvision_IncludesDefaultOutlookDeviceInformation(t *testing.T) {
	var first eas.ProvisionRequest
	var second eas.ProvisionRequest
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req eas.ProvisionRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if err := wbxml.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		calls++
		if calls == 1 {
			first = req
			writeWBXML(t, w, &eas.ProvisionResponse{
				Status: int32(eas.StatusSuccess),
				Policies: eas.PoliciesResponse{
					Policy: []eas.PolicyResponse{{
						PolicyType: eas.PolicyTypeWBXML,
						PolicyKey:  "temp-key",
						Status:     int32(eas.StatusSuccess),
						Data: &eas.EASProvisionDoc{
							DevicePasswordEnabled:              1,
							MinDevicePasswordLength:            4,
							MaxInactivityTimeDeviceLock:        900,
							MaxDevicePasswordFailedAttempts:    8,
							AllowSimpleDevicePassword:          1,
							AllowStorageCard:                   1,
							AllowCamera:                        1,
							RequireDeviceEncryption:            0,
							AlphanumericDevicePasswordRequired: 0,
						},
					}},
				},
			})
			return
		}
		second = req
		writeWBXML(t, w, &eas.ProvisionResponse{
			Status: int32(eas.StatusSuccess),
			Policies: eas.PoliciesResponse{
				Policy: []eas.PolicyResponse{{
					PolicyType: eas.PolicyTypeWBXML,
					PolicyKey:  "final-key",
					Status:     int32(eas.StatusSuccess),
				}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := New(Config{
		BaseURL:    srv.URL + EndpointPath,
		HTTPClient: srv.Client(),
		Auth:       &BasicAuth{Username: "u", Password: "p"},
		DeviceID:   "DEV",
		DeviceType: "Outlook",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.Provision(context.Background(), "user@example.com"); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}

	if first.DeviceInformation == nil {
		t.Fatal("initial ProvisionRequest.DeviceInformation is nil")
	}
	if second.DeviceInformation == nil {
		t.Fatal("ack ProvisionRequest.DeviceInformation is nil")
	}
	got := first.DeviceInformation.Set
	if got.Model != defaultDeviceModel {
		t.Errorf("Model = %q", got.Model)
	}
	if got.FriendlyName != defaultDeviceModel {
		t.Errorf("FriendlyName = %q", got.FriendlyName)
	}
	if got.OS != "iOS 17.5" {
		t.Errorf("OS = %q", got.OS)
	}
	if got.OSLanguage != "ru" {
		t.Errorf("OSLanguage = %q", got.OSLanguage)
	}
	if got.UserAgent != "Outlook-iOS-Android/1.0" {
		t.Errorf("UserAgent = %q", got.UserAgent)
	}
	if got.MobileOperator != "Apple" {
		t.Errorf("MobileOperator = %q", got.MobileOperator)
	}
}
