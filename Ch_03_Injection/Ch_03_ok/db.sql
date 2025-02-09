CREATE TABLE IF NOT EXISTS pokedex (
    id SERIAL PRIMARY KEY,
    trainerId integer NOT NULL,
    maxPokemon integer NOT NULL
);

INSERT INTO pokedex (trainerId, maxPokemon) VALUES (1, 6);
