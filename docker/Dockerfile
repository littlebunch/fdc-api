FROM golang:alpine 

WORKDIR $GOPATH/src/github.com/littlebunch/gnut-bfpd-api/api

ENV PORT=8000

#RUN go build -o bfpd main.go routes.go

CMD ["go", "run", "main.go routes.go","-c ../config.yml"]

