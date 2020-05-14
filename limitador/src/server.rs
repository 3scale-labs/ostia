use crate::envoy::service::ratelimit::v2::rate_limit_service_server::{
    RateLimitService, RateLimitServiceServer,
};
use crate::envoy::service::ratelimit::v2::{RateLimitRequest, RateLimitResponse};
use limitador::limit::Limit;
use limitador::RateLimiter;
use std::collections::HashMap;
use std::env;
use std::sync::Arc;
use tonic::{transport::Server, Request, Response, Status};

const LIMITS_FILE_ENV: &str = "LIMITS_FILE";

include!("envoy.rs");

pub struct MyRateLimiter {
    limiter: Arc<tokio::sync::Mutex<RateLimiter>>,
}

impl MyRateLimiter {
    pub fn new() -> MyRateLimiter {
        match env::var(LIMITS_FILE_ENV) {
            Ok(val) => {
                let f = std::fs::File::open(val).unwrap();
                let limits: Vec<Limit> = serde_yaml::from_reader(f).unwrap();

                let mut rate_limiter = RateLimiter::new();
                for limit in limits {
                    rate_limiter.add_limit(limit);
                }

                MyRateLimiter {
                    limiter: Arc::new(tokio::sync::Mutex::new(rate_limiter)),
                }
            }
            _ => panic!("LIMITS_FILE env not set"),
        }
    }
}

impl Default for MyRateLimiter {
    fn default() -> Self {
        Self::new()
    }
}

#[tonic::async_trait]
impl RateLimitService for MyRateLimiter {
    async fn should_rate_limit(
        &self,
        request: Request<RateLimitRequest>,
    ) -> Result<Response<RateLimitResponse>, Status> {
        let mut values: HashMap<String, String> = HashMap::new();
        let req = request.into_inner();
        let domain = req.domain.to_string();

        // TODO: return error if domain is "".

        values.insert("namespace".to_string(), domain);

        // TODO: assume one descriptor for now.
        for entry in &req.descriptors[0].entries {
            values.insert(entry.key.clone(), entry.value.clone());
        }

        // TODO: lock rate_limited and update_counters together.
        // Although some users might decide to call them separately.

        let rate_limited = self.limiter.lock().await.is_rate_limited(&values);

        self.limiter.lock().await.update_counters(&values).unwrap();

        let overall_code = if rate_limited.unwrap() { 2 } else { 1 };

        let reply = RateLimitResponse {
            overall_code,
            statuses: vec![],
            headers: vec![],
            request_headers_to_add: vec![],
        };

        Ok(Response::new(reply))
    }
}

/// This function will get called on each inbound request, if a `Status`
/// is returned, it will cancel the request and return that status to the
/// client.
fn intercept(req: Request<()>) -> Result<Request<()>, Status> {
    println!("Intercepting request: {:?}", req);
    Ok(req)
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let host = env::var("HOST").unwrap_or_else(|_| String::from("[::1]"));
    let port = env::var("PORT").unwrap_or_else(|_| String::from("50052"));

    let addr = format!("{host}:{port}", host = host, port = port).parse()?;

    println!("Listening on {}", addr);

    let rate_limiter = MyRateLimiter::default();
    let svc = RateLimitServiceServer::with_interceptor(rate_limiter, intercept);

    Server::builder()
        .add_service(svc)
        .serve(addr)
        .await?;

    Ok(())
}
