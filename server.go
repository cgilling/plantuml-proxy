package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/cgilling/plantuml-proxy/plantuml"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
)

const DefaultListenAddr = ":8080"

type Config struct {
	ListenAddr string
	ServerURL  string
}

type App struct {
	config Config
	client *plantuml.Client
}

func main() {
	var config Config
	applyConfigDefault(&config)

	app, err := NewApp(config)
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.POST("/:format", app.ProxyPOST)

	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, router)))
}

func NewApp(config Config) (*App, error) {
	client, err := plantuml.NewClient(plantuml.ClientConfig{
		Doer: http.DefaultClient,
		URL:  config.ServerURL,
	})
	if err != nil {
		return nil, err
	}
	return &App{
		config: config,
		client: client,
	}, nil
}

func applyConfigDefault(config *Config) {
	if config.ListenAddr == "" {
		config.ListenAddr = DefaultListenAddr
	}
	if config.ServerURL == "" {
		config.ServerURL = plantuml.DefaultClientURL
	}
}

func (app *App) ProxyPOST(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	format := ps.ByName("format")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	b, err := app.client.Convert(body, format)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.Write(b)
}
