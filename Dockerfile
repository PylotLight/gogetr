FROM golang:1.20-alpine3.17 as builder
RUN mkdir /build && apk --no-cache add ca-certificates
ADD . /build/
WORKDIR /build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w -X 'main.AppVersion=v0.9.0' -extldflags '-static'" -o main .
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.AppVersion=`date -u +.%Y%m%d.%H%M%S` -extldflags '-static'" -o main .

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

# docker build -t light37/gogetr:latest -f \\wsl.localhost\Debian\home\light\goget\Dockerfile \\wsl.localhost\Debian\home\light\goget
# sudo docker build -t light37/gogetr:latest -f Dockerfile .
# sudo docker push light37/gogetr:latest
#sudo docker build -t light37/gogetr:latest -f Dockerfile . ; sudo docker push light37/gogetr:latest ; krr gogetr


# FROM golang:alpine as build # Redundant, current golang images already include ca-certificates 
# RUN apk --no-cache add ca-certificates 
# WORKDIR /go/src/app 
# COPY . . 
# RUN CGO_ENABLED=0 go-wrapper install -ldflags '-extldflags "-static"'  
# FROM scratch # copy the ca-certificate.crt from the build stage 
# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ 
# COPY --from=build /go/bin/app /app 
# ENTRYPOINT ["/app"] 