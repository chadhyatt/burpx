package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func ExtractItems(rt *Root, out string) (err error) {
	dirMap := make(map[string]bool)
	dupePathMap := make(map[string]int)

	for _, item := range rt.Items {
		u, err := url.Parse(item.Url)
		if err != nil {
			continue
		}

		path := fixedFilePath(u)
		for _, dir := range ParentDirs(path) {
			dirMap[dir] = true
		}
	}

	for _, item := range rt.Items {
		if *skipNonSuccess && item.Status < 200 || item.Status >= 299 {
			slog.Info("Skipping non-2xx response status for item", "url", item.Url)
			continue
		} else if *skipNonGet && item.Method != "GET" {
			slog.Info("Skipping non-GET response for item", "url", item.Url)
			continue
		}

		var resp *http.Response
		{
			var respPayload []byte
			if item.Response.IsBase64 {
				if respPayload, err = base64.StdEncoding.DecodeString(item.Response.Data); err != nil {
					slog.Error("Failed to parse base64 HTTP response payload from item", "err", err, "url", item.Url)
					continue
				}
			} else {
				respPayload = []byte(item.Response.Data)
			}

			respPayload = normalizeHttpResp(respPayload)

			payloadReader := bufio.NewReader(bytes.NewReader(respPayload))
			if resp, err = http.ReadResponse(payloadReader, nil); err != nil {
				slog.Error("Failed to parse HTTP response payload from item", "err", err, "url", item.Url)
				continue
			}
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")

		u, err := url.Parse(item.Url)
		if err != nil {
			slog.Error("Failed to parse url for item", "err", err, "url", item.Url)
			continue
		}

		path := fixedFilePath(u)

		{
			dupIdx, ok := dupePathMap[path]
			if !ok {
				dupIdx = 0
			}
			if dupIdx > 0 {
				if !*shouldWriteDups {
					slog.Warn("Skipping duplicate path", "url", item.Url)
					continue
				}

				path = fmt.Sprintf("%s_%d", path, dupIdx)
			}
			dupePathMap[path] = dupIdx + 1
		}

		if _, ok := dirMap[path]; ok {
			path = filepath.Join(path, "index")
		}

		if !strings.Contains(filepath.Base(path), ".") {
			ext := DefaultExtForContentType(contentType)
			if ext != "" {
				path = path + ext
			}
		}

		realOutPath := filepath.Join(out, path)

		slog.Info(fmt.Sprintf("Extracting %s", item.Url))

		if err := os.MkdirAll(filepath.Dir(realOutPath), 0777); err != nil {
			slog.Error("Failed to create dir for item", "err", err, "url", item.Url)
			continue
		}
		if err := os.WriteFile(realOutPath, respBody, 0777); err != nil {
			slog.Error("Failed to write file for item", "err", err, "url", item.Url)
			continue
		}
	}

	return nil
}

func fixedFilePath(u *url.URL) string {
	return "/" + filepath.Join(u.Host, filepath.Clean(u.EscapedPath()))
}

// Do I blame Burp or Go (or both) for this blasphemy?
func normalizeHttpResp(payload []byte) []byte {
	from := "HTTP/2"
	to := "HTTP/2.0"

	if len(payload) == 0 {
		return payload
	}

	eolIdx := bytes.IndexByte(payload, '\n')
	if eolIdx == -1 {
		eolIdx = bytes.IndexByte(payload, '\r')
		if eolIdx == -1 {
			eolIdx = len(payload)
		}
	}

	start := 0
	for start < eolIdx {
		b := payload[start]
		if b != ' ' && b != '\t' {
			break
		}
		start++
	}

	if start >= eolIdx || eolIdx-start < len(from) || !bytes.Equal(payload[start:start+len(from)], []byte(from)) {
		return payload
	}

	if start+len(from) < len(payload) {
		next := payload[start+len(from)]
		if next == '.' {
			return payload
		}
	}

	out := []byte{}
	out = append(out, payload[:start]...)
	out = append(out, to...)
	out = append(out, payload[start+len(from):]...)

	return out
}
