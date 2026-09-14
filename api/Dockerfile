FROM golang as builder

ARG GIT_COMMIT
ENV GIT_COMMIT=$GIT_COMMIT

WORKDIR /http-sms

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go get github.com/swaggo/swag/gen@latest
RUN go get github.com/swaggo/swag/cmd/swag@latest
RUN go install github.com/swaggo/swag/cmd/swag
RUN swag init --requiredByDefault --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$GIT_COMMIT" -o /bin/http-sms .

FROM alpine:latest

RUN addgroup -S http-sms && adduser -S http-sms -G http-sms \
    && mkdir -p /app/logs && chown -R http-sms:http-sms /app/logs

USER http-sms
WORKDIR /home/http-sms

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
COPY --from=builder /bin/http-sms ./
COPY --from=builder /http-sms/root.crt ./

ENV ZONEINFO=/zoneinfo.zip

EXPOSE 8000

ENTRYPOINT ["./http-sms", "--dotenv=false"]
