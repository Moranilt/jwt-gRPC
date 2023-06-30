package http_transport

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Moranilt/jwt-http2/config"
	"github.com/Moranilt/jwt-http2/logger"
	"github.com/gorilla/mux"
)

func New(addr string, log *logger.Logger, cfg *config.Config, consulKey string) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/watch", MakeWatchHandler(log, cfg, consulKey)).Methods(http.MethodPost)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	server.Handler = router
	return server
}

func MakeWatchHandler(log *logger.Logger, cfg *config.Config, consulKey string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			return
		}

		var cb []config.WatchConsulBody
		err = json.Unmarshal(b, &cb)
		if err != nil {
			log.Error(err)
			return
		}

		err = cfg.WatchConsul(r.Context(), consulKey, cb)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	})
}
