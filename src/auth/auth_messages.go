package auth


type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TokenString string `json:"token_string"`
}

type LoginResponce struct {
	Roles []string `json:"roles, omitempty"`
	Mesg string `json:"mesg"`
	TokenString string `json:"token_string"`
	Err string `json:"err, omitempty"`
}

type LogoutRequest struct {
	TokenString string `json:"token_string"`
	Username string `json:"username"`
}

type LogoutResponce struct {
	Status bool `json:"status"`
}

type HealthRequest struct {}

type HealthResponse struct {
	Status bool `json:"status"`
}
