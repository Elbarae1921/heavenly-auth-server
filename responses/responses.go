package responses

import (
	"main/exceptions"
)

type Response struct {
	Error int
	Data  []byte
}

// exceptions enum
var EXCEPTIONS = map[string]int{
	exceptions.INTERNAL_ERROR:      111,
	exceptions.INVALID_CREDENTIALS: 112,
}

func Success(data []byte) Response {
	// if data, err := utils.Encode(data); err != nil {
	// 	return Response{
	// 		Error: EXCEPTIONS[exceptions.INTERNAL_ERROR],
	// 	}
	// } else {
	return Response{
		Error: 0,
		Data:  data,
	}
	// }
}

func Fail(exceptionString string) Response {
	if exception, ok := EXCEPTIONS[exceptionString]; ok {
		return Response{
			Error: exception,
			Data:  nil,
		}
	}
	return Response{
		Error: EXCEPTIONS[exceptions.INTERNAL_ERROR],
		Data:  nil,
	}
}
