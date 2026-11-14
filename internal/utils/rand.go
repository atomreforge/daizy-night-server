package utils

import (
	"crypto/rand"
	"math/big"
)

// GenRandID returns a random 7-digit decimal id, range [1000000, 9999999].
func GenRandID() (uint, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9000000)) // [0, 9000000)
	if err != nil {
		return 0, err
	}
	return uint(n.Int64()) + 1000000, nil
}
