FROM alpine:3 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
ARG TARGETPLATFORM
# Copy the CA bundle so Telegram API requests succeed inside the container.
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY $TARGETPLATFORM/telegram-owl /usr/bin/
ENTRYPOINT ["/usr/bin/telegram-owl"]
