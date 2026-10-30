// Package transfer safely maps files into an SSH session's private workspace.
// An SSH terminal cannot itself accept a GUI drag-and-drop transfer: use SFTP.
package transfer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MaxUpload = 16 << 20 // 16 MiB bounded memory for initial WhatsApp uploads.

// ReadUpload only accepts regular files contained under root, including after
// resolving symlinks. The upload directory must not be shared with untrusted users.
func ReadUpload(root, name string) ([]byte, string, error) {
	if strings.TrimSpace(name) == "" {
		return nil, "", errors.New("file name is required")
	}
	base, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, "", fmt.Errorf("upload directory: %w", err)
	}
	base, err = filepath.Abs(base)
	if err != nil {
		return nil, "", err
	}
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(base, path)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil, "", err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, "", err
	}
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, "", errors.New("file is outside the upload directory")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, "", err
	}
	if !info.Mode().IsRegular() {
		return nil, "", errors.New("only regular files may be uploaded")
	}
	if info.Size() > MaxUpload {
		return nil, "", fmt.Errorf("file exceeds %d-byte limit", MaxUpload)
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxUpload+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > MaxUpload {
		return nil, "", errors.New("file grew beyond upload size limit")
	}
	return data, filepath.Base(path), nil
}

// SafeName ignores path separators, shell metacharacters and control chars in
// remote filenames. The resulting name is for local display/storage only.
func SafeName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		if b.Len() >= 80 {
			break
		}
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), ". ")
	if out == "" {
		return "attachment.bin"
	}
	return out
}

// SaveDownload writes exclusively inside a private directory. It removes
// partially written files after network failures, and never overwrites files.
func SaveDownload(root, name string, write func(*os.File) error) (string, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	if err := os.Chmod(root, 0700); err != nil {
		return "", err
	}
	// Avoid following a symlink provided as the download directory.
	info, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("download root is not a real directory")
	}
	basename := SafeName(name)
	for i := 0; i < 1000; i++ {
		candidate := basename
		if i > 0 {
			candidate = fmt.Sprintf("%d-%s", i, basename)
		}
		path := filepath.Join(root, candidate)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		writeErr := write(f)
		closeErr := f.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(path)
			return "", errors.Join(writeErr, closeErr)
		}
		return path, nil
	}
	return "", errors.New("too many duplicate download names")
}
