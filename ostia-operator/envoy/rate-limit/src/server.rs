use tonic::{transport::Server, Request, Response, Status};

use envoy::service::ratelimit::v2::rate_limit_service_server::{RateLimitService, RateLimitServiceServer};
use envoy::service::ratelimit::v2::{RateLimitRequest, RateLimitResponse};
use crate::envoy::api::v2::ratelimit::{{RateLimitDescriptor, rate_limit_descriptor::Entry}};

include!("envoy.rs");

#[derive(Debug, Default)]
pub struct MyRateLimitService {}

fn find_descriptor_entry<'a>(descriptor: &'a RateLimitDescriptor, key: &str) -> Option<&'a Entry> {
    for entry in &descriptor.entries {
        if entry.key == key {
            return Some(entry)
        }
    }
    None
}

fn find_descriptors_entry<'a>(request: &'a RateLimitRequest, key: &str) -> Option<&'a Entry> {
    for desc in &request.descriptors {
        let entry = find_descriptor_entry(desc, key);

        println!("Got descriptor: {:?}", desc);

        match entry {
            Some(e) => return Some(e),
            _ => {}
        };
    }
    None
}

#[tonic::async_trait]
impl RateLimitService for MyRateLimitService {
    async fn should_rate_limit(
        &self,
        request: Request<RateLimitRequest>,
    ) -> Result<Response<RateLimitResponse>, Status> {
        println!("Got a request: {:?}", request);

        let msg = request.get_ref();

        let entry = find_descriptors_entry(msg, "ratelimitkey");
        println!("Got descriptor entry: {:?}",entry );

        let reply = envoy::service::ratelimit::v2::RateLimitResponse {
            headers: (&[]).to_vec(),
            request_headers_to_add: (&[]).to_vec(),
            statuses: (&[]).to_vec(),
            overall_code: envoy::service::ratelimit::v2::rate_limit_response::Code::Ok as i32,
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
    let addr = "0.0.0.0:50052".parse()?;
    let rate_limiter = MyRateLimitService::default();

    let svc = RateLimitServiceServer::with_interceptor(rate_limiter, intercept);

    Server::builder()
        .add_service(svc)
        .serve(addr)
        .await?;

    Ok(())
}
