package binary

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// extractEntry pulls a single file out of an archive (.tar.gz / .tgz / .zip)
// and writes it to destPath. entryPath is the path of the target file inside
// the archive.
func extractEntry(archivePath, entryPath, destPath string) error {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(archivePath, entryPath, destPath)
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(archivePath, entryPath, destPath)
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func extractTarGz(archivePath, entryPath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}
		if hdr.Name == entryPath {
			return writeStream(tr, destPath)
		}
	}
	return fmt.Errorf("entry %q not found in archive", entryPath)
}

func extractZip(archivePath, entryPath, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == entryPath {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			return writeStream(rc, destPath)
		}
	}
	return fmt.Errorf("entry %q not found in zip", entryPath)
}

func writeStream(r io.Reader, destPath string) error {
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}
