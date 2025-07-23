package rest

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var Salt = func() string { return rand.Text() }
var SessionKey = Salt()

func SignSession(username string) string {
	data := fmt.Sprintf("%s:%d", username, time.Now().Unix())
	sum := sha256.Sum256([]byte(SessionKey + data))
	sig := base32.StdEncoding.EncodeToString(sum[:])[:16]
	return fmt.Sprintf("%s.%s", data, sig)
}

func VerifySession(session string) (string, bool) {
	parts := strings.Split(session, ".")
	if len(parts) != 2 {
		return "", false
	}
	data, sig := parts[0], parts[1]
	sum := sha256.Sum256([]byte(SessionKey + data))
	expectedSig := base32.StdEncoding.EncodeToString(sum[:])[:16]
	if sig != expectedSig {
		return "", false
	}
	if parts = strings.Split(data, ":"); len(parts) == 2 {
		if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			if time.Now().Unix()-ts < 86400 { // 24 hours
				return parts[0], true
			}
		}
	}
	return "", false
}
