# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23 as build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/telegrambot ./cmd/telegrambot

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/telegrambot /
COPY ./image /var/labyrinth/image
ENTRYPOINT ["/telegrambot"]