# Build every binary, then copy them into a small runtime image.
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 go build -o /out/email_sender ./cmd/worker/email_sender \
 && CGO_ENABLED=0 go build -o /out/push_notifications ./cmd/worker/push_notifications \
 && CGO_ENABLED=0 go build -o /out/event_dispatcher ./cmd/worker/event_dispatcher

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/ /app/
COPY configs ./configs
COPY migrations ./migrations
COPY locales ./locales
COPY templates ./templates
EXPOSE 8080 9090
ENTRYPOINT ["/app/server"]
