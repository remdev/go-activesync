package wbxml

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// TokenKind classifies a logical WBXML token after SWITCH_PAGE handling.
type TokenKind int

const (
	KindTag TokenKind = iota + 1
	KindEnd
	KindString
	KindOpaque
)

func (k TokenKind) String() string {
	switch k {
	case KindTag:
		return "Tag"
	case KindEnd:
		return "End"
	case KindString:
		return "String"
	case KindOpaque:
		return "Opaque"
	default:
		return fmt.Sprintf("TokenKind(%d)", int(k))
	}
}

// Token is a decoded WBXML logical event.
type Token struct {
	Kind       TokenKind
	Page       byte   // valid for KindTag
	Tag        byte   // 6-bit identity, valid for KindTag
	HasAttrs   bool   // valid for KindTag
	HasContent bool   // valid for KindTag
	String     string // valid for KindString
	Bytes      []byte // valid for KindOpaque
}

// Decoder is a streaming WBXML 1.3 decoder. It transparently consumes
// SWITCH_PAGE tokens and reflects the active page on subsequent tag tokens.
type Decoder struct {
	r           *bufio.Reader
	page        byte
	pageInit    bool
	stringTable []byte
}

// NewDecoder returns a new decoder reading from r. The active code page is
// initialised to 0 per OMA-WBXML 1.3 §5.5.
func NewDecoder(r io.Reader) *Decoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Decoder{r: br, page: 0, pageInit: true}
}

// ReadHeader parses the document header.
func (d *Decoder) ReadHeader() (Header, error) {
	var h Header
	if err := h.Read(d.r); err != nil {
		return h, err
	}
	d.stringTable = h.StringTable
	return h, nil
}

// NextToken reads and returns the next logical token. SWITCH_PAGE is consumed
// internally and is not surfaced as a token.
func (d *Decoder) NextToken() (Token, error) {
	for {
		b, err := d.r.ReadByte()
		if err != nil {
			return Token{}, err
		}
		switch b {
		case SwitchPage:
			page, err := d.r.ReadByte()
			if err != nil {
				return Token{}, fmt.Errorf("wbxml: SWITCH_PAGE: %w", err)
			}
			if _, ok := PageByID(page); !ok {
				return Token{}, fmt.Errorf("wbxml: SWITCH_PAGE to unknown code page %d", page)
			}
			d.page = page
			d.pageInit = true
		case End:
			return Token{Kind: KindEnd}, nil
		case StrI:
			s, err := readNulString(d.r)
			if err != nil {
				return Token{}, err
			}
			return Token{Kind: KindString, String: s}, nil
		case StrT:
			off, _, err := ReadMbUint32(d.r)
			if err != nil {
				return Token{}, err
			}
			s, err := stringFromTable(d.stringTable, off)
			if err != nil {
				return Token{}, err
			}
			return Token{Kind: KindString, String: s}, nil
		case Opaque:
			n, _, err := ReadMbUint32(d.r)
			if err != nil {
				return Token{}, err
			}
			payload := make([]byte, n)
			if _, err := io.ReadFull(d.r, payload); err != nil {
				return Token{}, err
			}
			return Token{Kind: KindOpaque, Bytes: payload}, nil
		default:
			// All remaining bytes are tag bytes; their identity must fall in
			// the 6-bit range and the active page must already be known.
			if !d.pageInit {
				return Token{}, errors.New("wbxml: tag byte before any SWITCH_PAGE")
			}
			return Token{
				Kind:       KindTag,
				Page:       d.page,
				Tag:        TagIdentity(b),
				HasAttrs:   TagHasAttributes(b),
				HasContent: TagHasContent(b),
			}, nil
		}
	}
}

func readNulString(r io.ByteReader) (string, error) {
	var buf []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", io.ErrUnexpectedEOF
			}
			return "", err
		}
		if b == 0x00 {
			return string(buf), nil
		}
		buf = append(buf, b)
	}
}

func stringFromTable(tbl []byte, off uint32) (string, error) {
	if int(off) >= len(tbl) {
		return "", fmt.Errorf("wbxml: STR_T offset %d out of range (table=%d)", off, len(tbl))
	}
	end := int(off)
	for end < len(tbl) && tbl[end] != 0x00 {
		end++
	}
	return string(tbl[off:end]), nil
}
