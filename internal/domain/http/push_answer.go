package http

type PushAnswerRequest struct {
	Email         string `binding:"required,email" json:"email"`     // Email
	PushAnswerSID string `binding:"required" json:"push-answer-sid"` // Push Answer SID
	Answer        bool   `json:"answer"`                             // Answer
}
