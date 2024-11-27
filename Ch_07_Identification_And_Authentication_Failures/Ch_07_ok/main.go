package main

import (
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

func run() error {
	http.HandleFunc("/private", CSPMiddleware(LoginMiddleware(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", username)
	})))

	http.HandleFunc("/torneos/volcano", CSPMiddleware(LoginMiddleware(TournamentMiddleware(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		fmt.Fprintf(w, "Welcome again to the Volcano Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
	}))))

	http.HandleFunc("/torneos/island",
		CSPMiddleware(
			LoginMiddleware(
				TournamentMiddleware(
					func(w http.ResponseWriter, r *http.Request) {
						username := r.Context().Value(usernameKey("username")).(string)

						fmt.Fprintf(w, "Welcome again to the Island Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
					},
				),
			),
		),
	)

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

		fmt.Fprintf(w, "Cool, %s! Your new Pokedex is ready! You can have %d pokemons", trainer.Name, trainer.pokedex.MaxPokemon)
	})))

	return http.ListenAndServe(":8080", nil)
}

func (t *Trainer) HasBadge(badge string) bool {
	for _, b := range t.Insignias {
		// comparamos sin distinción entre mayúsculas y minúsculas
		if strings.EqualFold(b, badge) {
			return true
		}
	}
	return false
}
