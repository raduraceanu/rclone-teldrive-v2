// Package api provides request and response types for the TelDrive REST API.
//
// These types map directly to the JSON payloads exchanged with the TelDrive
// server. They cover authentication (Session), file operations (CRUD, move,
// copy, share), upload parts, SSE events, and storage categories/quotas.
//
// The TelDrive API follows a flat file model where both files and folders
// are identified by unique IDs, organized via parentId references.
package api

import "time"

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e Error) Error() string {
	out := "api error"
	if e.Message != "" {
		out += ": " + e.Message
	}
	return out
}

type Part struct {
	ID   int64  `json:"id"`
	Salt string `json:"salt,omitempty"`
}

// FileInfo represents a file when listing folder contents
type FileHash struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

type FileInfo struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	MimeType   string    `json:"mimeType"`
	Size       int64     `json:"size"`
	ParentId   string    `json:"parentId"`
	Type       string    `json:"kind"`
	ModTime    time.Time `json:"modTime"`
	Hash       *FileHash `json:"hash,omitempty"`
	Encryption bool      `json:"encryption"`
}

type Meta struct {
	NextCursor string `json:"nextCursor,omitempty"`
}

type ReadMetadataResponse struct {
	Files      []FileInfo `json:"items"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

type MetadataRequestOptions struct {
	Cursor string
	Limit  int64
	Status string
}

type PartFile struct {
	Name      string `json:"name"`
	PartId    int    `json:"partId"`
	PartNo    int    `json:"partNo"`
	Size      int64  `json:"size"`
	ChannelID int64  `json:"channelId"`
	Encrypted bool   `json:"encrypted"`
	Salt      string `json:"salt"`
}

type UploadCreateRequest struct {
	ParentID          string    `json:"parentId,omitempty"`
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	MimeType          string    `json:"mimeType,omitempty"`
	ModTime           time.Time `json:"modTime"`
	Encryption        bool      `json:"encryption"`
	ConflictPolicy    string    `json:"conflictPolicy"`
	PreferredPartSize int64     `json:"preferredPartSize,omitempty"`
}

type UploadSession struct {
	ID           string `json:"id"`
	ParentID     string `json:"parentId"`
	Name         string `json:"name"`
	ExpectedSize int64  `json:"expectedSize"`
	PartSize     int64  `json:"partSize"`
	State        string `json:"state"`
	Encryption   bool   `json:"encryption"`
	FileID       string `json:"fileId"`
}

type UploadPart struct {
	UploadID  string `json:"uploadId"`
	PartNo    int    `json:"partNo"`
	State     string `json:"state"`
	PlainSize int64  `json:"plainSize"`
}

type UploadPartList struct {
	Items      []UploadPart `json:"items"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

type FolderCreateRequest struct {
	ParentID       string    `json:"parentId,omitempty"`
	Name           string    `json:"name"`
	ConflictPolicy string    `json:"conflictPolicy,omitempty"`
	ModTime        time.Time `json:"modTime,omitempty"`
}

type CreateFileRequest struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Path      string    `json:"path,omitempty"`
	MimeType  string    `json:"mimeType,omitempty"`
	Size      int64     `json:"size,omitempty"`
	ChannelID int64     `json:"channelId,omitempty"`
	Encrypted bool      `json:"encrypted,omitempty"`
	Parts     []Part    `json:"parts,omitempty"`
	ParentId  string    `json:"parentId,omitempty"`
	ModTime   time.Time `json:"updatedAt"`
	UploadId  string    `json:"uploadId"`
}

type MoveFileRequest struct {
	Destination     string   `json:"parentId,omitempty"`
	DestinationLeaf string   `json:"name,omitempty"`
	Files           []string `json:"ids,omitempty"`
	ConflictPolicy  string   `json:"conflictPolicy,omitempty"`
}

type UpdateFileInformation struct {
	Name      string     `json:"name,omitempty"`
	ModTime   *time.Time `json:"modTime,omitempty"`
	Parts     []Part     `json:"parts,omitempty"`
	Size      int64      `json:"size,omitempty"`
	UploadId  string     `json:"uploadId,omitempty"`
	ChannelID int64      `json:"channelId,omitempty"`
	ParentID  string     `json:"parentId,omitempty"`
	Encrypted bool       `json:"encrypted,omitempty"`
}

type RemoveFileRequest struct {
	Files []string `json:"ids,omitempty"`
}
type FileCopy struct {
	Destination    string    `json:"parentId,omitempty"`
	NewName        string    `json:"name,omitempty"`
	UpdatedAt      time.Time `json:"modTime,omitempty"`
	ConflictPolicy string    `json:"conflictPolicy,omitempty"`
}

type Session struct {
	UserName  string    `json:"userName"`
	UserId    int64     `json:"userId"`
	Name      string    `json:"name"`
	IsPremium bool      `json:"isPremium"`
	SessionID string    `json:"sessionId"`
	Expires   time.Time `json:"expires"`
}

type FileShare struct {
	ID        string     `json:"id,omitempty"`
	Protected bool       `json:"protected,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type FileShareCreate struct {
	Password  string     `json:"password,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type CategorySize struct {
	Category   string `json:"category"`
	TotalFiles int64  `json:"totalFiles"`
	TotalSize  int64  `json:"totalSize"`
}

type EventSource struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ParentId     string `json:"parentId"`
	DestParentId string `json:"destParentId,omitempty"`
}

type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	CreatedAt time.Time   `json:"createdAt"`
	Source    EventSource `json:"source"`
}
