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

func (b Ball) Strength() int {
	return b.strength
}

func (b Ball) Throw() string {
	return fmt.Sprintf("Pokéball (%d) thrown!", b.strength)
}
