package crypto

import (
	"crypto/sha1"
	"math/big"
)

func HashKey(key string) *big.Int {
	h := sha1.New()
	h.Write([]byte(key))
	return new(big.Int).SetBytes(h.Sum(nil))
}
