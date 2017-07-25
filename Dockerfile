FROM golang:alpine
MAINTAINER Leonardo Javier Gago <ljgago@gmail.com>

RUN apk update && apk add git ffmpeg ca-certificates && update-ca-certificates

RUN CGO_ENABLED=0 go get github.com/ljgago/MusicBot

RUN mkdir /bot

WORKDIR /bot

CMD ["MusicBot", "-f", "bot.toml"]
