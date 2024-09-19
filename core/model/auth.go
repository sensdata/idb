package model

type Login struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResult struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
