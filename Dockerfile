FROM golang:1.19-alpine3.17 as builder

WORKDIR /app

COPY go.mod /app
COPY go.sum /app

RUN go mod download

COPY main.go /app
COPY util /app/util
COPY transforms /app/transforms

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/jq-pilot

FROM node:18-alpine as frontend

WORKDIR /app

COPY ./frontend/package.json /app/package.json

RUN npm install

COPY ./frontend/public /app/public
COPY ./frontend/src /app/src

RUN REACT_APP_ENV=production npm run build

from node:18-alpine as final

COPY --from=frontend /app/build /app/build
COPY --from=frontend /app/node_modules /app/node_modules
COPY --from=builder /app/jq-pilot /app/jq-pilot

EXPOSE 8000

ENTRYPOINT ["/app/jq-pilot", "--MODE=prod"]
