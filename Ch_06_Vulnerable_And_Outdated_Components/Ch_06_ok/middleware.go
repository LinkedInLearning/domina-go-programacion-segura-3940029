package main

import "net/http"

// CSPMiddleware aplica una política de seguridad de contenido a todas las respuestas HTTP.
// Para ello escribe una cabecera Content-Security-Policy en cada respuesta, evitando ataques XSS.
// Más información: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
func CSPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}
