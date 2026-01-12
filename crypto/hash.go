package crypto

import (
	sha256 "github.com/zeebo/blake3"
)

func Sha256(bytes []byte) []byte {
	hasher := sha256.New()
	hasher.Write(bytes)
	return hasher.Sum(nil)
}
