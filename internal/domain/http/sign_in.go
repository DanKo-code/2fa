package http

type SignInRequest struct {
	Email    string `binding:"required,email" json:"email"`    // Email
	Password string `binding:"required"       json:"password"` // Пароль
}
