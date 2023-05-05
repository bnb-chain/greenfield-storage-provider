package gateway

import (
	"path/filepath"
	"strings"
	"unicode/utf8"
)

//
//const HashLength = 32
//
//// IsHexHash verifies whether a string can represent a valid hex-encoded hash
//func IsHexHash(s string) bool {
//	if has0xPrefix(s) {
//		s = s[2:]
//	}
//	return len(s) == 2*HashLength && isHex(s)
//}
//
//// has0xPrefix validates str begins with '0x' or '0X'.
//func has0xPrefix(str string) bool {
//	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
//}
//
//// isHex validates whether each byte is valid hexadecimal string.
//func isHex(str string) bool {
//	if len(str)%2 != 0 {
//		return false
//	}
//	for _, c := range []byte(str) {
//		if !isHexCharacter(c) {
//			return false
//		}
//	}
//	return true
//}
//
//// isHexCharacter returns bool of c being a valid hexadecimal.
//func isHexCharacter(c byte) bool {
//	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
//}

func hasInvalidPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, p := range strings.Split(path, "/") {
		switch strings.TrimSpace(p) {
		case ".":
			return true
		case "..":
			return true
		}
	}
	return false
}

func isValidObjectPrefix(prefix string) bool {
	if hasInvalidPath(prefix) {
		return false
	}
	if !utf8.ValidString(prefix) {
		return false
	}
	if strings.Contains(prefix, `//`) {
		return false
	}
	return true
}
