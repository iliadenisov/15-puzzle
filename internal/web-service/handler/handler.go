package handler

import (
	"15-puzzle/internal/model"
	"15-puzzle/internal/validator"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type ctxUserID string

const (
	WebAppInitDataHeader            = "Web-App-Init-Data"
	WebAppExtraCodeHeader           = "Web-App-Extra-Code"
	WebAppHtmlFile                  = "tgwebapp.html"
	ctxDataUserID         ctxUserID = "user_id"
)

type Repository interface {
	RegisterGameStart(UserID int) (model.User, error)
	RegisterGameSolve(UserID, moves int) (model.User, error)
	Stats(UserID int) (model.User, error)
	Monitoring() (model.Monitoring, error)
	Rating() []int
}

func NewHandler(repo Repository, token, code, ctxRoot, staticDir, projectLink string) http.Handler {
	mux := http.NewServeMux()

	if abs, err := filepath.Abs(staticDir); err == nil {
		staticDir = abs
	}

	mux.Handle(http.MethodGet+" /static/", http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))
	mux.Handle(http.MethodGet+" /puzzle.html", staticFileHandler(path.Join(staticDir, WebAppHtmlFile)))

	apiMux := http.NewServeMux()
	apiMux.Handle(http.MethodGet+" /info", apiInfoHandler(model.Info{ProjectLink: projectLink}))
	apiMux.Handle(http.MethodPut+" /start", apiStartHandler(repo))
	apiMux.Handle(http.MethodPut+" /solve", apiSolveHandler(repo))
	apiMux.Handle(http.MethodGet+" /stats", apiStatsHandler(repo))
	apiMux.Handle(http.MethodGet+" /monitoring", apiMonitoringHandler(repo, code))
	apiKey := validator.EncodeHmacSha256([]byte(token), []byte("WebAppData"))
	mux.Handle("/api/", authHandler(apiKey, http.StripPrefix("/api", apiMux)))

	root := http.NewServeMux()
	root.Handle(strings.TrimRight(ctxRoot, "/")+"/", http.StripPrefix(strings.TrimRight(ctxRoot, "/"), mux))

	return root
}

func staticFileHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

func authHandler(apiKey []byte, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := apiKey
		if v, ok := r.Header[WebAppInitDataHeader]; !ok || len(v) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if ok, userID, err := validator.ValidUser(key, v[0]); err != nil {
			slog.Error(fmt.Sprintf("validator: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxDataUserID, userID)))
		}
	})
}

func apiInfoHandler(i model.Info) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, model.ApiResponse{Info: &i})
	})
}

func apiStartHandler(repo Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, r, repo.Rating, repo.RegisterGameStart)
	})
}

func apiSolveHandler(repo Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query().Get("moves")
		if moves, err := strconv.Atoi(r.URL.Query().Get("moves")); err != nil {
			errorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid value: %q", v))
		} else {
			respond(w, r, repo.Rating, func(u int) (model.User, error) { return repo.RegisterGameSolve(u, moves) })
		}
	})
}

func apiStatsHandler(repo Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond(w, r, repo.Rating, repo.Stats)
	})
}

func apiMonitoringHandler(repo Repository, code string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code == "" || code != r.Header.Get(WebAppExtraCodeHeader) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		m, err := repo.Monitoring()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, fmt.Errorf("fetch monitoring: %s", err))
			return
		}
		writeResponse(w, model.ApiResponse{Monitoring: &m})
	})
}

func respond(w http.ResponseWriter, r *http.Request, rating func() []int, action func(int) (model.User, error)) {
	userID, ok := r.Context().Value(ctxDataUserID).(int)
	if !ok {
		errorResponse(w, http.StatusInternalServerError, fmt.Errorf("unsupported context value type: %T", userID))
		return
	}

	u, err := action(userID)
	if err != nil {
		if u.UserID == 0 {
			errorResponse(w, http.StatusNotFound, fmt.Errorf("user_id=%d not found", userID))
		} else {
			errorResponse(w, http.StatusInternalServerError, fmt.Errorf("user_id=%d game action: %s", userID, err))
		}
		return
	}

	stats := &model.Stats{
		GamesStarted: u.GamesStarted,
		GamesSolved:  u.GamesSolved,
		Rank:         rankPosition(userID, rating()),
	}

	writeResponse(w, model.ApiResponse{Stats: stats, Monitoring: u.Monitoring})
}

func rankPosition(UserID int, rating []int) int {
	for i, uid := range rating {
		if uid == UserID {
			return i + 1
		}
	}
	return -1
}

func errorResponse(w http.ResponseWriter, code int, err error) {
	slog.Error(err.Error())
	w.WriteHeader(code)
	s := err.Error()
	writeResponse(w, model.ApiResponse{Err: &s})
}

func writeResponse(w http.ResponseWriter, r model.ApiResponse) {
	b, err := json.Marshal(&r)
	if err != nil {
		msg := fmt.Sprintf("response json marshal: %s", err)
		slog.Error(msg)
		b = []byte(msg)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	if _, err := w.Write(b); err != nil {
		slog.Error(fmt.Sprintf("send json response: %s", err))
	}
}
