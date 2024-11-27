package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// CSPMiddleware aplica una política de seguridad de contenido a todas las respuestas HTTP.
// Para ello escribe una cabecera Content-Security-Policy en cada respuesta, evitando ataques XSS.
// Más información: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
func CSPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

// usernameKey es un tipo para la clave de contexto que contiene el nombre de usuario.
// De esta forma, podemos pasar el nombre de usuario a través de los middlewares,
// y finalmente recuperarlo en el handler final.
type usernameKey string

// LoginMiddleware es un middleware que comprueba si la solicitud HTTP tiene un token válido,
// con el formato: "Authorization: username:token", y además verifica
// que el usuario sea correcto (exista en la base de datos en memoria,
// y además tenga las 8 medallas)
func LoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, errors.New("missing Authorization header").Error(), http.StatusForbidden)
			return
		}

		parts := strings.Split(header, ":")
		if len(parts) != 2 {
			http.Error(w, errors.New("malformed Authorization header").Error(), http.StatusForbidden)
			return
		}

		username := parts[0]
		token := parts[1]

		if err := validateUser(username, token); err != nil {
			http.Error(w, fmt.Errorf("validate user: %w", err).Error(), http.StatusForbidden)
			return
		}

		// pasar el usuario al siguiente handler
		r = r.WithContext(context.WithValue(r.Context(), usernameKey("username"), username))

		next.ServeHTTP(w, r)
	})
}

// TournamentMiddleware es un middleware que comprueba si el usuario tiene el badge necesario
// para acceder a la ruta HTTP. Si no lo tiene, devuelve un error 403 Forbidden.
// Para ello, extrae el nombre del badge de la ruta HTTP, y comprueba si el usuario tiene ese badge.
// El formato del endpoint es: */tournaments/{badge}/, de modo que la segunda posición del path
// contiene el nombre del badge.
func TournamentMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		// leer el nombre de la ruta HTTP, y de ahí extraer el nombre del badge a buscar
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "missing badge name", http.StatusBadRequest)
			return
		}

		badgeName := parts[2] + " badge"

		if !trainers[username].HasBadge(badgeName) {
			http.Error(w, "missing "+badgeName, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
