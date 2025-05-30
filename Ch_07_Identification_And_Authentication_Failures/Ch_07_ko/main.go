package main

import (
	"errors"
	"fmt"
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
				// Misty tiene 8 insignias, pero no la de "Volcano", por lo que no podrá acceder.
				"Thunder Badge", "Marsh Badge", "Soul Badge", "Island Badge",
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
	http.HandleFunc("/private", CSPMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", username)
	}))

	http.HandleFunc("/torneos/volcano", CSPMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// introducimos un bug aquí, puesto que olvidamos verificar si el entrenador
		// tiene la insignia "Volcano Badge" para acceder a este torneo.

		fmt.Fprintf(w, "Welcome again to the Volcano Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
	}))

	http.HandleFunc("/torneos/island", CSPMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// comprobar si la solicitud HTTP tiene un token válido
		username, err := checkAccess(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if trainers[username].HasBadge("Island Badge") {
			fmt.Fprintf(w, "Welcome again to the Island Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
			return
		}

		http.Error(w, errors.New("sorry, you need the Island Badge to enter this tournament").Error(), http.StatusForbidden)
	}))

	// en este handler, un entrenador puede configurar su Pokedex
	http.HandleFunc("/pokedex", CSPMiddleware(func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Fprintf(w, "Cool, %s! Your new Pokedex is ready! You can have %d pokemons", trainer.Name, trainer.pokedex.MaxPokemon)
	}))

	return http.ListenAndServe(":8080", nil)
}

func (t *Trainer) HasBadge(badge string) bool {
	for _, b := range t.Insignias {
		if b == badge {
			return true
		}
	}
	return false
}
