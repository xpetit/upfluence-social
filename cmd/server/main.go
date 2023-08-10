package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"time"

	social "github.com/xpetit/upfluence-social"
	"github.com/xpetit/upfluence-social/analysis"
)

type App struct {
	Addr        string
	EventStream *social.EventStream
}

func (a *App) Analyze(values url.Values) (map[string]int, error) {
	var (
		dimensionVal = values.Get("dimension")
		durationVal  = values.Get("duration")
	)
	if dimensionVal == "" {
		return nil, errors.New("missing dimension query parameter")
	}
	if durationVal == "" {
		return nil, errors.New("missing duration query parameter")
	}

	dimensionI := slices.Index(social.Dimensions[:], dimensionVal)
	if dimensionI == -1 {
		return nil, fmt.Errorf("unknown dimension, must be one of: %v", social.Dimensions)
	}
	dimension := social.Dimension(dimensionI)

	duration, err := time.ParseDuration(durationVal)
	if err != nil {
		return nil, err
	}
	if duration < 0 {
		return nil, errors.New("invalid duration")
	}

	stats := analysis.Gather(a.EventStream, duration, dimension)

	m := map[string]int{
		"total_posts":       stats.TotalPosts,
		"minimum_timestamp": stats.MinimumTimestamp,
		"maximum_timestamp": stats.MaximumTimestamp,
	}
	// TODO: is it really necessary to use the dimension as a key prefix in the returned object?
	m[dimensionVal+"_p50"] = stats.P50
	m[dimensionVal+"_p90"] = stats.P90
	m[dimensionVal+"_p99"] = stats.P99

	return m, nil
}

func (app *App) Run() error {
	eventStream, err := social.OpenEventStream("https://stream.upfluence.co/stream")
	if err != nil {
		return err
	}
	app.EventStream = eventStream

	http.Handle("/analysis", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		stats, err := app.Analyze(r.URL.Query())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(struct{ Error string }{err.Error()})
		} else {
			enc.Encode(stats)
		}
	}))

	log.Println("listening to", app.Addr)
	return http.ListenAndServe(app.Addr, nil)
}

func main() {
	addr := flag.String("addr", "localhost:8080", "network address to listen to")
	flag.Parse()

	app := App{Addr: *addr}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
