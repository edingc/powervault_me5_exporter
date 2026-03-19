# build stage
FROM golang:1.25.7-alpine3.23 AS builder

ARG VERSION=v1.0.0
ARG BRANCH=unknown
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s \
      -X github.com/prometheus/common/version.Version=${VERSION} \
      -X github.com/prometheus/common/version.Revision=${COMMIT} \
      -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE} \
      -X github.com/prometheus/common/version.Branch=${BRANCH}" \
    -o /prometheus-powervault-me5-exporter \
    ./cmd/powervault_me5_exporter

# final stage 
FROM alpine:3.23

RUN apk --no-cache add ca-certificates gcompat

COPY --from=builder /prometheus-powervault-me5-exporter /prometheus-powervault-me5-exporter

EXPOSE 9850

ENTRYPOINT ["/prometheus-powervault-me5-exporter"]
