package site

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const manifestName = ".visual-guide-manifest.json"

// Replace rather than truncate, so an existing hard link cannot change a file
// outside the build directory and a failed write cannot leave a partial file.
func replaceFile(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".visual-guide-write-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(0644); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Write updates only an empty directory or a verified previous build. Unknown
// files, modified generated files, and symlinks cause a refusal before writes.
// No directory trees are removed; stale generated files are removed individually.
func Write(output string, files map[string][]byte) error {
	root, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	for dir := root; ; dir = filepath.Dir(dir) {
		info, err := os.Lstat(dir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output uses symlink: %s", dir)
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	old := map[string]string{}
	if info, err := os.Lstat(filepath.Join(root, manifestName)); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("build manifest must be a regular file")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(root, manifestName))
	if err == nil {
		if err := json.Unmarshal(manifest, &old); err != nil {
			return fmt.Errorf("invalid build manifest: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	for name := range old {
		if !fs.ValidPath(name) || name == "." || name == manifestName {
			return fmt.Errorf("unsafe manifest path %q", name)
		}
		if info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(name))); err == nil && !info.Mode().IsRegular() {
			return fmt.Errorf("manifest entry is not a regular file: %s", name)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	for name := range files {
		if !fs.ValidPath(name) || name == "." || name == manifestName {
			return fmt.Errorf("unsafe output path %q", name)
		}
	}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) && path == root {
			return nil
		}
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing output symlink: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("refusing non-regular output: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == manifestName {
			return nil
		}
		hash, ok := old[name]
		if !ok {
			return fmt.Errorf("refusing unrecognized output file: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if digest(data) != hash {
			return fmt.Errorf("refusing modified output file: %s", path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	next := map[string]string{}
	for name, data := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := replaceFile(path, data); err != nil {
			return err
		}
		next[name] = digest(data)
	}
	for name := range old {
		if _, ok := next[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(name))); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	return replaceFile(filepath.Join(root, manifestName), append(data, '\n'))
}
