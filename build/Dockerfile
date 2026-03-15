FROM golang:1.26.0 AS builder

WORKDIR /app

COPY . ./

RUN go build -o bin/companies cmd/companies/main.go

# -

FROM debian:stable-slim

COPY --from=builder /app/bin/companies /companies

CMD [ "/companies" ]