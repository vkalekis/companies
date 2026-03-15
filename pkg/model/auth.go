package model

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWT struct {
	JWT string `json:"jwt"`
}
