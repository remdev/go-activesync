package wbxml

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// SPEC: OMA-WBXML-1.3/global.SWITCH_PAGE
// SPEC: OMA-WBXML-1.3/global.END
// SPEC: OMA-WBXML-1.3/global.STR_I
// SPEC: OMA-WBXML-1.3/global.STR_T
// SPEC: OMA-WBXML-1.3/global.OPAQUE
func TestGlobalTokens_Constants(t *testing.T) {
	cases := []struct {
		name string
		got  byte
		want byte
	}{
		{"SwitchPage", SwitchPage, 0x00},
		{"End", End, 0x01},
		{"Entity", Entity, 0x02},
		{"StrI", StrI, 0x03},
		{"Literal", Literal, 0x04},
		{"ExtI0", ExtI0, 0x40},
		{"ExtI1", ExtI1, 0x41},
		{"ExtI2", ExtI2, 0x42},
		{"PI", PI, 0x43},
		{"LiteralC", LiteralC, 0x44},
		{"ExtT0", ExtT0, 0x80},
		{"ExtT1", ExtT1, 0x81},
		{"ExtT2", ExtT2, 0x82},
		{"StrT", StrT, 0x83},
		{"LiteralA", LiteralA, 0x84},
		{"Ext0", Ext0, 0xC0},
		{"Ext1", Ext1, 0xC1},
		{"Ext2", Ext2, 0xC2},
		{"Opaque", Opaque, 0xC3},
		{"LiteralAC", LiteralAC, 0xC4},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = 0x%02X, want 0x%02X", c.name, c.got, c.want)
		}
	}
}

// SPEC: OMA-WBXML-1.3/tag.bits
func TestTagBits(t *testing.T) {
	cases := []struct {
		name           string
		b              byte
		hasAttrs       bool
		hasContent     bool
		identity       byte
	}{
		{"plain", 0x05, false, false, 0x05},
		{"content", 0x45, false, true, 0x05},
		{"attrs", 0x85, true, false, 0x05},
		{"both", 0xC5, true, true, 0x05},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := TagHasAttributes(c.b); got != c.hasAttrs {
				t.Errorf("TagHasAttributes(0x%02X) = %v, want %v", c.b, got, c.hasAttrs)
			}
			if got := TagHasContent(c.b); got != c.hasContent {
				t.Errorf("TagHasContent(0x%02X) = %v, want %v", c.b, got, c.hasContent)
			}
			if got := TagIdentity(c.b); got != c.identity {
				t.Errorf("TagIdentity(0x%02X) = 0x%02X, want 0x%02X", c.b, got, c.identity)
			}
			if got := EncodeTag(c.identity, c.hasAttrs, c.hasContent); got != c.b {
				t.Errorf("EncodeTag(0x%02X, %v, %v) = 0x%02X, want 0x%02X", c.identity, c.hasAttrs, c.hasContent, got, c.b)
			}
		})
	}
}

// SPEC: OMA-WBXML-1.3/mb_u_int32.encoding
// SPEC: OMA-WBXML-1.3/mb_u_int32.boundaries
func TestMbUint32_RoundTrip(t *testing.T) {
	cases := []struct {
		v       uint32
		encoded []byte
	}{
		{0x00000000, []byte{0x00}},
		{0x0000007F, []byte{0x7F}},
		{0x00000080, []byte{0x81, 0x00}},
		{0x000000A4, []byte{0x81, 0x24}},
		{0x00003FFF, []byte{0xFF, 0x7F}},
		{0x00004000, []byte{0x81, 0x80, 0x00}},
		{0x001FFFFF, []byte{0xFF, 0xFF, 0x7F}},
		{0x00200000, []byte{0x81, 0x80, 0x80, 0x00}},
		{0x0FFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0x7F}},
		{0x10000000, []byte{0x81, 0x80, 0x80, 0x80, 0x00}},
		{0xFFFFFFFF, []byte{0x8F, 0xFF, 0xFF, 0xFF, 0x7F}},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := WriteMbUint32(&buf, c.v); err != nil {
			t.Fatalf("WriteMbUint32(%#x): %v", c.v, err)
		}
		if !bytes.Equal(buf.Bytes(), c.encoded) {
			t.Errorf("WriteMbUint32(%#x) = % X, want % X", c.v, buf.Bytes(), c.encoded)
		}
		got, n, err := ReadMbUint32(bytes.NewReader(c.encoded))
		if err != nil {
			t.Fatalf("ReadMbUint32(% X): %v", c.encoded, err)
		}
		if got != c.v {
			t.Errorf("ReadMbUint32(% X) = %#x, want %#x", c.encoded, got, c.v)
		}
		if n != len(c.encoded) {
			t.Errorf("ReadMbUint32(% X) consumed %d bytes, want %d", c.encoded, n, len(c.encoded))
		}
	}
}

// SPEC: OMA-WBXML-1.3/mb_u_int32.encoding
func TestMbUint32_Truncated(t *testing.T) {
	if _, _, err := ReadMbUint32(bytes.NewReader([]byte{0x81})); !errors.Is(err, io.ErrUnexpectedEOF) && err == nil {
		t.Fatalf("expected error on truncated mb_u_int32")
	}
}

// SPEC: OMA-WBXML-1.3/header.version
// SPEC: OMA-WBXML-1.3/header.publicid
// SPEC: OMA-WBXML-1.3/header.charset
// SPEC: OMA-WBXML-1.3/header.stringtable
func TestHeader_RoundTripDefault(t *testing.T) {
	h := Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}
	var buf bytes.Buffer
	if err := h.Write(&buf); err != nil {
		t.Fatalf("Header.Write: %v", err)
	}
	want := []byte{0x03, 0x01, 0x6A, 0x00}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("Header bytes = % X, want % X", buf.Bytes(), want)
	}
	var got Header
	if err := got.Read(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("Header.Read: %v", err)
	}
	if got.Version != 0x03 || got.PublicID != 0x01 || got.Charset != 0x6A || len(got.StringTable) != 0 {
		t.Fatalf("Header.Read got %+v, want default", got)
	}
}

// SPEC: OMA-WBXML-1.3/header.stringtable
func TestHeader_StringTable(t *testing.T) {
	h := Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A, StringTable: []byte("hello\x00world\x00")}
	var buf bytes.Buffer
	if err := h.Write(&buf); err != nil {
		t.Fatalf("Header.Write: %v", err)
	}
	want := append([]byte{0x03, 0x01, 0x6A, byte(len(h.StringTable))}, h.StringTable...)
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("Header bytes = % X, want % X", buf.Bytes(), want)
	}
	var got Header
	if err := got.Read(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("Header.Read: %v", err)
	}
	if !bytes.Equal(got.StringTable, h.StringTable) {
		t.Fatalf("Header.Read StringTable = % X, want % X", got.StringTable, h.StringTable)
	}
}
