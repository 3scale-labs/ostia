# Based on https://shaneutt.com/blog/rust-fast-small-docker-image-builds/

# ------------------------------------------------------------------------------
# Build Stage
# ------------------------------------------------------------------------------

FROM rust:1.43 as limitador-build

RUN apt-get update \
 && apt-get install musl-tools -y

RUN rustup target add x86_64-unknown-linux-musl \
 && rustup component add rustfmt

# get Envoy protobufs
WORKDIR /usr/src/envoy-xds-grpc
RUN git clone https://github.com/envoyproxy/data-plane-api.git \
 && git clone https://github.com/googleapis/googleapis.git \
 && git clone https://github.com/envoyproxy/protoc-gen-validate.git \
 && git clone https://github.com/cncf/udpa.git

WORKDIR /usr/src/limitador

COPY ./Cargo.lock ./Cargo.lock
COPY ./Cargo.toml ./Cargo.toml

RUN mkdir src/

RUN echo "fn main() {println!(\"if you see this, the build broke\")}" > src/main.rs \
 && echo "fn main() {println!(\"if you see this, the build broke\")}" > src/server.rs

RUN RUSTFLAGS=-Clinker=musl-gcc cargo build --release --target=x86_64-unknown-linux-musl

# avoid downloading and compiling all the dependencies when there's a change in
# our code.
RUN rm -f target/x86_64-unknown-linux-musl/release/deps/limitador* \
 && rm -f target/x86_64-unknown-linux-musl/release/deps/ratelimit*

COPY . .

RUN RUSTFLAGS=-Clinker=musl-gcc cargo build --release --target=x86_64-unknown-linux-musl


# ------------------------------------------------------------------------------
# Run Stage
# ------------------------------------------------------------------------------

FROM alpine:3.11

RUN addgroup -g 1000 limitador \
 && adduser -D -s /bin/sh -u 1000 -G limitador limitador

WORKDIR /home/limitador/bin/

COPY --from=limitador-build /usr/src/limitador/examples/limits.yaml ../
COPY --from=limitador-build /usr/src/limitador/target/x86_64-unknown-linux-musl/release/ratelimit-server .

RUN chown limitador:limitador ratelimit-server

USER limitador

ENV LIMITS_FILE=/home/limitador/limits.yaml

CMD ["./ratelimit-server"]
