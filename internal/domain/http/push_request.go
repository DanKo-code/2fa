package http

type PushRequestRequest struct {
	Email string `binding:"required,email" json:"email"` // Email
}
