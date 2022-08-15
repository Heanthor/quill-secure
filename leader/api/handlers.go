package api

import (
	"encoding/json"
	"github.com/Heanthor/quill-secure/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

type API struct {
	r  chi.Router
	db *db.DB
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewRouter(db *db.DB, dashboardStatsDays int) *API {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	a := API{r: r, db: db}

	r.Get("/dashboard/stats", a.getDashboardStats(dashboardStatsDays))
	r.Get("/", a.easterEgg)

	return &a
}

// Listen listens on the router and blocks
func (a *API) Listen(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), a.r)
}

func (a *API) getDashboardStats(dashboardStatsDays int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := a.db.GetRecentStats(dashboardStatsDays)
		if err != nil {
			log.Err(err).Msg("getDashboardStats db error")
			respondInternalServerError(w, err.Error())
			return
		}

		if len(stats) == 0 {
			writeMessage(w, "no stats found", http.StatusNotFound)
			return
		}

		writeJSON(w, stats)
	}
}

func (a *API) easterEgg(w http.ResponseWriter, r *http.Request) {
	payload := `
    ____        _ _ _  _____                          
  / __ \      (_) | |/ ____|                         
 | |  | |_   _ _| | | (___   ___  ___ _   _ _ __ ___ 
 | |  | | | | | | | |\___ \ / _ \/ __| | | | '__/ _ \
 | |__| | |_| | | | |____) |  __/ (__| |_| | | |  __/
  \___\_\\__,_|_|_|_|_____/ \___|\___|\__,_|_|  \___|
                                                     
                                                     
`
	w.Write([]byte(payload))
}

func respondInternalServerError(w http.ResponseWriter, err string) {
	writeJSON(w, ErrorResponse{Error: err}, http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, payload any, status ...int) {
	w.Header().Add("Content-Type", "application/json")
	if len(status) > 0 {
		w.WriteHeader(status[0])
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Err(err).Msg("writeJSON encoding error")
	}
}

func writeMessage(w http.ResponseWriter, payload string, status ...int) {
	msg := map[string]string{
		"message": payload,
	}

	writeJSON(w, msg, status...)
}
