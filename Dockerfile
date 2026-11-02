FROM alpine:latest

RUN apk add --no-cache ca-certificates \
	&& adduser -D -h /var/lib/goshare -s /sbin/nologin goshare \
	&& mkdir -p /var/lib/goshare/data \
	&& chown -R goshare:goshare /var/lib/goshare

COPY goshare /usr/local/bin/goshare

EXPOSE 8080

ENV GOSHARE_SERVER_PORT=8080 \
	GOSHARE_STORAGE_DIR=/var/lib/goshare/data \
	GOSHARE_MAX_UPLOAD_SIZE=104857600 \
	GOSHARE_AUTH_ENABLED=false \
	GOSHARE_AUTH_USERNAME=admin \
	GOSHARE_AUTH_PASSWORD=password

USER goshare
WORKDIR /var/lib/goshare

VOLUME ["/var/lib/goshare/data"]

ENTRYPOINT ["/usr/local/bin/goshare"]
