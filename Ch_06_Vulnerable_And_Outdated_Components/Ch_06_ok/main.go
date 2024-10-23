package main

import (
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

	"outdated-components/pokeball"
)

const (
	MinimumInsignias = 8
	MaxPokemon       = 100
)

type Trainer struct {
	Name      string
	Role      string
	token     string // necesario para autenticar las peticiones de cada usuario
	Insignias []string
	pokedex   Pokedex
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
	secret := []byte("Pikachu")

	// inicializar los entrenadores. Cada entrenador definirá su password,
	// la cual será encriptada.
	trainers = map[string]*Trainer{
		"ash": {
			Name: "Ash",
			Role: "Trainer",
			Insignias: []string{
				"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
				"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
			},
			token: "PokéPass_123((&^%$#	)).",
		},
		"misty": {
			Name: "Misty",
			Role: "Trainer",
			Insignias: []string{
				"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
				"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
			},
			token: "PokéPass_456((&^%$#	)).",
		},
	}

	// encriptar las contraseñas de modo que podamos utilizarlas en la demo.
	// En un entorno de producción, las contraseñas deberían estar almacenadas
	// en otro repositorio.
	for key, tr := range trainers {
		encrypted, err := encrypt(secret, tr.token)
		if err != nil {
			log.Fatalf("could not encrypt the secret for trainer %s: %v", tr.Name, err)
		}
		tr.token = encrypted
		// log del secreto cifrado, para que podamos usarlo en la solicitud HTTP.
		// En un entorno de producción, esto sería un error de seguridad!!!
		log.Printf("%s:%s\n", key, tr.token)
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

	if token != trainer.token {
		return fmt.Errorf("user token is not valid: %s != %s", token, trainer.token)
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

// checkAccess comprueba si la solicitud HTTP tiene un token válido,
// con el formato: "Authorization: username:token", y además verifica
// que el usuario sea correcto (exista en la base de datos en memoria,
// y además tenga las 8 medallas)
func checkAccess(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(header, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed Authorization header")
	}

	username := parts[0]
	token := parts[1]

	if err := validateUser(username, token); err != nil {
		return "", fmt.Errorf("check access: %w", err)
	}

	return username, nil
}

func run() error {
	http.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", username)
	})

	http.HandleFunc("/capture", func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// identificar al pokemon a capturar
		pokemon := PokemonAppears()

		fmt.Fprintf(w, "A %s appeared. It's attack is %d\n", pokemon.Name, pokemon.Attack)

		// el nuevo componente requiere al entrenador para lanzar una pokeball
		ball := pokeball.NewBall(username)

		// lanzar la pokeball e imprimir el resultado
		fmt.Fprintln(w, ball.Throw())

		if ball.Strength() > pokemon.Attack {
			// Si la pokeball es más fuerte que el Pokemon, entonces el Pokemon es capturado.
			fmt.Fprintf(w, "Cool, %s! You captured a %s", username, pokemon.Name)
		} else {
			fmt.Fprintf(w, "Oh no, %s! The %s ran away", username, pokemon.Name)
		}
	})

	// en este handler, un entrenador puede configurar su Pokedex
	http.HandleFunc("/pokedex", func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

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

		// en este diseño, todos los entrenadores tienen su propia Pokedex.
		trainer := trainers[username]

		trainer.pokedex = NewPokedex(maxPokemon)

		fmt.Fprintf(w, "Cool, %s! Your new Pokemon is ready! You can have %d pokemons", trainer.Name, trainer.pokedex.MaxPokemon)
	})

	return http.ListenAndServe(":8080", nil)
}
