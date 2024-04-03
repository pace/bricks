// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.18
// +build go1.18

package scan

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime/debug"

	"golang.org/x/vuln/internal/buildinfo"
	"golang.org/x/vuln/internal/client"
	"golang.org/x/vuln/internal/derrors"
	"golang.org/x/vuln/internal/govulncheck"
	"golang.org/x/vuln/internal/vulncheck"
)

// runBinary detects presence of vulnerable symbols in an executable or its minimal blob representation.
func runBinary(ctx context.Context, handler govulncheck.Handler, cfg *config, client *client.Client) (err error) {
	defer derrors.Wrap(&err, "govulncheck")

	bin, err := createBin(cfg.patterns[0])
	if err != nil {
		return err
	}

	p := &govulncheck.Progress{Message: binaryProgressMessage}
	if err := handler.Progress(p); err != nil {
		return err
	}
	return vulncheck.Binary(ctx, handler, bin, &cfg.Config, client)
}

func createBin(path string) (*vulncheck.Bin, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// First check if the path points to a Go binary. Otherwise, blob
	// parsing might json decode a Go binary which takes time.
	//
	// TODO(#64716): use fingerprinting to make this precise, clean, and fast.
	mods, packageSymbols, bi, err := buildinfo.ExtractPackagesAndSymbols(f)
	if err == nil {
		return &vulncheck.Bin{
			Modules:    mods,
			PkgSymbols: packageSymbols,
			GoVersion:  bi.GoVersion,
			GOOS:       findSetting("GOOS", bi),
			GOARCH:     findSetting("GOARCH", bi),
		}, nil
	}

	// Otherwise, see if the path points to a valid blob.
	bin := parseBlob(f)
	if bin != nil {
		return bin, nil
	}

	return nil, errors.New("unrecognized binary format")
}

// parseBlob extracts vulncheck.Bin from a valid blob. If it
// cannot recognize a valid blob, returns nil.
func parseBlob(from io.Reader) *vulncheck.Bin {
	dec := json.NewDecoder(from)

	var h header
	if err := dec.Decode(&h); err != nil {
		return nil // no header
	} else if h.Name != extractModeID || h.Version != extractModeVersion {
		return nil // invalid header
	}

	var b vulncheck.Bin
	if err := dec.Decode(&b); err != nil {
		return nil // no body
	}
	if dec.More() {
		return nil // we want just header and body, nothing else
	}
	return &b
}

// findSetting returns value of setting from bi if present.
// Otherwise, returns "".
func findSetting(setting string, bi *debug.BuildInfo) string {
	for _, s := range bi.Settings {
		if s.Key == setting {
			return s.Value
		}
	}
	return ""
}
