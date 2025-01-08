package model

type AuthResponse struct {
	User  AuthUser `json:"User"`
	Token string   `json:"AccessToken"`
}

type AuthUser struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
}

type AuthRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}
