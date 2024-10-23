package main

import "math/rand"

type Pokemon struct {
	Name    string
	Attack  int
	Defense int
	Life    int
	Types   []string
	Moves   []string
}

func Pikachu() Pokemon {
	return Pokemon{
		Name:    "Pikachu",
		Attack:  55,
		Defense: 40,
		Life:    35,
		Types:   []string{"Electric"},
		Moves:   []string{"Thunder Shock", "Quick Attack"},
	}
}

func Charmander() Pokemon {
	return Pokemon{
		Name:    "Charmander",
		Attack:  52,
		Defense: 43,
		Life:    39,
		Types:   []string{"Fire"},
		Moves:   []string{"Ember", "Scratch"},
	}
}

func Bulbasaur() Pokemon {
	return Pokemon{
		Name:    "Bulbasaur",
		Attack:  49,
		Defense: 49,
		Life:    45,
		Types:   []string{"Grass", "Poison"},
		Moves:   []string{"Vine Whip", "Tackle"},
	}
}

func Squirtle() Pokemon {
	return Pokemon{
		Name:    "Squirtle",
		Attack:  48,
		Defense: 65,
		Life:    44,
		Types:   []string{"Water"},
		Moves:   []string{"Water Gun", "Tackle"},
	}
}

// PokemonAppears simula la aparición de un pokemon de los cuatro posibles,
// devolviendo un pokemon aleatorio.
func PokemonAppears() Pokemon {
	pokemons := []Pokemon{Pikachu(), Charmander(), Bulbasaur(), Squirtle()}
	return pokemons[rand.Intn(4)]
}
