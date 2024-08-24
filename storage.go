package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/JeanLeonHenry/chirpy/internal/database"
)

type returnValues struct {
	CleanedBody string `json:"cleaned_body"`
	Error       string `json:"error"`
}

func (cfg *apiConfig) handlerSaveChirp(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprint("Error reading request body: ", err))
		return
	}
	chirp, err := cfg.db.CreateChirp(string(data))
	if err != nil {
		err_msg := fmt.Sprint("Error creating chirp: ", err)
		if err.Error() == database.ERR_CHIRP_TOO_LONG {
			respondWithError(w, http.StatusBadRequest, err_msg)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err_msg)
		return
	}
	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerReadChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerReadChirpById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	for _, c := range chirps {
		if c.Id == id {
			respondWithJSON(w, http.StatusOK, c)
			return
		}
	}
	respondWithError(w, http.StatusNotFound, "")
}
