package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/cgilling/plantuml-proxy/plantuml"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v1"
)

const (
	DefaultListenAddr   = ":8080"
	EnvPrefix           = "PLANTUML_PROXY"
	ModifiedTableFormat = `This application can be configured via the environment. The following environment
variables can be used:
KEY	TYPE	DESCRIPTION
{{range .}}{{usage_key .}}	{{usage_type .}}	{{usage_description .}}
{{end}}`
)

var (
	configPath         = flag.String("config", "", "path to yaml config file, if not provided, env vars are used")
	printDefaultConfig = flag.Bool("print-default-config", false, "if set to true, will print the default config yaml, end exit")
)

type Config struct {
	ListenAddr string `yaml:"listen_addr" json:"listen_addr" envconfig:"listen_addr" desc:"address to listen on for requests"`
	ServerURL  string `yaml:"plantuml_url" json:"plantuml_url" envconfig:"plantuml_url" desc:"url for plantuml server to handle conversion requests"`
}

type App struct {
	config Config
	client *plantuml.Client
}

func main() {
	var config Config
	applyConfigDefault(&config)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %s
plantuml-proxy is a simple proxy service that allows for posting of
unencoded uml files for conversion rather than needing to use the custom
plantuml url encoding scheme.
Flags:
`, os.Args[0])
		flag.PrintDefaults()
		fmt.Println("")
		tabs := tabwriter.NewWriter(os.Stderr, 1, 0, 4, ' ', 0)
		envconfig.Usagef(EnvPrefix, &config, tabs, ModifiedTableFormat)
		tabs.Flush()
	}

	flag.Parse()

	if *printDefaultConfig {
		b, _ := yaml.Marshal(&config)
		fmt.Print(string(b))
		return
	}
	if *configPath != "" {
		b, err := ioutil.ReadFile(*configPath)
		if err != nil {
			log.Fatalf("failed to read config file %q: %v", *configPath, err)
		}
		if err = yaml.Unmarshal(b, &config); err != nil {
			log.Fatalf("failed to parse config yaml: %v", err)
		}
	}
	if err := envconfig.Process(EnvPrefix, &config); err != nil {
		log.Fatalf("failed to process environment: %v", err)
	}

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
