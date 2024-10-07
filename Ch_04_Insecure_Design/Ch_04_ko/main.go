package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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
	secret := []byte("Pikachu")
	passphrase := "PokéPass_876((&^%$#	))."
	encrypted, err := encrypt(secret, passphrase)
	if err != nil {
		log.Fatalf("could not encrypt the secret: %v", err)
	}

	// log del secreto cifrado, para que podamos usarlo en la solicitud HTTP.
	// En un entorno de producción, esto sería un error de seguridad!!!
	log.Println("Encrypted secret:", encrypted)

	if err := run(encrypted); err != nil {
		log.Fatalln(err)
	}
}

// checkAccess comprueba si el entrenador tiene las insignias necesarias
// para acceder a una ruta privada: al menos 8 insignias.
func checkAccess(trainer Trainer) error {
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

// hasToken comprueba si la solicitud HTTP tiene un token válido.
func hasToken(r *http.Request, encrypted string) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return fmt.Errorf("missing token")
	}

	if token != encrypted {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func run(encrypted string) error {
	// Ash tiene las 8 insignias, por lo que ya debería poder acceder a la ruta privada
	trainer := Trainer{Name: "Ash", Role: "Trainer", Insignias: []string{
		"Thunder Badge", "Marsh Badge", "Soul Badge", "Volcano Badge",
		"Earth Badge", "Cascade Badge", "Boulder Badge", "Rainbow Badge",
	}}

	http.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		if err := checkAccess(trainer); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// comprobar si la solicitud HTTP tiene un token válido
		if err := hasToken(r, encrypted); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", trainer.Name)
	})

	// en este handler, un entrenador puede configurar su Pokedex
	// este handler no comprueba si la solicitud HTTP tiene un token válido,
	// lo cual es un error de seguridad.
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

		// en este diseño, todos los entrenadores usando este handler estarían configurando
		// la Pokedex de Ash, lo cual es un bug y al mismo tiempo un error de diseño.
		pokedex := NewPokedex(trainer, maxPokemon)

		fmt.Fprintf(w, "Cool, %s! Your new Pokemon is ready! You can have %d pokemons", trainer.Name, pokedex.MaxPokemon)
	})

	return http.ListenAndServe(":8080", nil)
}
