package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/evcc-io/evcc/core/site"
	"github.com/evcc-io/evcc/util"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Assets is the embedded assets file system
var Assets fs.FS

type route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// HTTPd wraps an http.Server and adds the root router
type HTTPd struct {
	*http.Server
}

type ContextKey struct{}

var (
	CtxSite, CtxLoadpoint ContextKey
)

func siteHandlerContext(site site.API) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, CtxSite, site)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loadpointHandlerContext(lp int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			site, ok := ctx.Value(CtxSite).(site.API)
			if !ok {
				http.Error(w, "invalid site context", http.StatusInternalServerError)
				return
			}

			loadpoints := site.LoadPoints()
			if lp >= len(loadpoints) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			ctx = context.WithValue(ctx, CtxLoadpoint, loadpoints[lp])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// NewHTTPd creates HTTP server with configured routes for loadpoint
func NewHTTPd(url string, site site.API, hub *SocketHub, cache *util.Cache) *HTTPd {
	router := mux.NewRouter().StrictSlash(true)

	// websocket
	router.HandleFunc("/ws", socketHandler(hub))

	// static - individual handlers per root and folders
	static := router.PathPrefix("/").Subrouter()
	static.Use(handlers.CompressHandler)

	static.HandleFunc("/", indexHandler(site))
	for _, dir := range []string{"css", "js", "ico"} {
		static.PathPrefix("/" + dir).Handler(http.FileServer(http.FS(Assets)))
	}

	// api
	api := router.PathPrefix("/api").Subrouter()
	api.Use(jsonHandler)
	api.Use(handlers.CompressHandler)
	api.Use(handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Accept", "Accept-Language", "Content-Language", "Content-Type", "Origin",
		}),
	))
	api.Use(siteHandlerContext(site))

	// site api
	routes := map[string]route{
		"health": {[]string{"GET"}, "/health", healthHandler},
		"state":  {[]string{"GET"}, "/state", stateHandler(cache)},
	}

	for _, r := range routes {
		api.Methods(r.Methods...).Path(r.Pattern).Handler(r.HandlerFunc)
	}

	// loadpoint api
	for lp := 0; lp <= 9; lp++ {
		api := api.PathPrefix(fmt.Sprintf("/loadpoints/%d", lp)).Subrouter()
		api.Use(loadpointHandlerContext(lp))

		routes := map[string]route{
			"mode":          {[]string{"POST", "OPTIONS"}, "/mode/{value:[a-z]+}", chargeModeHandler},
			"targetsoc":     {[]string{"POST", "OPTIONS"}, "/targetsoc/{value:[0-9]+}", targetSoCHandler},
			"minsoc":        {[]string{"POST", "OPTIONS"}, "/minsoc/{value:[0-9]+}", minSoCHandler},
			"mincurrent":    {[]string{"POST", "OPTIONS"}, "/mincurrent/{value:[0-9]+}", minCurrentHandler},
			"maxcurrent":    {[]string{"POST", "OPTIONS"}, "/maxcurrent/{value:[0-9]+}", maxCurrentHandler},
			"phases":        {[]string{"POST", "OPTIONS"}, "/phases/{value:[0-9]+}", phasesHandler},
			"targetcharge":  {[]string{"POST", "OPTIONS"}, "/targetcharge/{soc:[0-9]+}/{time:[0-9TZ:-]+}", targetChargeHandler},
			"targetcharge2": {[]string{"DELETE", "OPTIONS"}, "/targetcharge", targetChargeRemoveHandler},
			"remotedemand":  {[]string{"POST", "OPTIONS"}, "/remotedemand/{demand:[a-z]+}/{source::[0-9a-zA-Z_-]+}", remoteDemandHandler},
		}

		for _, r := range routes {
			api.Methods(r.Methods...).Path(r.Pattern).Handler(r.HandlerFunc)
		}
	}

	srv := &HTTPd{
		Server: &http.Server{
			Addr:         url,
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			ErrorLog:     log.ERROR,
		},
	}
	srv.SetKeepAlivesEnabled(true)

	return srv
}

// Router returns the main router
func (s *HTTPd) Router() *mux.Router {
	return s.Handler.(*mux.Router)
}
