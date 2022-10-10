Web Frameworks Benchmark: https://web-frameworks-benchmark.netlify.app/result?l=rust
https://github.com/flosse/rust-web-framework-comparison

rust-web-frameworks-benchmark: https://github.com/rousan/rust-web-frameworks-benchmark


Hashes

MD, SHA, Bcrypt, Scrypt and Argon
Argon2 - holy grail
Argon2i employs an isolated memory access, which is best for password storage. 


Prometheus
https://romankudryashov.com/blog/2021/11/monitoring-rust-web-application/
https://blog.logrocket.com/using-prometheus-metrics-in-a-rust-web-service/





psql -h localhost -p 5432 -U postgres -d postgres

CREATE USER app WITH PASSWORD 'devops123';
CREATE DATABASE lesson_130;
GRANT ALL PRIVILEGES ON DATABASE lesson_130 TO app;

CREATE TABLE rust_users (
    user_id SERIAL primary key,
    email varchar(254) NOT NULL,
    password_hash varchar(254) NOT NULL
    );

CREATE TABLE go_users (
    user_id SERIAL primary key,
    email varchar(254) NOT NULL,
    password_hash varchar(254) NOT NULL
    );


INSERT INTO users(email, password_hash)
VALUES ('test@gmail.com', 'ajhgd7623fd23gf');

curl -d '{"email": "hafgd@skhgf.com"}' http://localhost:8080/users


https://github.com/goriunov/actix-tokio-postgres-example/blob/master/src/main.rs








gcc -Wall -o main main.c

https://docs.aws.amazon.com/lambda/latest/dg/runtimes-walkthrough.html