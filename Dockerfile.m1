FROM golang:1.16-stretch as with-mem-mongo

RUN wget https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-debian92-4.4.0.tgz
RUN tar xzvf mongodb-linux-x86_64-debian92-4.4.0.tgz
RUN mkdir -p /root/.cache/dp-mongodb-in-memory/mongodb-linux-x86_64-debian92-4.4.0.tgz/
RUN mv mongodb-linux-x86_64-debian92-4.4.0/bin/mongod /root/.cache/dp-mongodb-in-memory/mongodb-linux-x86_64-debian92-4.4.0.tgz/

FROM with-mem-mongo as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

# TEST
FROM build as test

