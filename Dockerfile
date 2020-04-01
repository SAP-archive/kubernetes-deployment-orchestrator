FROM alpine

WORKDIR /app
ENV HOME=/app
COPY * /usr/bin/

ENTRYPOINT ["/usr/bin/shalm","controller"]
