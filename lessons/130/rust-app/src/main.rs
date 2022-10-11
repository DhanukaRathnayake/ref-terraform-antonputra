use actix_web::{get, post, web, App, HttpResponse, HttpServer};
use argon2::{self, Config, ThreadMode, Variant, Version};
use lazy_static::lazy_static;
use prometheus::register_histogram;
use prometheus::{self, Encoder, Histogram, TextEncoder};
use rand::RngCore;
use serde::{Deserialize, Serialize};
use std::time::Instant;
use tokio;
use tokio_postgres::{Client, Error, NoTls};

const SALT_LENGTH: usize = 16;

lazy_static! {
    static ref GENERATE_HASH_HISTOGRAM: Histogram = register_histogram!(
        "generate_hash_duration_seconds",
        "Duration to generate argon2 hash for the user.",
        vec![0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.11, 0.12, 0.13, 0.14, 0.15]
    )
    .unwrap();
    static ref SAVE_USER_HISTOGRAM: Histogram = register_histogram!(
        "save_user_duration_seconds",
        "Duration to save user into the database.",
        vec![0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1]
    )
    .unwrap();
}

#[derive(Debug, Serialize, Deserialize)]
struct User {
    email: String,
    password: String,
}

#[post("/users")]
async fn create_user(user: web::Json<User>) -> HttpResponse {
    let config = Config {
        variant: Variant::Argon2id,
        version: Version::Version13,
        mem_cost: 64 * 1024,
        time_cost: 3,
        lanes: 2,
        thread_mode: ThreadMode::Parallel,
        secret: &[],
        ad: &[],
        hash_length: 32,
    };

    let encoded_hash = generate_from_password(user.0.password, config);
    save_user(user.0.email, encoded_hash).await.unwrap();

    HttpResponse::Created().body("User created.")
}

#[get("/metrics")]
async fn metrics() -> HttpResponse {
    let encoder = TextEncoder::new();
    let mut buffer = vec![];
    encoder
        .encode(&prometheus::gather(), &mut buffer)
        .expect("Failed to encode metrics");

    let response = String::from_utf8(buffer.clone()).expect("Failed to convert bytes to string");
    buffer.clear();
    HttpResponse::Ok().body(response)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| App::new().service(create_user).service(metrics))
        .bind(("0.0.0.0", 8080))?
        .run()
        .await
}

fn generate_from_password(password: String, config: Config) -> String {
    let start = Instant::now();
    let salt = generate_random_bytes();

    let hash = argon2::hash_encoded(password.as_bytes(), &salt, &config).unwrap();
    GENERATE_HASH_HISTOGRAM.observe(start.elapsed().as_secs_f64());
    hash
}

fn generate_random_bytes() -> [u8; SALT_LENGTH] {
    let mut data = [0u8; SALT_LENGTH];
    rand::thread_rng().fill_bytes(&mut data);
    data
}

async fn save_user(email: String, hash: String, client: Client) -> Result<(), Error> {
    let start = Instant::now();
    // let (client, connection) = tokio_postgres::connect(
    //     "user=app password=devops123 dbname=lesson_130 host=postgres",
    //     NoTls,
    // )
    // .await?;

    // tokio::spawn(async move {
    //     if let Err(e) = connection.await {
    //         eprintln!("connection error: {}", e);
    //     }
    // });

    client
        .execute(
            "INSERT INTO rust_users(email, password_hash) VALUES($1, $2)",
            &[&email, &hash],
        )
        .await?;

    SAVE_USER_HISTOGRAM.observe(start.elapsed().as_secs_f64());

    Ok(())
}

async fn init_db() -> Result<(), Error> {
    let (client, connection) = tokio_postgres::connect(
        "user=app password=devops123 dbname=lesson_130 host=postgres",
        NoTls,
    )
    .await?;

    tokio::spawn(async move {
        if let Err(e) = connection.await {
            eprintln!("connection error: {}", e);
        }
    });
    Ok(())
}
