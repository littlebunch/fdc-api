# Dockerfile References: https://docs.docker.com/engine/reference/builder/
FROM golang:latest as builder
LABEL maintainer="Gary Moore <littlebunch@gmail.com>"

# Stage 1
WORKDIR /app
# Attached a volume for logging
ARG LOG_DIR=/app/logs
RUN mkdir -p ${LOG_DIR}
ENV LOG_FILE_LOCATION=${LOG_DIR}/fdcapi.log
#
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fdcapi ./api/main.go ./api/routes.go


#### Stage 2
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/fdcapi .
ADD api/dist/ ./dist
EXPOSE 8000
VOLUME [${LOG_DIR}]
CMD ["./fdcapi"]

