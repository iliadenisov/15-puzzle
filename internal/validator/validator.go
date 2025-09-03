package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
)

// ValidUser returns validation result of initData with key used for hashing.
// If validation was successfull then value of boolean is true and initData's user.id value is returned as int.
// Otherwise, including when error is not nil, the returned boolean will be false and int will be zero.
//
// Implemented according to https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidUser(key []byte, initData string) (bool, int, error) {
	m, err := url.ParseQuery(initData)
	if err != nil {
		return false, 0, fmt.Errorf("parse init data %q: %s", initData, err)
	}
	var hash []byte
	if v, ok := m["hash"]; ok && len(v) > 0 {
		hash, err = hex.DecodeString(v[0])
		if err != nil {
			return false, 0, fmt.Errorf("decode original hash %q: %s", v[0], err)
		}
		delete(m, "hash")
	} else {
		return false, 0, nil
	}

	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)

	u := new(struct {
		ID *int `json:"id"`
	})
	fields := make([]string, 0)
	for _, k := range keys {
		if v, ok := m[k]; ok && len(v) == 1 {
			if k == "user" {
				if err := json.Unmarshal([]byte(v[0]), &u); err != nil {
					return false, 0, fmt.Errorf("unmarshall user data %s: %s", v[0], err)
				}
				if u.ID == nil {
					return false, 0, fmt.Errorf("user.id field not set: %s: ", v[0])
				}
			}
			fields = append(fields, k+"="+v[0])
		}
	}
	if slices.Equal(EncodeHmacSha256([]byte(strings.Join(fields, "\n")), key), hash) {
		return true, *u.ID, nil
	}

	return false, 0, nil
}

func EncodeHmacSha256(data, key []byte) []byte {
	sig := hmac.New(sha256.New, key)
	sig.Write(data)
	return sig.Sum(nil)
}
