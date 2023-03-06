package storage

type Token struct {
	URL      string `json:"url,omitempty"`
	CABundle []byte `json:"cabundle,omitempty"`
	Token    []byte `json:"token,omitempty"`
}

func NewToken(reqID string) Token {
	return Token{
		URL:   "cbt-datapath:9000",
		Token: []byte(reqID),
	}
}

func validToken(token string) bool {
	// TOD
	return true
}
