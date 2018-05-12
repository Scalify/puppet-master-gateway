FROM scalify/glide:0.13.0 as builder
WORKDIR /go/src/gitlab.com/scalifyme/puppet-master/puppet-master/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/puppet-master .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/gitlab.com/scalifyme/puppet-master/puppet-master/bin/puppet-master .
RUN chmod +x puppet-master
ENTRYPOINT ["./puppet-master"]
