package mediautil

import (
	"bytes"
	"encoding/binary"
)

// The algorithm uses at most sniffLen bytes to make its decision.
const sniffLen = 512

// DetectContentType implements the algorithm described
// at https://mimesniff.spec.whatwg.org/ to determine the
// Content-Type of the given data. It considers at most the
// first 512 bytes of data. DetectContentType always returns
// a valid MIME type: if it cannot determine a more specific one, it
// returns "application/octet-stream".
func DetectContentType(data []byte) string {
	if len(data) > sniffLen {
		data = data[:sniffLen]
	}

	// Index of the first non-whitespace byte in data.
	firstNonWS := 0
	for ; firstNonWS < len(data) && isWS(data[firstNonWS]); firstNonWS++ {
	}

	for _, sig := range sniffSignatures {
		if ct := sig.match(data, firstNonWS); ct != "" {
			return ct
		}
	}

	return "application/octet-stream" // fallback
}

// isWS reports whether the provided byte is a whitespace byte (0xWS).
func isWS(b byte) bool {
	switch b {
	case '\t', '\n', '\x0c', '\r', ' ':
		return true
	}
	return false
}

type sniffSig interface {
	// match returns the MIME type of the data, or "" if unknown.
	match(data []byte, firstNonWS int) string
}

// Data matching the table.
var sniffSignatures = []sniffSig{
	// Image types
	&exactSig{[]byte("\x00\x00\x01\x00"), "image/x-icon"},
	&exactSig{[]byte("\x00\x00\x02\x00"), "image/x-icon"},
	&exactSig{[]byte("BM"), "image/bmp"},
	&exactSig{[]byte("GIF87a"), "image/gif"},
	&exactSig{[]byte("GIF89a"), "image/gif"},
	&maskedSig{
		mask: []byte("\xFF\xFF\xFF\xFF\x00\x00\x00\x00\xFF\xFF\xFF\xFF\xFF\xFF"),
		pat:  []byte("RIFF\x00\x00\x00\x00WEBPVP"),
		ct:   "image/webp",
	},
	&exactSig{[]byte("\x89PNG\x0D\x0A\x1A\x0A"), "image/png"},
	&exactSig{[]byte("\xFF\xD8\xFF"), "image/jpeg"},
	&exactSig{[]byte("\x49\x49\x2A\x00"), "image/tiff"}, // little endian
	&exactSig{[]byte("\x4D\x4D\x00\x2A"), "image/tiff"}, // big endian

	// Video types
	&maskedSig{
		mask: []byte("\xFF\xFF\xFF\xFF\xFF"),
		pat:  []byte("OggS\x00"),
		ct:   "application/ogg",
	},
	&maskedSig{
		mask: []byte("\xFF\xFF\xFF\xFF\x00\x00\x00\x00\xFF\xFF\xFF\xFF"),
		pat:  []byte("RIFF\x00\x00\x00\x00AVI "),
		ct:   "video/avi",
	},
	// video/mp4
	mp4Sig{},
	// video/webm
	&exactSig{[]byte("\x1A\x45\xDF\xA3"), "video/webm"},
	&exactSig{[]byte("\x30\x26\xB2\x75\x8E\x66\xCF"), "video/wmv"},
	&exactSig{[]byte("FLV"), "video/flv"},
	&exactSig{[]byte("\x00\x00\x01\xBA"), "video/mpeg"}, // mpeg-2
	&exactSig{[]byte("\x1A\x45\xDF\xA3"), "video/mkv"},
	&maskedSig{
		mask: []byte("\xFF\xFF\xFF\x00\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"),
		pat:  []byte("\x00\x00\x00\x00ftypqt  "),
		ct:   "video/mov",
	},
}

type exactSig struct {
	sig []byte
	ct  string
}

func (e *exactSig) match(data []byte, firstNonWS int) string {
	if bytes.HasPrefix(data, e.sig) {
		return e.ct
	}
	return ""
}

type maskedSig struct {
	mask, pat []byte
	skipWS    bool
	ct        string
}

func (m *maskedSig) match(data []byte, firstNonWS int) string {
	if m.skipWS {
		data = data[firstNonWS:]
	}
	if len(m.pat) != len(m.mask) {
		return ""
	}
	if len(data) < len(m.pat) {
		return ""
	}
	for i, pb := range m.pat {
		maskedData := data[i] & m.mask[i]
		if maskedData != pb {
			return ""
		}
	}
	return m.ct
}

var mp4ftype = []byte("ftyp")
var mp4 = []byte("mp4")

type mp4Sig struct{}

func (mp4Sig) match(data []byte, firstNonWS int) string {
	if len(data) < 12 {
		return ""
	}
	boxSize := int(binary.BigEndian.Uint32(data[:4]))
	if len(data) < boxSize || boxSize%4 != 0 {
		return ""
	}
	if !bytes.Equal(data[4:8], mp4ftype) {
		return ""
	}
	for st := 8; st < boxSize; st += 4 {
		if st == 12 {
			continue
		}
		if bytes.Equal(data[st:st+3], mp4) {
			return "video/mp4"
		}
	}
	return ""
}
