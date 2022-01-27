# traefik-plugin-example

This example repository contains a simple traefik plugin for blocking request from certain user agent, using [mileusna/useragent](https://github.com/mileusna/useragent) Go library.
Plugin origin from [kucjac/traefik-block-ua](kucjac/traefik-block-ua)

This simple documentation will give an explanatory on how a traefik plugin structure and deploy it using custom Docker images

## Traefik Plugin Structure

Traefik Plugin is a middleware that provides `http.Handler` to perform specific processing of requests and responses. This plugin is not compiled, but interpreted using [Traefik Yaegi](https://github.com/traefik/yaegi) interpreter.

Structure of a Traefik plugin consists of these several Go objects :

- A configuration struct : `type Config struct { ... }`. Holds the necessary data needs to be processed by the plugn.
- A function to initialize plugin configuration `func CreateConfig() *Config`
- A function to instantiated the plugin `func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error)`
- `http.Handler` implementation

```golang
// Package example (an example plugin).
package example

import (
    "context"
    "net/http"
)

// Config the plugin configuration.
type Config struct {
    // ...
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
    return &Config{
        // ...
    }
}

// Example a plugin.
type Example struct {
    next     http.Handler
    name     string
    // ...
}

// New created a new plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
    // ...
    return &Example{
        // ...
    }, nil
}

func (e *Example) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    // ...
    e.next.ServeHTTP(rw, req)
}
```

## Plugin Dependencies

If your plugin contains dependencies outside standard Go packages, then they must be [vendored](https://golang.org/ref/mod#vendoring).
Run `go mod vendor` on your plugin repo to prepare them for usage

## Public Plugin

Plugin provided for public use at Github, should obey these checklist

- The traefik-plugin topic must be set on your repository.
- There must be a .traefik.yml file at the root of your project describing your plugin, and it must have a valid testData property for testing purposes.
- There must be a valid go.mod file at the root of your project.
- Your plugin must be versioned with a git tag.
- If you have package dependencies, they must be vendored and added to your GitHub repository.

They will be included on Traefik Pilot ecosystem

## Dockerization of Custom or Private Traefik Plugin

### Local Plugin Structure

Plugin that will go into your Docker image must follow this structure and place in `plugins-local` directory. This directory is relative to the running traefik binary on your Docker

```bash
./plugins-local/
    └── src
        └── github.com
            └── koinworks
                └── traefik-plugin-example
                    ├── plugin.go
                    ├── plugin_test.go
                    ├── go.mod
                    ├── vendor/
                    ├── Makefile
                    └── readme.md
```

### Dockerization of Custom Plugin

- Set build environment variable for your docker build. Adjust them with your environment setting
  
  ```bash
  export DOCKER_IMAGE=traefik-plugin-demo
  ```

- Create Dockerfile for your plugin
  
  ```bash
  FROM alpine:latest as build
  ARG PLUGIN_MODULE=github.com/traefik/plugindemo
  ARG PLUGIN_GIT_REPO=https://github.com/traefik/plugindemo.git
  ARG PLUGIN_GIT_BRANCH=master
  RUN apk update && \
    apk add git && \
    git clone ${PLUGIN_GIT_REPO} /plugins-local/src/${PLUGIN_MODULE} \
      --depth 1 --single-branch --branch ${PLUGIN_GIT_BRANCH}

  FROM traefik:v2.5
  COPY --from=build /plugins-local /plugins-local
  ```

- Build the image
  
  ```bash
  docker build -f Dockerfile -t ${DOCKER_IMAGE} .
  ```
### Dockerization of Private Traefik Plugin

- Set build environment variable for your docker build and adjust them
  
  ```bash
  export DOCKER_IMAGE=traefik-plugin-example
  export PLUGIN_MODULE=github.com/koinworks/traefik-plugin-example
  export PLUGIN_GIT_REPO=https://github.com/koinworks/traefik-plugin-example.git
  export PLUGIN_GIT_BRANCH=master
  ```

- Create Dockerfile for your plugin

  ```bash
  FROM alpine:latest as build

  # Clone your plugin git repositories:
  ARG PLUGIN_MODULE=github.com/traefik/plugindemo
  ARG PLUGIN_GIT_REPO=git@github.com:traefik/plugindemo.git
  ARG PLUGIN_GIT_BRANCH=master
  RUN apk add --update git openssh && \
    mkdir -m 700 /root/.ssh && \
    touch -m 600 /root/.ssh/known_hosts && \
    ssh-keyscan github.com > /root/.ssh/known_hosts
  RUN --mount=type=ssh git clone \
    --depth 1 --single-branch --branch ${PLUGIN_GIT_BRANCH} \
    ${PLUGIN_GIT_REPO} /plugins-local/src/${PLUGIN_MODULE} 
    
  FROM traefik:v2.5
  COPY --from=build /plugins-local /plugins-local
  ```

- Build the image
  ```bash
  docker build -f Dockerfile.private \
  --ssh default --tag ${DOCKER_IMAGE} \
  --build-arg PLUGIN_MODULE \
  --build-arg PLUGIN_GIT_REPO \
  --build-arg PLUGIN_GIT_BRANCH .
  ```

- Push the image into your private docker repository

### Plugin Usage

In this example, `traefik-plugin-example` will be loaded from the path `./plugins-local/src/github.com/koinworks/traefik-plugin-example.
This example of embedded configuration will show you on how to load the plugin "

```yaml

http:
  routers:
    my-router:
      rule: host(`midgard.localhost`)
      service: service-midgard
      entryPoints:
        - web
      middlewares:
        - plugin-example

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    plugin-example:
      plugin:
        traefik-plugin-example:
          userAgent:
            - Mozilla/5.0
```

## References

1) https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/
2) https://doc.traefik.io/traefik-pilot/plugins/plugin-dev/
