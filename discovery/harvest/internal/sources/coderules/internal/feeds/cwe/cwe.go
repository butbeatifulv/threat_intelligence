package cwe

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const zipURL = "https://cwe.mitre.org/data/xml/cwec_latest.xml.zip"

// ZipURL is the MITRE CWE catalog zip download URL.
func ZipURL() string { return zipURL }

// CatalogWriter receives one CWE weakness row from the MITRE zip stream.
type CatalogWriter interface {
	UpsertCWECatalog(ctx context.Context, cweID, name, description, status string) error
}

// StreamMITRE downloads the CWE catalog zip and streams Weakness elements to w.
func StreamMITRE(ctx context.Context, w CatalogWriter, maxWeakness int) error {
	if maxWeakness <= 0 {
		maxWeakness = 5000
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zipURL, nil)
	if err != nil {
		return err
	}
	cl := &http.Client{Timeout: 10 * time.Minute}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("cwe zip: %s: %s", resp.Status, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 40<<20))
	if err != nil {
		return err
	}
	return StreamMITREFromZip(ctx, w, maxWeakness, body)
}

// StreamMITREFromZip parses a CWE catalog zip and streams Weakness elements to w.
func StreamMITREFromZip(ctx context.Context, w CatalogWriter, maxWeakness int, body []byte) error {
	if maxWeakness <= 0 {
		maxWeakness = 5000
	}
	br := bytes.NewReader(body)
	zr, err := zip.NewReader(br, int64(len(body)))
	if err != nil {
		return err
	}
	var xmlRC io.ReadCloser
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, ".xml") && strings.HasPrefix(f.Name, "cwec_v") {
			xmlRC, err = f.Open()
			if err != nil {
				return err
			}
			break
		}
	}
	if xmlRC == nil {
		return fmt.Errorf("cwe zip: no cwec_v*.xml found")
	}
	defer xmlRC.Close()

	dec := xml.NewDecoder(xmlRC)
	dec.Strict = false
	n := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Weakness" {
			continue
		}
		var wk struct {
			ID          string `xml:"ID,attr"`
			Name        string `xml:"Name,attr"`
			Status      string `xml:"Status,attr"`
			Description string `xml:"Description"`
		}
		if err := dec.DecodeElement(&wk, &se); err != nil {
			continue
		}
		cid := "CWE-" + strings.TrimPrefix(strings.TrimSpace(wk.ID), "CWE-")
		desc := strings.TrimSpace(wk.Description)
		if err := w.UpsertCWECatalog(ctx, cid, wk.Name, desc, wk.Status); err != nil {
			return err
		}
		n++
		if n >= maxWeakness {
			break
		}
	}
	return nil
}
