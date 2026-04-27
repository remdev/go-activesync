package wbxml

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_TagBeforeSwitchPage_OK(t *testing.T) {
	// Default page is 0 (AirSync) and is initialised, so a tag byte at the
	// start of a token stream is valid. We verify by feeding only a tag byte.
	d := NewDecoder(bytes.NewReader([]byte{0x05}))
	tok, err := d.NextToken()
	if err != nil {
		t.Fatalf("NextToken: %v", err)
	}
	if tok.Kind != KindTag || tok.Page != 0 || tok.Tag != 0x05 {
		t.Fatalf("Token = %+v", tok)
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_SwitchPageUnknown(t *testing.T) {
	d := NewDecoder(bytes.NewReader([]byte{SwitchPage, 0xFE, 0x05}))
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected unknown code page error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_SwitchPageTruncated(t *testing.T) {
	d := NewDecoder(bytes.NewReader([]byte{SwitchPage}))
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected SWITCH_PAGE truncation error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_StrTUnknown(t *testing.T) {
	// Header with empty string table, then STR_T offset 5 → out of range.
	var buf bytes.Buffer
	(&Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}).Write(&buf)
	buf.WriteByte(StrT)
	buf.Write(AppendMbUint32(nil, 5))

	d := NewDecoder(&buf)
	if _, err := d.ReadHeader(); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected STR_T offset error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_StrTTruncatedOffset(t *testing.T) {
	var buf bytes.Buffer
	(&Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}).Write(&buf)
	buf.WriteByte(StrT) // missing offset bytes

	d := NewDecoder(&buf)
	if _, err := d.ReadHeader(); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected truncated STR_T error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_OpaqueTruncatedHeader(t *testing.T) {
	d := NewDecoder(bytes.NewReader([]byte{Opaque}))
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected OPAQUE length error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_OpaqueTruncatedPayload(t *testing.T) {
	// Length = 5 but only 2 bytes available.
	d := NewDecoder(bytes.NewReader([]byte{Opaque, 0x05, 0xAA, 0xBB}))
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected OPAQUE payload truncation")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_StrIUnterminated(t *testing.T) {
	// STR_I with no trailing NUL.
	d := NewDecoder(bytes.NewReader([]byte{StrI, 'a', 'b'}))
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected unterminated STR_I error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecoder_StringTableLookup(t *testing.T) {
	// Header with table "hi\0bye\0".
	var buf bytes.Buffer
	(&Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A, StringTable: []byte("hi\x00bye\x00")}).Write(&buf)
	buf.WriteByte(StrT)
	buf.Write(AppendMbUint32(nil, 3)) // offset 3 = "bye"

	d := NewDecoder(&buf)
	if _, err := d.ReadHeader(); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	tok, err := d.NextToken()
	if err != nil {
		t.Fatalf("NextToken: %v", err)
	}
	if tok.Kind != KindString || tok.String != "bye" {
		t.Fatalf("Token = %+v", tok)
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
func TestEncoder_StartTagUnknownPage(t *testing.T) {
	enc := NewEncoder(&bytes.Buffer{})
	if err := enc.StartTag(0xFF, 0x05, false, true); err == nil {
		t.Fatal("expected unknown page error")
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
func TestEncoder_WriteHeader(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty header")
	}
}

// SPEC: OMA-WBXML-1.3/header.version
func TestHeader_StringTable_RoundTrip(t *testing.T) {
	src := Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A, StringTable: []byte("a\x00b\x00")}
	var buf bytes.Buffer
	if err := src.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	var got Header
	if err := got.Read(&buf); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !bytes.Equal(got.StringTable, src.StringTable) {
		t.Fatalf("StringTable=%v want %v", got.StringTable, src.StringTable)
	}
}

// SPEC: OMA-WBXML-1.3/header.version
func TestHeader_PublicIDFromTable(t *testing.T) {
	// Encoded form: version, 0x00, mbu(7), charset, table-len 0.
	raw := []byte{0x03, 0x00, 0x07, 0x6A, 0x00}
	var h Header
	if err := h.Read(bytes.NewReader(raw)); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if h.PublicID != 7 {
		t.Fatalf("PublicID = %d", h.PublicID)
	}
}

// SPEC: OMA-WBXML-1.3/header.version
func TestHeader_ReadTruncated(t *testing.T) {
	cases := [][]byte{
		{},                // no version
		{0x03},             // missing public id
		{0x03, 0x00},       // missing public-id offset after 0
		{0x03, 0x01},       // missing charset
		{0x03, 0x01, 0x6A}, // missing string-table length
		{0x03, 0x01, 0x6A, 0x05, 'a', 'b'}, // truncated string table
	}
	for i, raw := range cases {
		var h Header
		if err := h.Read(bytes.NewReader(raw)); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

// SPEC: OMA-WBXML-1.3/mb_u_int32.encoding
func TestReadMbUint32_Errors(t *testing.T) {
	// 6 bytes with high bit set → too long.
	br := bufio.NewReader(bytes.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}))
	if _, _, err := ReadMbUint32(br); err == nil {
		t.Fatal("expected too-long error")
	}
	// EOF before any byte.
	br2 := bufio.NewReader(bytes.NewReader(nil))
	if _, _, err := ReadMbUint32(br2); err == nil {
		t.Fatal("expected EOF error")
	}
	// EOF mid-sequence.
	br3 := bufio.NewReader(bytes.NewReader([]byte{0x80}))
	if _, _, err := ReadMbUint32(br3); err == nil {
		t.Fatal("expected unexpected EOF error")
	}
}

// SPEC: OMA-WBXML-1.3/mb_u_int32.encoding
func TestWriteMbUint32(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteMbUint32(&buf, 300); err != nil {
		t.Fatalf("WriteMbUint32: %v", err)
	}
	got, n, err := ReadMbUint32(bufio.NewReader(&buf))
	if err != nil || got != 300 {
		t.Fatalf("round-trip got %d (n=%d, err=%v)", got, n, err)
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestEncodeTag_PanicsOnLargeIdentity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = EncodeTag(0x40, false, false)
}

// SPEC: MS-ASWBXML/decoder.tag
func TestByteReader_PassThrough(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("hi"))
	if got := byteReader(br); got != br {
		t.Fatal("byteReader did not pass through existing ByteReader")
	}
	got := byteReader(strings.NewReader("hi"))
	if _, ok := any(got).(io.ByteReader); !ok {
		t.Fatal("byteReader did not adapt plain Reader")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_NonStruct(t *testing.T) {
	x := 42
	if _, err := Marshal(&x); err == nil {
		t.Fatal("expected error for non-struct pointer target")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestUnmarshal_HeaderEOF(t *testing.T) {
	if err := Unmarshal(nil, &allKinds{}); err == nil {
		t.Fatal("expected header read error")
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestUnmarshal_RootMismatch(t *testing.T) {
	// Build a doc whose root is AirSync.SyncKey rather than AirSync.Sync.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}); err != nil {
		t.Fatalf("header: %v", err)
	}
	if err := enc.StartTag(0, 0x0B /* SyncKey */, false, false); err != nil {
		t.Fatalf("StartTag: %v", err)
	}
	if err := Unmarshal(buf.Bytes(), &allKinds{}); err == nil {
		t.Fatal("expected root-tag mismatch")
	}
}

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
func TestAssignScalar_Errors(t *testing.T) {
	type intHolder struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		N       int32    `wbxml:"AirSync.WindowSize"`
	}
	type uintHolder struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		N       uint32   `wbxml:"AirSync.WindowSize"`
	}
	type boolHolder struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		B       bool     `wbxml:"AirSync.GetChanges"`
	}
	build := func(tagID byte, value string) []byte {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
		enc.StartTag(0, 0x05, false, true)
		enc.StartTag(0, tagID, false, true)
		enc.StrI(value)
		enc.EndTag()
		enc.EndTag()
		return buf.Bytes()
	}
	if err := Unmarshal(build(0x15 /* WindowSize */, "abc"), &intHolder{}); err == nil {
		t.Errorf("int parse error expected")
	}
	if err := Unmarshal(build(0x15, "-1"), &uintHolder{}); err == nil {
		t.Errorf("uint parse error expected")
	}
	if err := Unmarshal(build(0x13 /* GetChanges */, "maybe"), &boolHolder{}); err == nil {
		t.Errorf("bool parse error expected")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestParseTag_AllErrors(t *testing.T) {
	cases := []string{"NoDot", "AirSync.NoSuchTag", "BadPage.Sync", "AirSync.Sync,bogus"}
	for _, c := range cases {
		if _, err := parseTag(c); err == nil {
			t.Errorf("parseTag(%q): expected error", c)
		}
	}
}

// Marshal-time error from a sub-struct with an unknown tag option.
type badInner struct {
	X string `wbxml:"AirSync.SyncKey,wat"`
}
type badOuter struct {
	XMLName struct{}  `wbxml:"AirSync.Sync"`
	Inner   *badInner `wbxml:"AirSync.Collection"`
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_NestedInfoForError(t *testing.T) {
	in := badOuter{Inner: &badInner{X: "y"}}
	if _, err := Marshal(&in); err == nil {
		t.Fatal("expected nested infoFor error")
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

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestSkipElement_TruncatedReturnsErr(t *testing.T) {
	// Missing END after the opening tag triggers skipElement's read error path.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05, false, true)
	enc.StartTag(0, 0x10 /* Class */, false, true)
	// no end markers
	if err := Unmarshal(buf.Bytes(), &wrappedTag{}); err == nil {
		t.Fatal("expected truncation error")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestReadElementValue_NestedTagError(t *testing.T) {
	// Build a doc where SyncKey contains a nested tag (not allowed for scalar).
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05, false, true)
	enc.StartTag(0, 0x0B /* SyncKey */, false, true)
	enc.StartTag(0, 0x15 /* WindowSize */, false, false)
	enc.EndTag()
	enc.EndTag()

	if err := Unmarshal(buf.Bytes(), &wrappedTag{}); err == nil {
		t.Fatal("expected nested-tag-in-scalar error")
	}
}

// errReader fails the very first ReadByte after a header is parsed, exposing
// EOF propagation paths in NextToken.
type errReader struct{ remaining []byte }

func (r *errReader) Read(p []byte) (int, error) {
	if len(r.remaining) == 0 {
		return 0, errors.New("forced read error")
	}
	n := copy(p, r.remaining)
	r.remaining = r.remaining[n:]
	return n, nil
}

// SPEC: MS-ASWBXML/decoder.tag
func TestNextToken_ReadByteError(t *testing.T) {
	d := NewDecoder(&errReader{})
	if _, err := d.NextToken(); err == nil {
		t.Fatal("expected ReadByte error")
	}
}
