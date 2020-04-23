FROM envoyproxy/envoy-dev:latest

RUN apt-get update && apt-get -q install -y \
    curl
CMD /usr/local/bin/envoy -c /etc/envoy.yaml --service-cluster front-proxy -l ${LOG_LEVEL}
