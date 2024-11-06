package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var secret = []byte("Pikachu")

const (
	MinimumInsignias = 8
	passphrase       = "PokéPass_876((&^%$#	))."
	MaxPokemon       = 100
)

type Trainer struct {
	Name       string
	Role       string
	passphrase string // necesario para autenticar las peticiones de cada usuario
	Insignias  []string
	pokedex    Pokedex
}

type Pokedex struct {
	MaxPokemon int
}

func NewPokedex(maxPokemon int) Pokedex {
	return Pokedex{MaxPokemon: maxPokemon}
}

// trainers es una una representación en memoria de todos los
// entrenadores en el sistema, indexada por el nombre del
// entrenador, que debe ser único.
var trainers map[string]*Trainer

func main() {
	// inicializar los entrenadores. Cada entrenador definirá su passphrase,
	// la cual servirá para encriptar su token de autorización.
	trainers = map[string]*Trainer{
		"ash": {
			Name: "Ash",
			Role: "Trainer",
			Insignias: []string{
				"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
				"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
			},
			passphrase: "PokéPass_123((&^%$#	)).",
		},
		"misty": {
			Name: "Misty",
			Role: "Trainer",
			Insignias: []string{
				"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
				"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
			},
			passphrase: "PokéPass_456((&^%$#	)).",
		},
	}

	// encriptar las contraseñas de modo que podamos utilizarlas en la demo.
	// En un entorno de producción, las contraseñas deberían estar almacenadas
	// en otro repositorio.
	for key, tr := range trainers {
		encrypted, err := encrypt(secret, tr.passphrase)
		if err != nil {
			log.Fatalf("could not encrypt the secret for trainer %s: %v", tr.Name, err)
		}
		// log del secreto cifrado, para que podamos usarlo en la solicitud HTTP.
		// En un entorno de producción, esto sería un error de seguridad!!!
		log.Printf("%s:%s\n", key, encrypted)
	}

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

// validateUser comprueba si el entrenador existe, y si tiene las insignias necesarias
// para acceder a una ruta privada: al menos 8 insignias.
func validateUser(username string, token string) error {
	trainer, ok := trainers[username]
	if !ok {
		return errors.New("username does not exist")
	}

	decrypted, err := decrypt(token, trainer.passphrase)
	if err != nil {
		return fmt.Errorf("could not decrypt token: %v", err)
	}

	if string(decrypted) != string(secret) {
		// mostrar el secret aquí es un error de seguridad, pero lo hacemos
		// para demostrar que el token es incorrecto.
		return fmt.Errorf("user token is not valid: %s != %s", token, secret)
	}

	if len(trainer.Insignias) < MinimumInsignias {
		return fmt.Errorf("only trainers with %d or more insignias are allowed", MinimumInsignias)
	}
	return nil
}

// encrypt cifra los datos con AES-GCM y devuelve el texto cifrado en hexadecimal.
func encrypt(data []byte, passphrase string) (string, error) {
	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("could not create new GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("could not read random bytes: %v", err)
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, data, nil)), nil
}

// decrypt descifra los datos cifrados con AES-GCM y devuelve el texto original.
func decrypt(encrypted string, passphrase string) ([]byte, error) {
	data, err := hex.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("could not decode hex string: %v", err)
	}

	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return nil, fmt.Errorf("could not create new cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create new GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt data: %v", err)
	}

	return plaintext, nil
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

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

func run() error {
	http.HandleFunc("/private", CSPMiddleware(LoginMiddleware(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", username)
	})))

	// en este handler, un entrenador puede configurar su Pokedex
	http.HandleFunc("/pokedex", CSPMiddleware(LoginMiddleware(func(w http.ResponseWriter, r *http.Request) {
		maxPokemon := 6

		// por diseño, es posible configurar el número máximo de pokemons
		// que un entrenador puede tener en su Pokedex.
		if r.URL.Query().Get("max") != "" {
			fmt.Sscanf(r.URL.Query().Get("max"), "%d", &maxPokemon)
		}

		// no permitir que el entrenador tenga más pokemons que el máximo permitido
		if maxPokemon > MaxPokemon {
			maxPokemon = MaxPokemon
		}

		username := r.Context().Value(usernameKey("username")).(string)

		// en este diseño, todos los entrenadores tienen su propia Pokedex.
		trainer := trainers[username]

		trainer.pokedex = NewPokedex(maxPokemon)

		fmt.Fprintf(w, "Cool, %s! Your new Pokemon is ready! You can have %d pokemons", trainer.Name, trainer.pokedex.MaxPokemon)
	})))

	return http.ListenAndServe(":8080", nil)
}
