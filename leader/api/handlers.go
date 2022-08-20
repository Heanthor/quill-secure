package api

import (
	"encoding/json"
	"github.com/Heanthor/quill-secure/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	r  chi.Router
	db *db.DB
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DashboardStatsResponseItem struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float32   `json:"temperature"`
	Humidity    float32   `json:"humidity"`
	Pressure    float32   `json:"pressure"`
	Altitude    float32   `json:"altitude"`
	VOCIndex    float32   `json:"vocIndex"`

	UnixTS       int64   `json:"unixTS"`
	TemperatureF float32 `json:"temperatureF"`
}

func NewRouter(db *db.DB, dashboardStatsDays int) *API {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://quillsecure.com", "https://www.quillsecure.com"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	a := API{r: r, db: db}

	r.Route("/api", func(r chi.Router) {
		r.Get("/dashboard/stats", a.getDashboardStats(dashboardStatsDays))
	})

	r.Get("/whoami", a.easterEgg)

	return &a
}

func (a *API) GetRouter() chi.Router {
	return a.r
}

// Listen listens on the router and blocks
func (a *API) Listen(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), a.r)
}

func (a *API) getDashboardStats(dashboardStatsDays int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		days, err := strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil {
			days = dashboardStatsDays
		}
		stats, err := a.db.GetRecentStats(days)
		if err != nil {
			log.Err(err).Msg("getDashboardStats db error")
			respondInternalServerError(w, err.Error())
			return
		}

		resp := make([]DashboardStatsResponseItem, len(stats))
		for i, item := range stats {
			temperatureF := item.Temperature*9/5 + 32
			resp[i] = DashboardStatsResponseItem{
				Timestamp:    item.Timestamp,
				Temperature:  item.Temperature,
				Humidity:     item.Humidity,
				Pressure:     item.Pressure,
				Altitude:     item.Altitude,
				VOCIndex:     item.VOCIndex,
				UnixTS:       item.Timestamp.Unix(),
				TemperatureF: temperatureF,
			}
		}

		if len(stats) == 0 {
			writeMessage(w, "no stats found", http.StatusNotFound)
			return
		}

		writeJSON(w, resp)
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
