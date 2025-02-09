package pokeball

import "fmt"

const defaultStrength int = 10

type Ball struct {
	strength int
}

func NewBall() Ball {
	return Ball{
		strength: defaultStrength,
	}
}

func (b *Ball) Strength() int {
	return b.strength
}

func (b *Ball) Throw() string {
	// Introducimos un pequeño bug en la aplicación, de modo que la fuerza
	// de la bola disminuye con cada lanzamiento.
	b.strength--
	return fmt.Sprintf("Pokéball (%d) thrown!", b.strength)
}
