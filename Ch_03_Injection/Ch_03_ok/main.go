package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

const (
	MinimumInsignias = 8
)

type Trainer struct {
	ID        int
	Name      string
	Role      string
	Insignias []string
}

type Pokedex struct {
	Owner      string
	MaxPokemon int
}

func NewPokedex(owner Trainer, maxPokemon int) (Pokedex, error) {
	var p Pokedex
	err := updatePokedex(context.TODO(), owner.ID, maxPokemon)
	if err != nil {
		return p, fmt.Errorf("update pokedex: %v", err)
	}

	if err != nil {
		return p, fmt.Errorf("convert maxPokemon: %v", err)
	}

	p = Pokedex{Owner: owner.Name, MaxPokemon: maxPokemon}

	return p, nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Ash tiene las 8 insignias, por lo que ya debería poder acceder a la ruta privada
	trainer := Trainer{ID: 1, Name: "Ash", Role: "Trainer", Insignias: []string{
		"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
		"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
	}}

	// en este handler, un entrenador puede configurar su Pokedex
	http.HandleFunc("/pokedex", func(w http.ResponseWriter, r *http.Request) {
		maxParam := r.URL.Query().Get("max")

		// ahora sí verificamos que el valor del parámetro max sea un número
		max, err := strconv.Atoi(maxParam)
		if err != nil {
			http.Error(w, "max must be a number", http.StatusBadRequest)
			return
		}

		pokedex, err := NewPokedex(trainer, max)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Cool, %s! Your new Pokedex is ready! You can have %d pokemons", trainer.Name, pokedex.MaxPokemon)
	})

	return http.ListenAndServe(":8080", nil)
}
