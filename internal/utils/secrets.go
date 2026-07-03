package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func RandomSecret(bytes int) string {
	if bytes <= 0 {
		bytes = 24
	}
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "change-me-" + strings.Repeat("x", bytes)
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "=")
}

func SanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var out strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out.WriteRune(r)
		case r >= '0' && r <= '9':
			out.WriteRune(r)
		default:
			out.WriteByte('-')
		}
	}
	cleaned := strings.Trim(out.String(), "-")
	if cleaned == "" {
		return "default"
	}
	return cleaned
}
