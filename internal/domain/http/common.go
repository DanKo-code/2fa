package http

type Response[T any] struct {
	Data    T      `json:"data"`    // Данные
	Message string `json:"message"` // Сообщение
}

func NewErrorResponse(err string) Response[any] {
	return Response[any]{
		Message: err,
	}
}
