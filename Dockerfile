FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY main.go /app
COPY util /app/util
COPY transforms /app/transforms

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/jq-pilot

FROM node:22-alpine AS frontend

# pass in --build-arg ENV=local (or any value that's not production) to run this locally.
ARG ENV=production

WORKDIR /app

COPY ./frontend/package.json ./frontend/package-lock.json /app/

# Full install (no --omit=dev): the build toolchain (vite) lives in devDependencies.
RUN npm ci

COPY ./frontend/index.html /app/index.html
COPY ./frontend/vite.config.js /app/vite.config.js
COPY ./frontend/public /app/public
COPY ./frontend/src /app/src

RUN VITE_ENV=$ENV npm run build

FROM node:22-alpine AS final

WORKDIR /app

COPY --from=frontend /app/build /app/build
COPY --from=builder /app/jq-pilot /app/jq-pilot

EXPOSE 8000

ENTRYPOINT ["/app/jq-pilot", "--MODE=prod"]
