package cmd

import (
	"encoding/json"
	"net/http"
)

func respondError(w http.ResponseWriter, req *http.Request, msg string, statusCode int) {
	obj := make(map[string]interface{})
	obj["message"] = msg
	obj["type"] = "error"
	obj["code"] = statusCode

	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err == nil {
		http.Error(w, string(bytes), statusCode)
	} else {
		http.Error(w, "{\"type\": \"error\", \"message\": \"JSON marshal error\"}", http.StatusInternalServerError)
	}
}

func respondSuccess(w http.ResponseWriter, req *http.Request, val interface{}) {
	bytes, err := json.MarshalIndent(val, "", "  ")
	if err == nil {
		w.Write(bytes)
	} else {
		respondError(w, req, "Error serializing to JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func respond(w http.ResponseWriter, req *http.Request, val interface{}, err error) {
	if err == nil {
		respondSuccess(w, req, val)
	} else {
		respondError(w, req, err.Error(), 500)
	}
}

type Collection struct {
	Type         string      `json:"type"`
	ResourceType string      `json:"resourceType"`
	Data         interface{} `json:"data"`
}
