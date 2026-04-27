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

// SPEC: MS-ASWBXML/marshal.slice
type slicePtrNil struct {
	XMLName struct{}       `wbxml:"AirSync.Sync"`
	Items   []*innerScalar `wbxml:"AirSync.Collection"`
}

type innerScalar struct {
	K string `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.slice
func TestMarshal_SlicePtrWithNilEntries(t *testing.T) {
	in := slicePtrNil{Items: []*innerScalar{nil, {K: "x"}}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out slicePtrNil
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].K != "x" {
		t.Fatalf("Items = %+v", out.Items)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type allOmit struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Bytes   []byte   `wbxml:"AirSync.Class,omitempty"`
}

func TestMarshal_OmitEmptyBytes(t *testing.T) {
	in := allOmit{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got allOmit
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Fatalf("round-trip mismatch")
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceStructInner struct {
	K string `wbxml:"AirSync.SyncKey"`
}
type sliceStructDirect struct {
	XMLName struct{}           `wbxml:"AirSync.Sync"`
	Items   []sliceStructInner `wbxml:"AirSync.Collection"`
}

func TestMarshal_SliceOfStructDirect(t *testing.T) {
	in := sliceStructDirect{Items: []sliceStructInner{{K: "a"}, {K: "b"}}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceStructDirect
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Items) != 2 || out.Items[0].K != "a" || out.Items[1].K != "b" {
		t.Fatalf("Items = %+v", out.Items)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type ptrStructHolder struct {
	XMLName struct{}          `wbxml:"AirSync.Sync"`
	Inner   *sliceStructInner `wbxml:"AirSync.Collection"`
}

func TestMarshal_PointerStructFilled(t *testing.T) {
	in := ptrStructHolder{Inner: &sliceStructInner{K: "p"}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ptrStructHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner == nil || out.Inner.K != "p" {
		t.Fatalf("Inner = %+v", out.Inner)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type byteSlicePlain struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Class   []byte   `wbxml:"AirSync.Class"`
}

func TestMarshal_ByteSliceNonOpaque(t *testing.T) {
	in := byteSlicePlain{Class: []byte("Email")}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out byteSlicePlain
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(out.Class) != "Email" {
		t.Fatalf("Class = %q", out.Class)
	}
}

// SPEC: MS-ASWBXML/marshal.omitempty
type ptrOmit struct {
	XMLName struct{}     `wbxml:"AirSync.Sync"`
	Inner   *innerScalar `wbxml:"AirSync.Collection,omitempty"`
}

func TestMarshal_NilPointerWithOmitempty(t *testing.T) {
	in := ptrOmit{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ptrOmit
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner != nil {
		t.Fatalf("Inner should be nil")
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceBoolHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Bs      []bool   `wbxml:"AirSync.GetChanges"`
}

func TestMarshal_SliceOfBoolTrue(t *testing.T) {
	in := sliceBoolHolder{Bs: []bool{true, false, true}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceBoolHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in.Bs, out.Bs) {
		t.Fatalf("Bs = %v, want %v", out.Bs, in.Bs)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type byteSliceHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Class   []byte   `wbxml:"AirSync.Class,opaque"`
}

func TestUnmarshal_OpaqueByteSliceWithoutContent(t *testing.T) {
	// Build doc: Sync (content) → Class (no content) → End.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x10 /* Class */, false, false)
	enc.EndTag()
	var out byteSliceHolder
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Class != nil {
		t.Fatalf("Class = %v", out.Class)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type sliceIntHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Vals    []int32  `wbxml:"AirSync.WindowSize"`
}

func TestUnmarshal_SliceOfScalarWithoutContent(t *testing.T) {
	// Build doc: Sync (content) → WindowSize (no content) → end.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x15 /* WindowSize */, false, false)
	enc.EndTag()

	var out sliceIntHolder
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Vals) != 1 {
		t.Fatalf("Vals = %v", out.Vals)
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceUintHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Vals    []uint32 `wbxml:"AirSync.MIMETruncation"`
}

func TestMarshal_SliceOfUint(t *testing.T) {
	in := sliceUintHolder{Vals: []uint32{1, 2, 3}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceUintHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in.Vals, out.Vals) {
		t.Fatalf("Vals = %v", out.Vals)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestMarshal_OmitEmptyZeroScalars(t *testing.T) {
	type allZero struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		S       string   `wbxml:"AirSync.SyncKey,omitempty"`
		B       bool     `wbxml:"AirSync.GetChanges,omitempty"`
		I       int32    `wbxml:"AirSync.WindowSize,omitempty"`
		U       uint32   `wbxml:"AirSync.MIMETruncation,omitempty"`
	}
	in := allZero{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out allZero
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type wrappedTag struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Inner   string   `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestUnmarshal_ScalarWithoutContent(t *testing.T) {
	// Sync (with content) → SyncKey (no content) → end.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x0B /* SyncKey */, false, false)
	enc.EndTag()

	var out wrappedTag
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner != "" {
		t.Fatalf("Inner = %q", out.Inner)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestSkipElement_Nested(t *testing.T) {
	// Build a doc with a tag we know but inside it a deeply nested unknown
	// child tree, ensuring decodeStruct calls skipElement which recurses.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x0B /* SyncKey */, false, true)
	enc.StrI("k")
	enc.EndTag()
	// Unknown nested subtree (Class) with content and grand-children.
	enc.StartTag(0, 0x10 /* Class */, false, true)
	enc.StartTag(0, 0x15 /* WindowSize */, false, true)
	enc.StrI("1")
	enc.EndTag()
	enc.EndTag()
	enc.EndTag()

	var out wrappedTag
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner != "k" {
		t.Fatalf("Inner = %q", out.Inner)
	}
}

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

type withSkippable struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Key     string   `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestUnmarshal_SkipNestedUnknown(t *testing.T) {
	// Build a document with a nested unknown element to exercise skipElement.
	src := struct {
		XMLName struct{}   `wbxml:"AirSync.Sync"`
		Key     string     `wbxml:"AirSync.SyncKey"`
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

// SPEC: MS-ASWBXML/marshal.omitempty
func TestIsZero_MapAndPointer(t *testing.T) {
	m := map[string]int(nil)
	if !isZero(reflect.ValueOf(m)) {
		t.Fatal("nil map not zero")
	}
	mp := map[string]int{"a": 1}
	if isZero(reflect.ValueOf(mp)) {
		t.Fatal("populated map zero")
	}
	var ptr *int
	if !isZero(reflect.ValueOf(ptr)) {
		t.Fatal("nil pointer not zero")
	}
}
