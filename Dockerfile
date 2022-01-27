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
