package iq

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/unanet/go/pkg/log"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetUniqueName() string {
	hostname, set := os.LookupEnv("HOSTNAME")
	if set {
		return hostname
	}

	log.Logger.Warn("HOSTNAME unset as env variable, generating iq instance name")
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("unanet-%s-%s", randSeq(7), randSeq(5))
}
