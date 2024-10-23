package pokeball

import "fmt"

const defaultStrength int = 10_000

type Ball struct {
	trainerName string
	strength    int
}

func NewBall(trainerName string) Ball {
	return Ball{
		trainerName: trainerName,
		strength:    defaultStrength,
	}
}

func (b Ball) Strength() int {
	return b.strength
}

func (b Ball) Throw() string {
	return fmt.Sprintf("%s threw a Pokéball (%d)!", b.trainerName, b.strength)
}
