FROM rust:1.43

WORKDIR /usr/src/limitador

RUN rustup component add rustfmt

# copy over your manifests
COPY ./Cargo.lock ./Cargo.lock
COPY ./Cargo.toml ./Cargo.toml

RUN mkdir .cargo
RUN cargo vendor --locked > .cargo/config

COPY . .

RUN cargo install --verbose --path .

ENV LIMITS_FILE=./examples/limits.yaml

CMD ["ratelimit-server"]
