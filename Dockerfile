FROM golang:1.24.2

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /under-test

ENV PORT=8080
ENV HOST=0.0.0.0
EXPOSE ${PORT}

CMD ["/under-test"]

