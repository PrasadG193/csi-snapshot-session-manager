package controller

import (
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	tokenLength = 32
	ssdPrefix   = "csi-cbt-"
)

func newToken() string {
	return ssdPrefix + rand.String(tokenLength)
}
