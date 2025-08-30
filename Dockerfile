FROM golang:1.24

WORKDIR /usr/src

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

# RUN apt-get update && apt-get install -y ffmpeg && apt-get clean

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/app
EXPOSE 8000

CMD ["app"]
