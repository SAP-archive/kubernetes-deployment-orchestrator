FROM alpine

WORKDIR /app
ENV HOME=/app
COPY kapp /usr/bin/
COPY kubectl /usr/bin/
COPY kdo /usr/bin/

ENTRYPOINT ["/usr/bin/kdo","controller"]
