package gigya

type Response struct {
	ErrorCode    int         `json:"errorCode"`    // /accounts.login
	ErrorMessage string      `json:"errorMessage"` // /accounts.login
	SessionInfo  SessionInfo `json:"sessionInfo"`  // /accounts.login
	IDToken      string      `json:"id_token"`     // /accounts.getJWT
	Data         Data        `json:"data"`         // /accounts.getAccountInfo
}

type SessionInfo struct {
	CookieValue string `json:"cookieValue"`
}

type Data struct {
	PersonID string `json:"personId"`
}
