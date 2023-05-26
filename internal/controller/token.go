package controller

import (
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	tokenLength = 32
	ssdPrefix   = "csi-cbt-"
)

func newToken() string {
	return rand.String(tokenLength)
}

func generateSnapSessionDataName() string {
	return SnapSessionDataNameWithToken(newToken())
}

func SnapSessionDataNameWithToken(token string) string {
	return ssdPrefix + token
}
