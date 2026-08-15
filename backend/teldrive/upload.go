// Package teldrive implements the TelDrive v2 resumable upload protocol.
package teldrive

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/rclone/rclone/backend/teldrive/api"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/object"
	"github.com/rclone/rclone/lib/rest"
)

const memoryBufferThreshold = 10 * 1024 * 1024

type uploadInfo struct {
	existingChunks map[int]api.PartFile
	uploadID       string
	encryptFile    bool
	chunkSize      int64
	totalChunks    int64
	fileName       string
	dir            string
}

type objectChunkWriter struct {
	f          *Fs
	src        fs.ObjectInfo
	o          *Object
	uploadInfo *uploadInfo
}

func getMD5Hash(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}

func stableUUID(text string) string {
	v := getMD5Hash(text)
	return v[0:8] + "-" + v[8:12] + "-" + v[12:16] + "-" + v[16:20] + "-" + v[20:32]
}

func (w *objectChunkWriter) WriteChunk(ctx context.Context, chunkNumber int, reader io.ReadSeeker) (int64, error) {
	partNo := chunkNumber + 1
	if partNo < 1 {
		return 0, fmt.Errorf("invalid chunk number %d", chunkNumber)
	}
	if existing, ok := w.uploadInfo.existingChunks[partNo]; ok {
		return existing.Size, nil
	}
	size, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if _, err = reader.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	opts := rest.Opts{
		Method: "PUT", Body: reader, ContentLength: &size, ContentType: "application/octet-stream",
		Path: "/api/v1/uploads/" + w.uploadInfo.uploadID + "/parts/" + strconv.Itoa(partNo),
	}
	if w.f.opt.UploadHost != "" {
		opts.RootURL = joinAPIURL(w.f.opt.UploadHost, opts.Path)
		opts.Path = ""
	}
	err = w.f.pacer.Call(func() (bool, error) {
		resp, err := w.f.srv.Call(ctx, &opts)
		return shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return 0, fmt.Errorf("upload part %d: %w", partNo, err)
	}
	return size, nil
}

func (w *objectChunkWriter) Close(ctx context.Context) error {
	return w.o.createFile(ctx, w.src, w.uploadInfo)
}

func (w *objectChunkWriter) Abort(ctx context.Context) error {
	opts := rest.Opts{Method: "DELETE", Path: "/api/v1/uploads/" + w.uploadInfo.uploadID, NoResponse: true}
	_, err := w.f.srv.Call(ctx, &opts)
	return err
}

func (o *Object) prepareUpload(ctx context.Context, remote string, src fs.ObjectInfo) (*uploadInfo, error) {
	leaf, directoryID, err := o.fs.dirCache.FindPath(ctx, remote, true)
	if err != nil {
		return nil, err
	}
	payload := api.UploadCreateRequest{
		ParentID: directoryID, Name: leaf, Size: src.Size(), MimeType: fs.MimeType(ctx, src),
		ModTime: src.ModTime(ctx).UTC(), Encryption: o.fs.opt.EncryptFiles,
		ConflictPolicy: "replace", PreferredPartSize: int64(o.fs.opt.ChunkSize),
	}
	idempotencyKey := stableUUID(fmt.Sprintf("upload:%s:%s:%d:%d:%d", directoryID, leaf, src.Size(), src.ModTime(ctx).UTC().UnixNano(), o.fs.userId))
	opts := rest.Opts{
		Method: "POST", Path: "/api/v1/uploads",
		ExtraHeaders: map[string]string{"Idempotency-Key": idempotencyKey},
	}
	var session api.UploadSession
	err = o.fs.pacer.Call(func() (bool, error) {
		resp, err := o.fs.srv.CallJSON(ctx, &opts, &payload, &session)
		return shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return nil, fmt.Errorf("create upload: %w", err)
	}
	chunkSize := session.PartSize
	if chunkSize <= 0 {
		chunkSize = int64(o.fs.opt.ChunkSize)
	}
	existing := make(map[int]api.PartFile)
	params := url.Values{"limit": []string{"200"}}
	for {
		var page api.UploadPartList
		listOpts := rest.Opts{Method: "GET", Path: "/api/v1/uploads/" + session.ID + "/parts", Parameters: params}
		err = o.fs.pacer.Call(func() (bool, error) {
			resp, err := o.fs.srv.CallJSON(ctx, &listOpts, nil, &page)
			return shouldRetry(ctx, resp, err)
		})
		if err != nil {
			return nil, fmt.Errorf("list upload parts: %w", err)
		}
		for _, part := range page.Items {
			if part.State == "stored" {
				existing[part.PartNo] = api.PartFile{PartNo: part.PartNo, Size: part.PlainSize}
			}
		}
		if page.NextCursor == "" {
			break
		}
		params.Set("cursor", page.NextCursor)
	}
	totalChunks := int64(0)
	if src.Size() > 0 {
		totalChunks = (src.Size() + chunkSize - 1) / chunkSize
	}
	return &uploadInfo{
		existingChunks: existing, uploadID: session.ID, encryptFile: session.Encryption,
		chunkSize: chunkSize, totalChunks: totalChunks, fileName: leaf, dir: directoryID,
	}, nil
}

func (o *Object) uploadMultipart(ctx context.Context, remote string, in io.Reader, src fs.ObjectInfo) (*uploadInfo, error) {
	info, err := o.prepareUpload(ctx, remote, src)
	if err != nil {
		return nil, err
	}
	var uploaded int64
	for partNo := 1; partNo <= int(info.totalChunks); partNo++ {
		partSize := info.chunkSize
		if remaining := src.Size() - uploaded; remaining < partSize {
			partSize = remaining
		}
		if existing, ok := info.existingChunks[partNo]; ok {
			if _, err := io.CopyN(io.Discard, in, existing.Size); err != nil {
				return nil, err
			}
			uploaded += existing.Size
			continue
		}
		partReader := io.LimitReader(in, partSize)
		opts := rest.Opts{
			Method: "PUT", Body: partReader, ContentLength: &partSize, ContentType: "application/octet-stream",
			Path: "/api/v1/uploads/" + info.uploadID + "/parts/" + strconv.Itoa(partNo),
		}
		if o.fs.opt.UploadHost != "" {
			opts.RootURL = joinAPIURL(o.fs.opt.UploadHost, opts.Path)
			opts.Path = ""
		}
		err = o.fs.pacer.Call(func() (bool, error) {
			resp, err := o.fs.srv.Call(ctx, &opts)
			return shouldRetry(ctx, resp, err)
		})
		if err != nil {
			return nil, fmt.Errorf("upload part %d: %w", partNo, err)
		}
		uploaded += partSize
	}
	return info, nil
}

func (o *Object) createFile(ctx context.Context, _ fs.ObjectInfo, info *uploadInfo) error {
	opts := rest.Opts{
		Method: "POST", Path: "/api/v1/uploads/" + info.uploadID + "/complete",
		ExtraHeaders: map[string]string{"Idempotency-Key": stableUUID("complete:" + info.uploadID)},
	}
	err := o.fs.pacer.Call(func() (bool, error) {
		resp, err := o.fs.srv.CallJSON(ctx, &opts, nil, nil)
		return shouldRetry(ctx, resp, err)
	})
	if err != nil {
		return fmt.Errorf("complete upload: %w", err)
	}
	return nil
}

func (o *Object) uploadWithBuffering(ctx context.Context, remote string, in io.Reader, src fs.ObjectInfo) (*uploadInfo, int64, error) {
	var memory bytes.Buffer
	var file *os.File
	var size int64
	buf := make([]byte, 128*1024)
	for {
		n, readErr := in.Read(buf)
		if n > 0 {
			if file == nil && size+int64(n) > memoryBufferThreshold {
				file, readErr = os.CreateTemp("", "rclone-teldrive-*")
				if readErr != nil {
					return nil, 0, readErr
				}
				if _, readErr = file.Write(memory.Bytes()); readErr != nil {
					return nil, 0, readErr
				}
				memory.Reset()
			}
			if file != nil {
				_, readErr = file.Write(buf[:n])
			} else {
				_, readErr = memory.Write(buf[:n])
			}
			if readErr != nil {
				return nil, 0, readErr
			}
			size += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, 0, readErr
		}
	}
	withSize := object.NewStaticObjectInfo(remote, src.ModTime(ctx), size, false, nil, o.fs)
	if file != nil {
		name := file.Name()
		defer func() { _ = file.Close(); _ = os.Remove(name) }()
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, 0, err
		}
		info, err := o.uploadMultipart(ctx, remote, file, withSize)
		return info, size, err
	}
	info, err := o.uploadMultipart(ctx, remote, bytes.NewReader(memory.Bytes()), withSize)
	return info, size, err
}

// joinAPIURL preserves an api_host path prefix (for example /tgd behind Caddy).
func joinAPIURL(base, endpoint string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(endpoint, "/")
}
