use crate::probe_to_rust::{startProfileDebug, stopProfileDebug};
use hyper::header::CONTENT_TYPE;
use hyper::{
    service::{make_service_fn, service_fn},
    Body, Method, Request, Response, Server,
};
use log::{error, info};
use prometheus::{Encoder, TextEncoder};
use serde::{Deserialize, Serialize};
use std::convert::Infallible;
use std::sync::Arc;

use crate::metric_exporter::PromExporter;

const CODE_NOERROR: i32 = 1;
const CODE_START_WITH_ERROR: i32 = 22;
const CODE_STOP_WITH_ERROR: i32 = 3;
const CODE_NO_OPERATION: i32 = 4;

const MSG_START_SUCCESS: &str = "start success";
const MSG_STOP_SUCCESS: &str = "stop success";
const MSG_START_DEBUG_SUCCESS: &str = "start debug success";
const MSG_STOP_DEBUG_SUCCESS: &str = "stop debug success";
const MSG_STATUS_RUNNING: &str = "running";
const MSG_STATUS_STOPPED: &str = "stopped";

const RESP_MISS_PAGE: &str = "Missing Page";
const RESP_UNEXPECT_REQUEST: &str = "parse request failed";

#[derive(Deserialize, Debug)]
struct ControlRequest {
    operation: String,
    pid: i32,
    tid: i32,
}

#[derive(Serialize)]
struct ControlResponse {
    code: i32,
    msg: String,
}

pub async fn start_serve(
    prom_exporter: Arc<PromExporter>,
    addr: ([u8; 4], u16),
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let make_svc = make_service_fn(move |_conn| {
        let prom_exporter = prom_exporter.clone();
        // This is the `Service` that will handle the connection.
        // `service_fn` is a helper to convert a function that
        // returns a Response into a `Service`.
        async move { Ok::<_, Infallible>(service_fn(move |req| serve_req(req, prom_exporter.clone()))) }
    });
    let addr = addr.into();

    let server = Server::bind(&addr).serve(make_svc);
    info!("HTTP Listening on http://{addr}");

    server
        .await
        .map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)
}

pub async fn serve_req(
    req: Request<Body>,
    exporter: Arc<PromExporter>,
) -> Result<Response<Body>, hyper::Error> {
    println!("Receiving request at path {}", req.uri());
    match (req.method(), req.uri().path()) {
        (&Method::GET, "/metrics") => handle_metrics(exporter).await,
        (&Method::POST, "/profile/start_debug") => handle_profile_start_debug(req).await,
        (&Method::POST, "/profile/stop_debug") => handle_profile_stop_debug().await,
        _ => response_miss(),
    }
}

async fn handle_metrics(exporter: Arc<PromExporter>) -> Result<Response<Body>, hyper::Error> {
    let mut buffer = vec![];
    let encoder = TextEncoder::new();
    let metric_families = exporter.exporter.registry().gather();
    encoder.encode(&metric_families, &mut buffer).unwrap();

    Ok(Response::builder()
        .status(200)
        .header(CONTENT_TYPE, encoder.format_type())
        .body(Body::from(buffer))
        .unwrap())
}

async fn get_controller_request(req: Request<Body>) -> Option<ControlRequest> {
    match hyper::body::to_bytes(req).await {
        Ok(body) => match serde_json::from_reader(body.as_ref()) {
            Ok(result) => Some(result),
            Err(err) => {
                error!("Parse request failed: {}", err);
                None
            }
        },
        Err(err) => {
            error!("Read request failed: {}", err);
            None
        }
    }
}

async fn handle_profile_start_debug(req: Request<Body>) -> Result<Response<Body>, hyper::Error> {
    match get_controller_request(req).await {
        Some(request) => {
            // 调用C代码
            info!(
                "handle_profile_start_debug Pid: {}, Tid: {}",
                request.pid, request.tid
            );
            unsafe { startProfileDebug(request.pid, request.tid) }
            controller_response(CODE_NOERROR, String::from(MSG_START_DEBUG_SUCCESS))
        }
        None => response_unexpect_request(),
    }
}

async fn handle_profile_stop_debug() -> Result<Response<Body>, hyper::Error> {
    info!("handle_profile_stop_debug");
    unsafe { stopProfileDebug() }
    controller_response(CODE_NOERROR, String::from(MSG_STOP_DEBUG_SUCCESS))
}

fn response_miss() -> Result<Response<Body>, hyper::Error> {
    Ok(Response::builder()
        .status(404)
        .body(Body::from(RESP_MISS_PAGE))
        .unwrap())
}

fn response_unexpect_request() -> Result<Response<Body>, hyper::Error> {
    Ok(Response::builder()
        .status(500)
        .body(Body::from(RESP_UNEXPECT_REQUEST))
        .unwrap())
}

fn controller_response(code: i32, msg: String) -> Result<Response<Body>, hyper::Error> {
    let response = ControlResponse { code, msg };
    Ok(Response::builder()
        .status(200)
        .body(Body::from(serde_json::to_string(&response).unwrap()))
        .unwrap())
}
