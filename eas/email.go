package eas

// Importance enum values per MS-ASEMAIL §2.2.2.40.
const (
	ImportanceLow    int32 = 0
	ImportanceNormal int32 = 1
	ImportanceHigh   int32 = 2
)

// Email is the sync representation of an e-mail message (MS-ASEMAIL 14.1).
// It marshals as the AirSync.ApplicationData wrapper used inside Sync
// commands, with Email-page child elements.
type Email struct {
	XMLName      struct{} `wbxml:"AirSync.ApplicationData"`
	DateReceived string   `wbxml:"Email.DateReceived,omitempty"`
	Subject      string   `wbxml:"Email.Subject,omitempty"`
	From         string   `wbxml:"Email.From,omitempty"`
	To           string   `wbxml:"Email.To,omitempty"`
	Cc           string   `wbxml:"Email.Cc,omitempty"`
	ReplyTo      string   `wbxml:"Email.ReplyTo,omitempty"`
	DisplayTo    string   `wbxml:"Email.DisplayTo,omitempty"`
	ThreadTopic  string   `wbxml:"Email.ThreadTopic,omitempty"`
	Importance   int32    `wbxml:"Email.Importance,omitempty"`
	Read         bool     `wbxml:"Email.Read,omitempty"`
	MessageClass string   `wbxml:"Email.MessageClass,omitempty"`
	ContentClass string   `wbxml:"Email.ContentClass,omitempty"`
}
