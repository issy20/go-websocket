ARG VARIANT="1.18-bullseye"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

ENV PATH /go/bin:$PATH
RUN ["chmod", "+w", "/go/src"]