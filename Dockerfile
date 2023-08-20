FROM golang:1.20.6-alpine3.17 as builder
RUN mkdir /build && apk --no-cache add ca-certificates
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w -X 'main.AppVersion=v1.3`date +.%Y%m%d`' -extldflags '-static'" -o main .

FROM scratch
COPY --from=builder /build/main /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ 
WORKDIR /app
ENV APIKey=""
ENV ImportFolder="/"
ENV ExportFolder="/"
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip 
ENV ZONEINFO /zoneinfo.zip
EXPOSE 9000
CMD ["./main"]

# sudo docker build -t light37/gogetr:latest -t light37/gogetr:$(date +%Y%m%d) -f Dockerfile .
#sudo docker build -t light37/gogetr:latest -f Dockerfile . ; sudo docker push light37/gogetr:latest ; krr gogetr