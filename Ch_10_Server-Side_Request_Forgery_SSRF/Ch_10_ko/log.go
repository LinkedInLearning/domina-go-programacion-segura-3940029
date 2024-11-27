package main

import "net/http"

// loggingResponseWriter es un tipo que envuelve un http.ResponseWriter.
// Nos permite registrar el código de estado de la respuesta HTTP.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader escribe el código de estado de la respuesta HTTP.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
