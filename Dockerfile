FROM alpine

WORKDIR /app
ENV HOME=/app
COPY kapp /usr/bin/
COPY kubectl /usr/bin/
COPY ytt /usr/bin/
COPY shalm /usr/bin/

ENTRYPOINT ["/usr/bin/shalm","controller"]
