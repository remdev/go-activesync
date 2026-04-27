// Package wbxml implements the WAP Binary XML 1.3 codec used by the Microsoft
// Exchange ActiveSync wire protocol.
//
// The codec is split across a few files:
//
//   - tokens.go: global tokens, mb_u_int32, header, tag-byte bit semantics.
//   - codepages.go / codepages_eas.go: EAS code page tables (filled in by the
//     codepages stage).
//   - encoder.go / decoder.go: streaming encoder and decoder.
//   - marshal.go: reflection-based Go ↔ WBXML marshalling.
package wbxml

import (
	"errors"
	"fmt"
	"io"
)

// Global token codes from OMA-WBXML 1.3 §5.5.
const (
	SwitchPage byte = 0x00
	End        byte = 0x01
	Entity     byte = 0x02
	StrI       byte = 0x03
	Literal    byte = 0x04

	ExtI0    byte = 0x40
	ExtI1    byte = 0x41
	ExtI2    byte = 0x42
	PI       byte = 0x43
	LiteralC byte = 0x44

	ExtT0    byte = 0x80
	ExtT1    byte = 0x81
	ExtT2    byte = 0x82
	StrT     byte = 0x83
	LiteralA byte = 0x84

	Ext0      byte = 0xC0
	Ext1      byte = 0xC1
	Ext2      byte = 0xC2
	Opaque    byte = 0xC3
	LiteralAC byte = 0xC4
)

// Tag-byte bit semantics from OMA-WBXML 1.3 §5.8.4.2.
const (
	tagAttrBit    byte = 0x80
	tagContentBit byte = 0x40
	tagIDMask     byte = 0x3F
)

// TagHasAttributes reports whether bit 7 (0x80) is set on a tag byte.
func TagHasAttributes(b byte) bool { return b&tagAttrBit != 0 }

// TagHasContent reports whether bit 6 (0x40) is set on a tag byte.
func TagHasContent(b byte) bool { return b&tagContentBit != 0 }

// TagIdentity returns the 6-bit tag identity within the active code page.
func TagIdentity(b byte) byte { return b & tagIDMask }

// EncodeTag composes a tag byte from its identity and bit flags.
//
// The identity must fit in 6 bits; values >= 0x40 are treated as a programmer
// error and panic, since the WBXML wire format simply cannot represent them
// in a single tag byte.
func EncodeTag(identity byte, hasAttributes, hasContent bool) byte {
	if identity&^tagIDMask != 0 {
		panic(fmt.Sprintf("wbxml: tag identity 0x%02X does not fit in 6 bits", identity))
	}
	out := identity
	if hasAttributes {
		out |= tagAttrBit
	}
	if hasContent {
		out |= tagContentBit
	}
	return out
}

// WriteMbUint32 writes v as a multi-byte integer per OMA-WBXML 1.3 §5.1.
//
// The encoding splits v into 7-bit groups, big-endian, with the high bit of
// every byte except the last one set. The shortest possible encoding is used.
func WriteMbUint32(w io.ByteWriter, v uint32) error {
	var buf [5]byte
	i := len(buf) - 1
	buf[i] = byte(v) & 0x7F
	v >>= 7
	for v != 0 {
		i--
		buf[i] = byte(v)&0x7F | 0x80
		v >>= 7
	}
	for ; i < len(buf); i++ {
		if err := w.WriteByte(buf[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReadMbUint32 reads a multi-byte integer from r and returns its value plus
// the number of bytes consumed.
//
// The reader is consumed strictly: at most 5 bytes are read, since values
// outside the uint32 range cannot appear in well-formed WBXML 1.3 streams.
func ReadMbUint32(r io.ByteReader) (uint32, int, error) {
	var v uint32
	for n := 1; n <= 5; n++ {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && n > 1 {
				return 0, n - 1, io.ErrUnexpectedEOF
			}
			return 0, n - 1, err
		}
		v = (v << 7) | uint32(b&0x7F)
		if b&0x80 == 0 {
			return v, n, nil
		}
	}
	return 0, 5, errors.New("wbxml: mb_u_int32 longer than 5 bytes")
}

// Header is the per-document WBXML preamble from §5.4.
//
// PublicID is encoded inline as mb_u_int32; this implementation does not
// emit the "0 followed by STR_T offset" form, but Read accepts and resolves
// it when present, exposing the resulting integer value.
type Header struct {
	Version     byte   // 0x03 for WBXML 1.3
	PublicID    uint32 // 0x01 = unknown
	Charset     uint32 // IANA MIBenum, 106 = UTF-8
	StringTable []byte
}

// Write serialises h to w.
func (h *Header) Write(w io.Writer) error {
	bw := byteWriter(w)
	if err := bw.WriteByte(h.Version); err != nil {
		return err
	}
	if err := WriteMbUint32(bw, h.PublicID); err != nil {
		return err
	}
	if err := WriteMbUint32(bw, h.Charset); err != nil {
		return err
	}
	if err := WriteMbUint32(bw, uint32(len(h.StringTable))); err != nil {
		return err
	}
	if len(h.StringTable) > 0 {
		if _, err := w.Write(h.StringTable); err != nil {
			return err
		}
	}
	return nil
}

// Read parses the header from r.
func (h *Header) Read(r io.Reader) error {
	br := byteReader(r)
	v, err := br.ReadByte()
	if err != nil {
		return err
	}
	h.Version = v

	pubID, _, err := ReadMbUint32(br)
	if err != nil {
		return err
	}
	if pubID == 0 {
		// Public ID identifier sits in the string table; its mb_u_int32 offset
		// follows immediately. We surface the resolved offset as PublicID; the
		// upper layers only care about the integer value.
		off, _, err := ReadMbUint32(br)
		if err != nil {
			return err
		}
		h.PublicID = off
	} else {
		h.PublicID = pubID
	}

	if h.Charset, _, err = ReadMbUint32(br); err != nil {
		return err
	}
	stLen, _, err := ReadMbUint32(br)
	if err != nil {
		return err
	}
	if stLen == 0 {
		h.StringTable = nil
		return nil
	}
	st := make([]byte, stLen)
	if _, err := io.ReadFull(r, st); err != nil {
		return err
	}
	h.StringTable = st
	return nil
}
