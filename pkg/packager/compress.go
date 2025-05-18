package packager

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TarGz compresses sourceDir into outPath (.tar.gz), creates a .sha256 file, and deletes the sourceDir.
func TarGz(sourceDir, outPath string) (string, string, error) {
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", "", err
	}
	defer outFile.Close()

	// Prepare hasher
	hasher := sha256.New()

	// Create gzip writer + tar writer
	gz := gzip.NewWriter(io.MultiWriter(outFile, hasher))
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	// Walk and add files
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(sourceDir, path)
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		return "", "", err
	}

	// Finalize tar/gzip writers
	if err := tw.Close(); err != nil {
		return "", "", err
	}
	if err := gz.Close(); err != nil {
		return "", "", err
	}

	// Write SHA256 to .sha256 file
	checksum := hex.EncodeToString(hasher.Sum(nil))
	shaFile := outPath + ".sha256"
	if err := os.WriteFile(shaFile, []byte(checksum), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write checksum: %w", err)
	}

	// Cleanup sourceDir
	if err := os.RemoveAll(sourceDir); err != nil {
		return "", "", fmt.Errorf("failed to remove source directory %q: %w", sourceDir, err)
	}

	return outFile.Name(), shaFile, nil
}
