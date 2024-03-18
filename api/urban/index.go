package urban

import (
	"net/http"
)

func Handler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("it works!"))
}
