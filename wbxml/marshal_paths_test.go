package wbxml

import (
	"bytes"
	"reflect"
	"testing"
)

type allKinds struct {
	XMLName  struct{} `wbxml:"AirSync.Sync"`
	Str      string   `wbxml:"AirSync.SyncKey"`
	Bln      bool     `wbxml:"AirSync.GetChanges"`
	Int32V   int32    `wbxml:"AirSync.WindowSize"`
	Uint32V  uint32   `wbxml:"AirSync.MIMETruncation"`
	BytesRaw []byte   `wbxml:"AirSync.Class"`
	BytesOp  []byte   `wbxml:"AirSync.ApplicationData,opaque"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestMarshal_AllScalarKinds(t *testing.T) {
	in := allKinds{
		Str:      "key",
		Bln:      true,
		Int32V:   100,
		Uint32V:  200,
		BytesRaw: []byte("text"),
		BytesOp:  []byte{0x00, 0xFF, 0x10},
	}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out allKinds
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	out.XMLName = in.XMLName
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

type omitMe struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Str     string   `wbxml:"AirSync.SyncKey,omitempty"`
	Bln     bool     `wbxml:"AirSync.GetChanges,omitempty"`
	Int32V  int32    `wbxml:"AirSync.WindowSize,omitempty"`
	Uint32V uint32   `wbxml:"AirSync.MIMETruncation,omitempty"`
}

// SPEC: MS-ASWBXML/marshal.omitempty
func TestMarshal_OmitemptyAllZero(t *testing.T) {
	in := omitMe{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// header(4) + opening empty tag (no content) for AirSync.Sync.
	if !bytes.Contains(data, []byte{0x05}) {
		t.Fatalf("expected Sync tag in output, got %x", data)
	}
}

type nilPtr struct {
	XMLName struct{}    `wbxml:"AirSync.Sync"`
	Inner   *innerEmpty `wbxml:"AirSync.Collection"`
}

type innerEmpty struct {
	Key string `wbxml:"AirSync.SyncKey,omitempty"`
}

// SPEC: MS-ASWBXML/marshal.omitempty
func TestMarshal_NilPointerWithoutOmitempty(t *testing.T) {
	in := nilPtr{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out nilPtr
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

type sliceStructPtr struct {
	XMLName struct{}     `wbxml:"AirSync.Sync"`
	Items   []*innerItem `wbxml:"AirSync.Collection"`
}

type innerItem struct {
	Key string `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.slice
func TestMarshal_SliceOfPointerStruct(t *testing.T) {
	in := sliceStructPtr{
		Items: []*innerItem{{Key: "a"}, {Key: "b"}},
	}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceStructPtr
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Items) != 2 || out.Items[0].Key != "a" || out.Items[1].Key != "b" {
		t.Fatalf("Items = %+v", out.Items)
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestParseTag_RejectsBadFormats(t *testing.T) {
	if _, err := parseTag("AirSyncSync"); err == nil {
		t.Errorf("expected error for missing dot")
	}
	if _, err := parseTag("AirSync.NotATag"); err == nil {
		t.Errorf("expected error for unknown tag")
	}
	if _, err := parseTag("Bogus.Sync"); err == nil {
		t.Errorf("expected error for unknown page")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestUnmarshal_RejectsTruncated(t *testing.T) {
	if err := Unmarshal([]byte{}, &allKinds{}); err == nil {
		t.Fatalf("expected error on empty input")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestUnmarshal_RejectsNonPointer(t *testing.T) {
	if err := Unmarshal([]byte{0x03, 0x01, 0x6A, 0x00}, allKinds{}); err == nil {
		t.Fatalf("expected error on non-pointer")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_RejectsNonPointer(t *testing.T) {
	if _, err := Marshal(123); err == nil {
		t.Fatalf("expected error on non-struct")
	}
}

type withSkippable struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Key     string   `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestUnmarshal_SkipNestedUnknown(t *testing.T) {
	// Build a document with a nested unknown element to exercise skipElement.
	src := struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		Key     string   `wbxml:"AirSync.SyncKey"`
		Extra   *withInner `wbxml:"AirSync.Class"`
	}{Key: "k", Extra: &withInner{Inner: &withInnerInner{Name: "x"}}}
	data, err := Marshal(&src)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out withSkippable
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Key != "k" {
		t.Fatalf("Key = %q", out.Key)
	}
}

type withInner struct {
	Inner *withInnerInner `wbxml:"AirSync.Collection"`
}

type withInnerInner struct {
	Name string `wbxml:"AirSync.SyncKey"`
}
