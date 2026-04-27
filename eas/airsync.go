package eas

// SyncRequest is the MS-ASCMD Sync command request payload.
type SyncRequest struct {
	XMLName     struct{}        `wbxml:"AirSync.Sync"`
	Collections SyncCollections `wbxml:"AirSync.Collections"`
}

// SyncResponse is the MS-ASCMD Sync command response payload.
type SyncResponse struct {
	XMLName     struct{}        `wbxml:"AirSync.Sync"`
	Status      int32           `wbxml:"AirSync.Status,omitempty"`
	Collections SyncCollections `wbxml:"AirSync.Collections"`
}

// SyncCollections wraps the Collection entries in a Sync request/response.
type SyncCollections struct {
	Collection []SyncCollection `wbxml:"AirSync.Collection"`
}

// SyncCollection is a per-collection entry inside a Sync request/response.
type SyncCollection struct {
	SyncKey       string        `wbxml:"AirSync.SyncKey"`
	CollectionID  string        `wbxml:"AirSync.CollectionId"`
	Class         string        `wbxml:"AirSync.Class,omitempty"`
	GetChanges    int32         `wbxml:"AirSync.GetChanges,omitempty"`
	WindowSize    int32         `wbxml:"AirSync.WindowSize,omitempty"`
	Status        int32         `wbxml:"AirSync.Status,omitempty"`
	MoreAvailable int32         `wbxml:"AirSync.MoreAvailable,omitempty"`
	Options       *SyncOptions  `wbxml:"AirSync.Options,omitempty"`
	Commands      *SyncCommands `wbxml:"AirSync.Commands,omitempty"`
	Responses     *SyncCommands `wbxml:"AirSync.Responses,omitempty"`
}

// SyncOptions is the per-collection Options element for a Sync request.
type SyncOptions struct {
	FilterType     int32             `wbxml:"AirSync.FilterType,omitempty"`
	Class          string            `wbxml:"AirSync.Class,omitempty"`
	MIMESupport    int32             `wbxml:"AirSync.MIMESupport,omitempty"`
	MIMETruncation int32             `wbxml:"AirSync.MIMETruncation,omitempty"`
	MaxItems       int32             `wbxml:"AirSync.MaxItems,omitempty"`
	BodyPreference []BodyPreference  `wbxml:"AirSyncBase.BodyPreference,omitempty"`
}

// BodyPreference is the AirSyncBase preference declaration.
type BodyPreference struct {
	Type            int32 `wbxml:"AirSyncBase.Type"`
	TruncationSize  int32 `wbxml:"AirSyncBase.TruncationSize,omitempty"`
	AllOrNone       int32 `wbxml:"AirSyncBase.AllOrNone,omitempty"`
	Preview         int32 `wbxml:"AirSyncBase.Preview,omitempty"`
}

// SyncCommands wraps Add/Change/Delete/Fetch commands inside a Sync
// request/response.
type SyncCommands struct {
	Add    []SyncAdd    `wbxml:"AirSync.Add,omitempty"`
	Change []SyncChange `wbxml:"AirSync.Change,omitempty"`
	Delete []SyncDelete `wbxml:"AirSync.Delete,omitempty"`
	Fetch  []SyncFetch  `wbxml:"AirSync.Fetch,omitempty"`
}

// SyncAdd carries a server-pushed addition or a client-side new item.
type SyncAdd struct {
	ServerID        string  `wbxml:"AirSync.ServerId,omitempty"`
	ClientID        string  `wbxml:"AirSync.ClientId,omitempty"`
	ApplicationData *AppRaw `wbxml:"AirSync.ApplicationData,omitempty"`
}

// SyncChange carries an item modification.
type SyncChange struct {
	ServerID        string  `wbxml:"AirSync.ServerId"`
	ApplicationData *AppRaw `wbxml:"AirSync.ApplicationData,omitempty"`
}

// SyncDelete carries an item deletion notification.
type SyncDelete struct {
	ServerID string `wbxml:"AirSync.ServerId"`
}

// SyncFetch carries an explicit Fetch request.
type SyncFetch struct {
	ServerID string `wbxml:"AirSync.ServerId"`
}

// AppRaw is an opaque carrier for AirSync.ApplicationData.
//
// Real domain types (Email/Appointment/Contact/Task) marshal independently
// from AirSync.ApplicationData; for command-level round trips we only need
// the wrapper element to be preserved. Concrete decoding is done by the
// caller after extracting the raw bytes if needed.
type AppRaw struct{}
