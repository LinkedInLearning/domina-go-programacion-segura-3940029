package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/metrics"
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	appCounter := &counter{}

	if err := run(logger, appCounter); err != nil {
		log.Fatalln(err)
	}
}

func run(logger *slog.Logger, appCounter *counter) error {
	router := http.NewServeMux()

	router.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		// Cualquiera con las 8 insignias, y que conozca el token secreto,
		// puede liberar a Mewtwo,
		fmt.Fprintf(w, "Cool, %s! You can free Mewtwo now", username)
	})

	router.HandleFunc("/torneos/volcano", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		fmt.Fprintf(w, "Welcome again to the Volcano Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
	})

	router.HandleFunc("/torneos/island", func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(usernameKey("username")).(string)

		fmt.Fprintf(w, "Welcome again to the Island Tournament, %s (%s)!", username, strings.Join(trainers[username].Insignias, ","))
	})

	// en este handler, un entrenador puede configurar su Pokedex
	router.HandleFunc("/pokedex", func(w http.ResponseWriter, r *http.Request) {
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
	})

	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Get descriptions for all supported metrics.
		descs := metrics.All()

		// Create a sample for each metric.
		samples := make([]metrics.Sample, len(descs))
		for i := range samples {
			samples[i].Name = descs[i].Name
		}

		// Sample the metrics. Re-use the samples slice if you can!
		metrics.Read(samples)

		values := []any{}

		// Iterate over all results.
		for _, sample := range samples {
			// Pull out the name and value.
			name, value := sample.Name, sample.Value

			// Handle each sample.
			switch value.Kind() {
			case metrics.KindUint64:
				values = append(values, name, value.Uint64())
			case metrics.KindFloat64:
				values = append(values, name, value.Float64())
			case metrics.KindFloat64Histogram:
				// The histogram may be quite large, so let's just pull out
				// a crude estimate for the median for the sake of this example.
				values = append(values, name, value.Float64Histogram())
			case metrics.KindBad:
				// This should never happen because all metrics are supported
				// by construction.
				panic("bug in runtime/metrics package!")
			default:
				// This may happen as new metrics get added.
				//
				// The safest thing to do here is to simply log it somewhere
				// as something to look into, but ignore it for now.
				// In the worst case, you might temporarily miss out on a new metric.
				logger.Error("%s: unexpected metric Kind: %v\n", name, value.Kind())
			}
		}

		logger.Info("runtime metrics collected", values...)

		appMetrics := []any{"trainers/count", len(trainers), "requests/count", appCounter.Value()}

		logger.Info("application metrics collected", appMetrics...)
	})

	router.HandleFunc("/pokemon/", func(w http.ResponseWriter, r *http.Request) {
		// Extract Pokémon name from the URL
		path := strings.TrimPrefix(r.URL.Path, "/pokemon/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) < 2 || parts[1] != "image" {
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
			return
		}

		pokemonName := parts[0]
		imageURL := r.URL.Query().Get("image_url")

		logger.Debug("pokemon image update", "pokemon", pokemonName, "image_url", imageURL)

		// Fetch the image from the validated URL
		resp, err := http.Get(imageURL)
		if err != nil {
			http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read the image data
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read image data", http.StatusInternalServerError)
			return
		}

		// Process the image data (e.g., save to disk, display, etc.)
		// For demonstration, we'll just return the image size
		fmt.Fprintf(w, "Fetched image for Pokémon %s of size: %d bytes", pokemonName, len(imageData))
	})

	// configurar los middlewares en el orden correcto
	configuredRouter := LoggerMiddleware(logger, CounterMiddleware(appCounter, CSPMiddleware(LoginMiddleware(router))))

	return http.ListenAndServe(":8080", configuredRouter)
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
