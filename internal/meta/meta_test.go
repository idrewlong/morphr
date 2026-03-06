package meta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEXIFMissingFile(t *testing.T) {
	_, err := ReadEXIF("/nonexistent/photo.cr2")
	if err == nil {
		t.Error("should fail on missing file")
	}
}

func TestReadEXIFInvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	fakePath := filepath.Join(tmpDir, "fake.cr2")
	os.WriteFile(fakePath, []byte("not an image"), 0o644)

	_, err := ReadEXIF(fakePath)
	if err == nil {
		t.Error("should fail on invalid EXIF data")
	}
}

func TestReadXMPSidecarMissing(t *testing.T) {
	_, err := ReadXMPSidecar("/nonexistent/photo.cr2")
	if err == nil {
		t.Error("should fail when no sidecar exists")
	}
}

func TestReadXMPSidecarValid(t *testing.T) {
	tmpDir := t.TempDir()

	xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
      xmlns:xmp="http://ns.adobe.com/xap/1.0/"
      xmp:Rating="4"
      xmp:Label="Blue">
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>`

	rawPath := filepath.Join(tmpDir, "photo.cr2")
	xmpPath := filepath.Join(tmpDir, "photo.xmp")
	os.WriteFile(rawPath, []byte("fake raw"), 0o644)
	os.WriteFile(xmpPath, []byte(xmpContent), 0o644)

	data, err := ReadXMPSidecar(rawPath)
	if err != nil {
		t.Fatalf("read XMP: %v", err)
	}

	if data.Rating != 4 {
		t.Errorf("Rating = %d, want 4", data.Rating)
	}
	if data.Label != "Blue" {
		t.Errorf("Label = %q, want %q", data.Label, "Blue")
	}
}

func TestReadXMPSidecarInvalidXML(t *testing.T) {
	tmpDir := t.TempDir()

	rawPath := filepath.Join(tmpDir, "photo.cr2")
	xmpPath := filepath.Join(tmpDir, "photo.xmp")
	os.WriteFile(rawPath, []byte("fake"), 0o644)
	os.WriteFile(xmpPath, []byte("<<<invalid xml"), 0o644)

	_, err := ReadXMPSidecar(rawPath)
	if err == nil {
		t.Error("should fail on invalid XML")
	}
}

func TestPreserveEXIFMissingExiftool(t *testing.T) {
	// This tests the error path when exiftool is present but files don't exist.
	// If exiftool is not installed, we expect a specific error message.
	err := PreserveEXIF("/nonexistent/src.cr2", "/nonexistent/dst.jpg")
	if err == nil {
		t.Error("should fail on missing files or missing exiftool")
	}
}
