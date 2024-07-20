FROM golang:1.22.5 AS build
WORKDIR /app
COPY go.mod go.sum ./
COPY ./vendor ./vendor
COPY ./*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /avina

# Run the tests in the container
FROM build AS run-test
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build /avina /avina

USER nonroot:nonroot

ENTRYPOINT ["/avina"]
