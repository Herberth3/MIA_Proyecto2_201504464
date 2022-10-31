package main

import (
	command "File_Manager_GO/lib"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rs/cors"
)

type cmdstruct struct {
	Cmd string `json:"cmd"`
}

type respuesta struct {
	Consola string
	IsLogin int
	RepDot  string
}

func main() {

	//file, err := os.ReadFile("./entrada1.txt")

	mux := http.NewServeMux()

	mux.HandleFunc("/analizar", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		var Content cmdstruct
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &Content)

		command.Analizar(Content.Cmd)

		res := respuesta{
			Consola: command.Recolector.Salida,
			IsLogin: 0,
			RepDot:  command.Recolector.RepDot}

		jsonResponse, jsonError := json.Marshal(res)

		if jsonError != nil {
			fmt.Println("Unable to encode JSON")
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})

	fmt.Println("Server ON in port 5000")
	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(":5000", handler))

	//if err != nil {
	//	fmt.Println("Error: ", err)
	//} else {
	//	text := string(file)
	//	command.Analizar(text)
	//}
}
