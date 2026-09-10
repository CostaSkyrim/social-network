# NOTE: the build context is the repository root (see docker-compose.yml),
# so the COPY paths below are relative to the repo root, not to backend/.
FROM golang:1.24-alpine AS build

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.24-alpine

RUN apk add --no-cache gcc musl-dev sqlite-dev

RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app
COPY --from=build /go/pkg/mod /go/pkg/mod
COPY . .

# Use Docker-specific config
RUN cp backend/configs.docker.json backend/configs.json

ENTRYPOINT ["air"]
