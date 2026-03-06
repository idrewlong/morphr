package meta

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type XMPData struct {
	Rating      int
	Label       string
	Title       string
	Description string
	Creator     string
	Keywords    []string
}

// ReadXMPSidecar looks for a .xmp file alongside the given RAW file and parses it.
func ReadXMPSidecar(rawPath string) (*XMPData, error) {
	ext := filepath.Ext(rawPath)
	xmpPath := strings.TrimSuffix(rawPath, ext) + ".xmp"

	data, err := os.ReadFile(xmpPath)
	if err != nil {
		return nil, fmt.Errorf("no XMP sidecar found: %w", err)
	}

	var xmp xmpMeta
	if err := xml.Unmarshal(data, &xmp); err != nil {
		return nil, fmt.Errorf("parse XMP: %w", err)
	}

	result := &XMPData{
		Label: xmp.RDF.Description.Label,
	}
	if xmp.RDF.Description.Rating != "" {
		fmt.Sscanf(xmp.RDF.Description.Rating, "%d", &result.Rating)
	}

	return result, nil
}

type xmpMeta struct {
	XMLName xml.Name       `xml:"adobe:ns:meta/ xmpmeta"`
	RDF     xmpRDFWrapper  `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
}

type xmpRDFWrapper struct {
	Description xmpDescription `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# Description"`
}

type xmpDescription struct {
	Rating string `xml:"http://ns.adobe.com/xap/1.0/ Rating,attr"`
	Label  string `xml:"http://ns.adobe.com/xap/1.0/ Label,attr"`
}
