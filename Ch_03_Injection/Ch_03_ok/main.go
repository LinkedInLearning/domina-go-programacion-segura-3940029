package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	MinimumInsignias = 8
	MaxPokemon       = 100
)

type Trainer struct {
	Name      string
	Role      string
	Insignias []string
}

type Pokedex struct {
	Owner      string
	MaxPokemon int
}

func NewPokedex(owner Trainer, maxPokemon int) Pokedex {
	return Pokedex{Owner: owner.Name, MaxPokemon: maxPokemon}
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Ash tiene las 8 insignias, por lo que ya debería poder acceder a la ruta privada
	trainer := Trainer{Name: "Ash", Role: "Trainer", Insignias: []string{
		"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
		"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
	}}

	// en este handler, un entrenador puede configurar su Pokedex
	http.HandleFunc("/pokedex", func(w http.ResponseWriter, r *http.Request) {
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

		pokedex := NewPokedex(trainer, maxPokemon)

		fmt.Fprintf(w, "Cool, %s! Your new Pokemon is ready! You can have %d pokemons", trainer.Name, pokedex.MaxPokemon)
	})

	return http.ListenAndServe(":8080", nil)
}
