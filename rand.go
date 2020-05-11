package main

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func nonce(size int) string {
	bb := make([]byte, size)
	for i := 0; i < size; i++ {
		bb[i] = charset[rand.Intn(len(charset))]
	}
	return string(bb)
}
