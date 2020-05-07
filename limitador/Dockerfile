FROM rust:1.43

RUN rustup component add rustfmt

# copy Envoy protobufs
WORKDIR /usr/src/ostia/envoy-xds-grpc
COPY ./envoy-xds-grpc/ .

WORKDIR /usr/src/ostia/limitador

# copy over your manifests
COPY ./limitador/Cargo.lock ./Cargo.lock
COPY ./limitador/Cargo.toml ./Cargo.toml

RUN mkdir .cargo
RUN cargo vendor --locked > .cargo/config

COPY ./limitador .

RUN cargo install --verbose --path .

ENV LIMITS_FILE=./examples/limits.yaml

CMD ["ratelimit-server"]
