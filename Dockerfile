# NOTE: this file runs the app using the production configuration values
# Do not run Docker on this locally until this is fixed.
FROM golang:1.7.3

WORKDIR /go/src/github.com/chloearianne/protestpulse
ADD . /go/src/github.com/chloearianne/protestpulse

# Manually set the time zone
ENV TZ=America/Los_Angeles
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN go install github.com/chloearianne/protestpulse

ENTRYPOINT ["/go/bin/protestpulse"]

EXPOSE 8080
