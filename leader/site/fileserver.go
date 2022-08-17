package site

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type Fileserver struct {
	r   chi.Router
	dir http.FileSystem
}

func NewFileserver(r chi.Router, staticSitePath string) *Fileserver {
	filesDir := http.Dir(staticSitePath)

	return &Fileserver{r: r, dir: filesDir}
}

// FileRoutes conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
// from: https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go
func (f *Fileserver) FileRoutes(sitePath string) {
	if strings.ContainsAny(sitePath, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if sitePath != "/" && sitePath[len(sitePath)-1] != '/' {
		f.r.Get(sitePath, http.RedirectHandler(sitePath+"/", 301).ServeHTTP)
		sitePath += "/"
	}
	sitePath += "*"

	f.r.Get(sitePath, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(f.dir))
		fs.ServeHTTP(w, r)
	})
}
