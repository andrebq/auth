package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/andrebq/auth"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Handler(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/token", tokenAuth(db))
	mux.Handle("/auth/login", loginAuth(db))
	mux.Handle("/session", newSessionHandler(db))
	return mux
}

func tokenAuth(db *sql.DB) http.Handler {
	sampler := zerolog.Sometimes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token struct {
			Token string `json:"token"`
		}
		if !decode(&token, w, r) {
			return
		}
		uid, tokenType, err := auth.TokenLogin(r.Context(), db, token.Token)
		if err != nil {
			log := log.Logger.Sample(sampler)
			log.Error().Err(err).Msg("Authentication failed")
			encode(w, 0, UnauthorizedError("Invalid credentials"))
			return
		}
		tokenID, _ := auth.ExtractTokenID(token.Token)
		encode(w, http.StatusOK, struct {
			UID       string `json:"uid"`
			TokenID   string `json:"tokenID"`
			TokenType string `json:"tokenType"`
		}{
			UID:       uid,
			TokenType: tokenType,
			TokenID:   tokenID,
		})
	})
}

func loginAuth(db *sql.DB) http.Handler {
	sampler := zerolog.Sometimes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if !decode(&user, w, r) {
			return
		}
		var uid string
		var err error
		if uid, err = auth.Login(r.Context(), db, user.Login, []byte(user.Password)); err != nil {
			log := log.Logger.Sample(sampler)
			log.Error().Err(err).Msg("Authentication failed")
			encode(w, 0, UnauthorizedError("Invalid credentials"))
			return
		}
		encode(w, http.StatusOK, struct {
			UID string `json:"userID"`
		}{UID: uid})
	})
}

func newSessionHandler(db *sql.DB) http.Handler {
	sampler := zerolog.Sometimes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Login    string      `json:"login"`
			Password string      `json:"password"`
			TTL      apiDuration `json:"ttl"`
		}
		if !decode(&user, w, r) {
			return
		}
		if user.TTL == 0 {
			user.TTL = apiDuration(time.Minute * 24)
		}
		if time.Duration(user.TTL) < time.Second {
			user.TTL = apiDuration(time.Second)
		}
		_, err := auth.Login(r.Context(), db, user.Login, []byte(user.Password))
		if err != nil {
			log := log.Logger.Sample(sampler)
			log.Error().Err(err).Msg("Authentication failed")
			encode(w, 0, UnauthorizedError("Invalid credentials"))
			return
		}
		token, err := auth.CreateToken(r.Context(), db, user.Login, "session", time.Now().Add(time.Duration(user.TTL)))
		if err != nil {
			log.Error().Err(err).Msg("Unable to create token for user")
			encode(w, 0, InternalError())
		}
		encode(w, http.StatusOK, struct {
			Token string `json:"token"`
		}{
			Token: token,
		})
	})
}

func decode(out interface{}, w http.ResponseWriter, req *http.Request) bool {
	err := json.NewDecoder(req.Body).Decode(out)
	if err != nil {
		encode(w, 0, BadRequestError("cannot decode request body"))
		return false
	}
	return true
}

func encode(w http.ResponseWriter, status int, data interface{}) {
	if hasStatus, ok := data.(interface{ HTTPStatus() int }); ok && status == 0 {
		status = hasStatus.HTTPStatus()
	}

	buf, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Cannot encode response to user")
		return
	}
	w.Header().Add("Content-Length", strconv.Itoa(len(buf)))
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf)
}
