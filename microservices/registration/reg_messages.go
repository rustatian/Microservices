package registration

type RegRequest struct {
	Username   string `json:"username"`
	Fullname   string `json:"fullname"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsDisabled bool   `json:"is_disabled"`
}

type RegResponce struct {
	Status bool   `json:"status"`
	Err    string `json:"err, omitempty"`
}

type UsernameValidationRequest struct {
	User string `json:"user"`
}

type UsernameValidationResponce struct {
	Status bool   `json:"status"`
	Err    string `json:"err, omitempty"`
}

type EmailValidationRequest struct {
	Email string `json:"email"`
}

type EmailValidationResponce struct {
	Status bool   `json:"status"`
	Err    string `json:"err, omitempty"`
}

type HealthRequest struct {
}

type HealthResponse struct {
	Status bool `json:"status"`
}

type hashResponse struct {
	Hash string `json:"hash"`
	Err  string `json:"err, omitempty"`
}
