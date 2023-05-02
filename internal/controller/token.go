package controller

import "os"

type Token struct {
	URL      string `json:"url,omitempty"`
	CABundle []byte `json:"cabundle,omitempty"`
	Token    []byte `json:"token,omitempty"`
}

func NewToken(reqID string) Token {
	// TODO: Implement token generation algorithm
	return Token{
		URL:   os.Getenv("EXT_SNAP_SESSION_SVC_URL"),
		Token: []byte(reqID),
	}
}

func ValidToken(token string) bool {
	// TODO: Implement token validation algorithm
	return true
}
