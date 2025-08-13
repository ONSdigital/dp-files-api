FROM golang:1.24.5-bullseye as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

# TEST
FROM build as test

