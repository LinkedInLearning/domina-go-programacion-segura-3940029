package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	MinimumInsignias = 8
)

type Trainer struct {
	Name      string
	Role      string
	Insignias []string
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Ash no tiene insignias, por lo que no debería poder acceder a la ruta privada
	newbie := Trainer{Name: "Ash", Role: "Trainer", Insignias: []string{}}

	pokemons := []string{"Pikachu", "Charmander", "Bulbasaur", "Squirtle"}

	http.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		// respond with a list of pokemon, and the name of the trainer
		fmt.Fprintf(w, "Welcome, %s! We have these pokemons: [%s]", newbie.Name, pokemons)
	})

	return http.ListenAndServe(":8080", nil)
}
