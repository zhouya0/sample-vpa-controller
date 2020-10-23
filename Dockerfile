FROM debian:stretch-slim

COPY vpa /usr/local/bin

CMD ["/usr/local/bin/vpa"]