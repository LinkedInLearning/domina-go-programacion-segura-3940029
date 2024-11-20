package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// CSPMiddleware aplica una política de seguridad de contenido a todas las respuestas HTTP.
// Para ello escribe una cabecera Content-Security-Policy en cada respuesta, evitando ataques XSS.
// Más información: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
func CSPMiddleware(next http.Handler) http.Handler {
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
func LoginMiddleware(next http.Handler) http.Handler {
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

func LoggerMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
		}

		log.Info("request started", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr)

		now := time.Now()

		// envolvemos el http.ResponseWriter para poder registrar el código de estado
		// los handlers sucesivos podrán modificar el código de estado de la respuesta HTTP,
		// en el método WriteHeader.
		lrw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)

		after := time.Since(now)

		log.Info("request completed", "method", r.Method, "path", path, "remoteaddr", r.RemoteAddr,
			"statuscode", lrw.statusCode, "since", after.String())
	})
}

// TournamentMiddleware es un middleware que comprueba si el usuario tiene el badge necesario
// para acceder a la ruta HTTP. Si no lo tiene, devuelve un error 403 Forbidden.
// Para ello, extrae el nombre del badge de la ruta HTTP, y comprueba si el usuario tiene ese badge.
func TournamentMiddleware(next http.Handler) http.Handler {
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
			http.Error(w, "missing badge "+badgeName, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
