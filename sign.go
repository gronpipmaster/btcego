package btcego

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

var unreserved = make([]bool, 128)
var charsTof = "0123456789ABCDEF"

func init() {
	// RFC3986
	u := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890-_.~"
	for _, c := range u {
		unreserved[c] = true
	}
}

// Encode takes a string and URI-encodes it in a way suitable
// to be used in Btc-e signatures.
func Encode(s string) string {
	encode := false
	for i := 0; i != len(s); i++ {
		c := s[i]
		if c > 127 || !unreserved[c] {
			encode = true
			break
		}
	}
	if !encode {
		return s
	}
	e := make([]byte, len(s)*3)
	ei := 0
	for i := 0; i != len(s); i++ {
		c := s[i]
		if c > 127 || !unreserved[c] {
			e[ei] = '%'
			e[ei+1] = charsTof[c>>4]
			e[ei+2] = charsTof[c&0xF]
			ei += 3
		} else {
			e[ei] = c
			ei += 1
		}
	}
	return string(e[:ei])
}

type sign struct {
	signature, key string
}

//See https://btc-e.com/api/documentation section "Authentication"
func NewSign(auth Auth, params map[string]string) *sign {
	var sarray []string
	for k, v := range params {
		sarray = append(sarray, Encode(k)+"="+Encode(v))
	}
	payload := strings.Join(sarray, "&")
	hash := hmac.New(sha512.New, []byte(auth.SecretKey))
	hash.Write([]byte(payload))
	signature := strings.ToLower(hex.EncodeToString(hash.Sum(nil)))
	return &sign{signature, auth.AccessKey}
}
