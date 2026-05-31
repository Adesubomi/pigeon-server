package auth

type MeResponse struct {
	User *User `json:"user"`
}

type GitHubLoginResponse struct {
	URL string `json:"url"`
}

type GitHubExchangeRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type GitHubExchangeInput struct {
	Code        string
	RedirectURI string
}

type AuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        *User  `json:"user"`
}
