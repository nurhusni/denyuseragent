package traefik_plugin_example

import (
	"context"
	"fmt"
	"log"
	"net/http"

	agent "github.com/mileusna/useragent"
)

// Config defines the plugin dynamic configuration.
type Config struct {
	UserAgents []string
}

// CreateConfig creates a new config.
func CreateConfig() *Config {
	return &Config{}
}

// Plugin is the traefik plugin implementation.
type Plugin struct {
	next        http.Handler
	name        string
	knownAgents map[string]struct{}
}

// New creates plugin handler & return plugin instance
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {

	if cfg == nil {
		return nil, fmt.Errorf("no UserAgent config provided")
	}

	knownAgents := map[string]struct{}{}
	for _, ka := range cfg.UserAgents {
		knownAgents[ka] = struct{}{}
	}

	return &Plugin{
		next:        next,
		name:        name,
		knownAgents: knownAgents,
	}, nil

}

// ServeHTTP implements http.Handler interface
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	getAgent := agent.Parse(req.Header.Get("User-Agent"))

	if _, blocked := p.knownAgents[getAgent.Name]; blocked {
		log.Printf("%s : %s - access denied - user agent is blocked: %s", p.name, req.URL.String(), getAgent.Name)
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	p.next.ServeHTTP(rw, req)

}
