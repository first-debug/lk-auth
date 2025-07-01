package schemas

type Tokens struct {
	Access_token  string `json:"access_token"`
	Refresh_token string `json:"refresh_token"`
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
