package storage

type Token struct {
	URL      string `json:"url,omitempty"`
	CABundle []byte `json:"cabundle,omitempty"`
	Token    []byte `json:"token,omitempty"`
}

func NewToken(reqID string) Token {
	// TODO: Implement token generation algorithm
	return Token{
		URL:   "cbt-datapath:9000",
		Token: []byte(reqID),
	}
}

func validToken(token string) bool {
	// TODO: Implement token validation algorithm
	return true
}
