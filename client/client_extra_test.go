package client

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/remdev/go-activesync/eas"
)

// SPEC: MS-ASCMD/global.status.codes
func TestStatusError_Error(t *testing.T) {
	e := &StatusError{Command: "Sync", Status: 142}
	if msg := e.Error(); !strings.Contains(msg, "Sync") || !strings.Contains(msg, "142") {
		t.Fatalf("Error()=%q", msg)
	}
}

// SPEC: MS-ASCMD/global.status.codes
func TestNew_RequiresFields(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"no base url", Config{DeviceID: "d", DeviceType: "t"}},
		{"no device id", Config{BaseURL: "http://x", DeviceType: "t"}},
		{"no device type", Config{BaseURL: "http://x", DeviceID: "d"}},
	}
	for _, tc := range cases {
		if _, err := New(tc.cfg); err == nil {
			t.Errorf("%s: expected error", tc.name)
		}
	}
}

// SPEC: MS-ASCMD/global.status.codes
func TestGlobalStatus_AllResponses(t *testing.T) {
	cases := []any{
		&eas.FolderSyncResponse{Status: 1},
		&eas.ProvisionResponse{Status: 2},
		&eas.PingResponse{Status: 3},
		&eas.SyncResponse{Status: 4},
	}
	for _, c := range cases {
		if _, ok := globalStatus(c); !ok {
			t.Errorf("globalStatus failed for %T", c)
		}
	}
	if _, ok := globalStatus("nope"); ok {
		t.Errorf("globalStatus should reject unknown type")
	}
}

// SPEC: MS-ASCMD/global.status.codes
func TestPolicyKey_StoreError(t *testing.T) {
	c := &Client{PolicyStore: errStore{}}
	if _, err := c.policyKey(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPolicyKey_NilStore(t *testing.T) {
	c := &Client{}
	k, err := c.policyKey(context.Background())
	if err != nil || k != "0" {
		t.Fatalf("policyKey nil store: %q, %v", k, err)
	}
}

type errStore struct{}

func (errStore) Get(context.Context) (string, error) { return "", errors.New("boom") }
func (errStore) Set(context.Context, string) error   { return nil }
