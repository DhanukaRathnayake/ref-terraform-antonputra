FROM rust:1

COPY ./src ./src
COPY ./Cargo.lock ./Cargo.lock
COPY ./Cargo.toml ./Cargo.toml

RUN cargo build --release

EXPOSE 8080

CMD ["./target/release/rust-app"]
