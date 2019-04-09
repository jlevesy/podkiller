FROM golang:1.12 AS build
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -ldflags='-s -w' ./cmd/podkiller

FROM scratch AS runtime
COPY --from=build /app/podkiller /podkiller
COPY --from=build /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/podkiller"]
