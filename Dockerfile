FROM alpine:3.20 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
ARG TARGETPLATFORM
# copy CA bundle so HTTPS requests (e.g. Telegram API) succeed inside the container
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY $TARGETPLATFORM/telegram-owl /usr/bin/
ENTRYPOINT ["/usr/bin/telegram-owl"]
