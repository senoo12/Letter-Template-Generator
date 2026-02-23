package domain

type Letter struct {
	Data map[string]interface{}
}

func NewLetter(data map[string]interface{}) *Letter {
	return &Letter{
		Data: data,
	}
}