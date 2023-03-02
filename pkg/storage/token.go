package storage

type Token struct {
	URL      string `json:"url,omitempty"`
	CABundle []byte `json:"cabundle,omitempty"`
	Token    []byte `json:"token,omitempty"`
}

func NewToken(reqID string) Token {
	return Token{
		URL:   "cbt-svc:9000",
		Token: []byte(reqID),
	}
}
