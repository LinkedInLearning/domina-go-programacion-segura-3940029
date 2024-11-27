package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

type CommandPayload struct {
	Command string `json:"command"` // The command to execute
}

func main() {
	http.HandleFunc("/manage-pokemon", managePokemonHandler)
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func managePokemonHandler(w http.ResponseWriter, r *http.Request) {
	var payload CommandPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Ejecución del comando sin validar su contenido
	output, err := exec.Command("sh", "-c", payload.Command).Output()
	if err != nil {
		http.Error(w, "Command execution failed", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Command output: %s", output)
}
