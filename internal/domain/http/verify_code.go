package http

type VerifyCodeRequest struct {
	Email           string `binding:"required,email" json:"email"`       // Email
	FirstFactorCode string `binding:"required" json:"first-factor-code"` // First factor code
	FCMCode         string `binding:"required" json:"fcm-code"`          // FCM code
}
