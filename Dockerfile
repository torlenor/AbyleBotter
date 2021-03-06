# STEP 1: Used to get SSL root certificates
FROM alpine:3.8 as builder

# Install SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates

# STEP 2: Create the redseligg image
FROM scratch

# Copy SSL root certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY bin/* /usr/bin/

CMD ["/usr/bin/botterinstance"]

LABEL org.label-schema.vendor="Torlenor" \
      org.label-schema.url="https://github.com/torlenor/redseligg" \
      org.label-schema.name="Redseligg" \
      org.label-schema.description="An extensible multi-platform chat bot for Discord, Matrix, Mattermost and Slack written in GO"

