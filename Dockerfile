FROM docker.io/golang:1.20-bullseye AS build
LABEL authors="QianheYu"

WORKDIR /src
COPY . .

RUN apt-get update && apt-get install -y upx git

RUN make build && upx -9 bin/headscale-panel

FROM docker.io/debian:bullseye-slim
LABEL authors="QianheYu"
LABEL all-in-one=true

RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir -p /etc/headscale-panel && mkdir -p /etc/headscale && mkdir -p /var/lib/headscale && mkdir -p /var/run/headscale

COPY --from=build /src/bin/headscale-panel /bin/headscale-panel
ENV TZ UTC

CMD ["headscale-panel"]
