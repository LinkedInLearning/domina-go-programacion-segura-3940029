DROP DATABASE IF EXISTS `pokemon`;
CREATE DATABASE `pokemon`  /*!40100 DEFAULT CHARACTER SET utf8 */;

CREATE TABLE pokedex (
    id SERIAL PRIMARY KEY,
    trainerId integer NOT NULL,
    maxPokemon integer NOT NULL
);

INSERT INTO pokedex (trainerId, maxPokemon) VALUES (1, 6);
