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

		if err := json.NewEncoder(w).Encode(stats); err != nil {
			log.Err(err).Msg("getDashboardStats encoding error")
		}
	}
}

func respondInternalServerError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: err}); err != nil {
		log.Err(err).Msg("getDashboardStats encoding error")
	}
}
