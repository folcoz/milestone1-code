package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/folcoz/milestone1-code/secrets"
)

type (
	postInput struct {
		PlainText string `json:"plain_text"`
	}

	postOutput struct {
		Id string `json:"id"`
	}

	getOutput struct {
		Data string `json:"data"`
	}
)

const ctypeJSON string = "application/json"

func StartListener(address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthcheckHandler)
	mux.HandleFunc("/", secretHandler)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	fmt.Println("Starting server at", server.Addr)
	server.ListenAndServe()
}

func healthcheckHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "ok")
}

func secretHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		getSecretHandler(writer, request)
	case "POST":
		postSecretHandler(writer, request)
	default:
		writer.WriteHeader(405)
	}
}

func postSecretHandler(writer http.ResponseWriter, request *http.Request) {
	// Check request content is application/json
	if !contentTypeIs(request, ctypeJSON) {
		badRequest(writer, "Content-Type must be application/json")
		return
	}
	// Unmarshal to input struct
	var input postInput
	err := json.NewDecoder(request.Body).Decode(&input)
	if err != nil {
		badRequest(writer, err.Error())
		return
	}

	plainText := input.PlainText
	if plainText == "" {
		badRequest(writer, "plain_text value is empty")
		return
	}

	var hash string
	hash, err = secrets.SaveSecret(plainText)
	if err != nil {
		serverError(writer, err.Error())
		return
	}

	// Write json response
	sendSecretId(writer, hash)
}

func sendSecretId(writer http.ResponseWriter, hash string) {
	var output postOutput
	output.Id = hash
	writer.Header().Set("Content-Type", ctypeJSON)
	json.NewEncoder(writer).Encode(output)
}

func sendSecretData(writer http.ResponseWriter, value string) {
	var output getOutput
	output.Data = value
	writer.Header().Set("Content-Type", ctypeJSON)
	json.NewEncoder(writer).Encode(output)
}

func contentTypeIs(request *http.Request, expected string) bool {
	ctype := request.Header.Get("Content-Type")
	return ctype == ctypeJSON
}

func sendError(writer http.ResponseWriter, statusCode int, msg string) {
	writer.WriteHeader(statusCode)
	if msg != "" {
		var prefix string
		switch {
		case statusCode >= 500:
			prefix = "Server Error: "
		case statusCode >= 400:
			prefix = "Bad Request: "
		}
		fmt.Fprintf(writer, "%s%s", prefix, msg)
	}
}

func badRequest(writer http.ResponseWriter, msg string) {
	sendError(writer, 400, msg)
}

func serverError(writer http.ResponseWriter, msg string) {
	sendError(writer, 500, msg)
}

func sendSecretNotFound(writer http.ResponseWriter) {
	writer.WriteHeader(404)
	writer.Header().Set("Content-Type", ctypeJSON)
	fmt.Fprintln(writer, `{"data":""}`)
}

func getSecretHandler(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Path[1:]
	if id == "" {
		sendSecretNotFound(writer)
		return
	}

	value, err := secrets.LoadSecret(id)
	if err != nil {
		serverError(writer, err.Error())
		return
	}

	if value == "" {
		sendSecretNotFound(writer)
		return
	}

	sendSecretData(writer, value)
}
