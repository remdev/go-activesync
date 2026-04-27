package wbxml

import (
	"bytes"
	"reflect"
	"testing"
)

type scalarSliceMsg struct {
	XMLName  struct{} `wbxml:"AirSync.Sync"`
	Children []string `wbxml:"AirSync.SyncKey"`
	Counts   []int32  `wbxml:"AirSync.WindowSize"`
	Flags    []bool   `wbxml:"AirSync.GetChanges"`
}

// SPEC: MS-ASWBXML/marshal.slice
func TestMarshal_SliceOfScalar(t *testing.T) {
	in := scalarSliceMsg{
		Children: []string{"a", "b", "c"},
		Counts:   []int32{1, 2},
		Flags:    []bool{true, false},
	}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out scalarSliceMsg
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	out.XMLName = in.XMLName
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

type uintMsg struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Window  uint32   `wbxml:"AirSync.WindowSize"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestMarshal_UintScalar(t *testing.T) {
	in := uintMsg{Window: 42}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out uintMsg
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	out.XMLName = in.XMLName
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

type ptrChildMsg struct {
	XMLName struct{}  `wbxml:"AirSync.Sync"`
	Inner   *innerMsg `wbxml:"AirSync.Collection,omitempty"`
}

type innerMsg struct {
	Name string `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestMarshal_PointerStruct(t *testing.T) {
	in := ptrChildMsg{Inner: &innerMsg{Name: "abc"}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ptrChildMsg
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner == nil || out.Inner.Name != "abc" {
		t.Fatalf("Inner = %+v", out.Inner)
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_RejectsUnknownPage(t *testing.T) {
	type bad struct {
		XMLName struct{} `wbxml:"NotAPage.Sync"`
	}
	if _, err := Marshal(&bad{}); err == nil {
		t.Fatalf("Marshal: expected error for unknown page")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_RejectsBadTagOption(t *testing.T) {
	type bad struct {
		XMLName struct{} `wbxml:"AirSync.Sync,wrongoption"`
	}
	if _, err := Marshal(&bad{}); err == nil {
		t.Fatalf("Marshal: expected error for bad option")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_RejectsMissingXMLName(t *testing.T) {
	type bad struct {
		Field string `wbxml:"AirSync.SyncKey"`
	}
	if _, err := Marshal(&bad{}); err == nil {
		t.Fatalf("Marshal: expected error for missing XMLName")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestUnmarshal_RejectsRootMismatch(t *testing.T) {
	type a struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
	}
	type b struct {
		XMLName struct{} `wbxml:"AirSync.Collection"`
	}
	data, err := Marshal(&a{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := Unmarshal(data, &b{}); err == nil {
		t.Fatalf("Unmarshal: expected mismatch error")
	}
}

type simpleSync struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	SyncKey string   `wbxml:"AirSync.SyncKey,omitempty"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestUnmarshal_IgnoresUnknownChild(t *testing.T) {
	in := struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		Key     string   `wbxml:"AirSync.SyncKey"`
		Extra   string   `wbxml:"AirSync.Class"`
	}{Key: "k", Extra: "e"}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out simpleSync
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.SyncKey != "k" {
		t.Fatalf("SyncKey = %q", out.SyncKey)
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestTokenKind_String(t *testing.T) {
	cases := []struct {
		k    TokenKind
		want string
	}{
		{KindTag, "Tag"}, {KindEnd, "End"}, {KindString, "String"},
		{KindOpaque, "Opaque"}, {TokenKind(99), "TokenKind(99)"},
	}
	for _, c := range cases {
		if got := c.k.String(); got != c.want {
			t.Errorf("(%d).String() = %q, want %q", c.k, got, c.want)
		}
	}
}

// SPEC: MS-ASWBXML/decoder.string
func TestDecoder_StringTable(t *testing.T) {
	// Header with string-table populated, body referring offset 0 via STR_T.
	doc := []byte{
		0x03, 0x01, 0x6A, // version=1.3, publicid=1, charset=UTF-8 (mb_u_int32 0x6A)
		0x05,          // string-table length = 5
		'h', 'i', 0,
		'!', 0,
		// Tag: AirSync.Sync (page 0, tag 0x05) with content
		0x45,
		// STR_T at offset 0
		0x83, 0x00,
		// END
		0x01,
	}
	dec := NewDecoder(bytes.NewReader(doc))
	if _, err := dec.ReadHeader(); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	tok, err := dec.NextToken()
	if err != nil {
		t.Fatalf("first NextToken: %v", err)
	}
	if tok.Kind != KindTag {
		t.Fatalf("first kind = %s", tok.Kind)
	}
	str, err := dec.NextToken()
	if err != nil {
		t.Fatalf("string token: %v", err)
	}
	if str.Kind != KindString || str.String != "hi" {
		t.Fatalf("string token = %+v", str)
	}
	end, err := dec.NextToken()
	if err != nil {
		t.Fatalf("end token: %v", err)
	}
	if end.Kind != KindEnd {
		t.Fatalf("end token = %+v", end)
	}
}
