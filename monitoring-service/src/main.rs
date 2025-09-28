use std::env;
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use serde::{Serialize, Deserialize};
use serde_json;

#[derive(Serialize, Deserialize, Debug, Clone)]
struct Transaction {
    transaction_id: String,
    user_id: u32,
    amount: f64,
    currency: String,
    metadata: serde_json::Value,
    timestamp: String,
}

const SERVICE_NAME: &str = "monitoring-service";

// constants
const OK_RESPONSE: &str = "HTTP/1.1 200 OK\r\n\r\n";
const NOT_FOUND_RESPONSE: &str = "HTTP/1.1 404 NOT FOUND\r\n\r\n";
const INTERNAL_SERVER_ERROR_RESPONSE: &str = "HTTP/1.1 500 INTERNAL SERVER ERROR\r\n\r\n";

// store transactions in-memory (later you can move this to Redis or TigerBeetle)
static mut TRANSACTIONS: Vec<Transaction> = Vec::new();

fn main() {
    let port = env::var("MONITORING_SERVICE_PORT").unwrap_or("8082".to_string());
    let listener =
        TcpListener::bind(format!("localhost:{}", port)).expect("Could not bind to port");
    println!("{} listening on port {}", SERVICE_NAME, port);

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => handle_client(stream),
            Err(e) => eprintln!("Connection failed: {}", e),
        }
    }
}

fn request_with(request: &str, prefix: &str) -> bool {
    request.starts_with(prefix)
}

fn get_id(request: &str) -> &str {
    request.split("/").nth(2).unwrap_or_default().split_whitespace().next().unwrap_or_default()
}

fn get_transaction_from_request(request: &str) -> Option<Transaction> {
    let body = request.split("\r\n\r\n").nth(1)?;
    serde_json::from_str(body).ok()
}

fn handle_client(mut stream: TcpStream) {
    let mut buffer = [0; 1024];
    let mut request = String::new();

    match stream.read(&mut buffer) {
        Ok(size) => {
            request.push_str(String::from_utf8_lossy(&buffer[..size]).as_ref());

            let (status_line, content) = if request_with(&request, "POST /transactions") {
                handle_post_transaction(&request)
            } else if request_with(&request, "GET /transactions/") {
                handle_get_transaction(&request)
            } else {
                (NOT_FOUND_RESPONSE.to_string(), "Not Found".to_string())
            };

            stream.write_all(format!("{}{}", status_line, content).as_bytes()).unwrap();
        }
        Err(e) => {
            println!("Failed to read from connection: {}", e);
        }
    }
}

// POST /transactions
fn handle_post_transaction(request: &str) -> (String, String) {
    match get_transaction_from_request(request) {
        Some(tx) => {
            unsafe {
                TRANSACTIONS.push(tx.clone());
            }
            (OK_RESPONSE.to_string(), serde_json::to_string(&tx).unwrap())
        }
        None => (INTERNAL_SERVER_ERROR_RESPONSE.to_string(), "Invalid Transaction".to_string()),
    }
}

// GET /transactions/{id}
fn handle_get_transaction(request: &str) -> (String, String) {
    let id = get_id(request);
    unsafe {
        if let Some(tx) = TRANSACTIONS.iter().find(|t| t.transaction_id == id) {
            (OK_RESPONSE.to_string(), serde_json::to_string(tx).unwrap())
        } else {
            (NOT_FOUND_RESPONSE.to_string(), "Transaction Not Found".to_string())
        }
    }
}
