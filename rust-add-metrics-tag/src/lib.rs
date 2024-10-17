use std::io::Read;

use log::{info, error};
use proxy_wasm::traits::*;
use proxy_wasm::types::*;
use base64::{engine::general_purpose::STANDARD, read::DecoderReader};

proxy_wasm::main! {{
    proxy_wasm::set_log_level(LogLevel::Trace);
    proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> { Box::new(HttpHeadersRoot) });
}}

struct HttpHeadersRoot;

// 一些基础工具函数，加入这一句后，可以直接在self中调用各种工具函数
//  比如：self.set_property(path, value)
impl Context for HttpHeadersRoot {}

impl RootContext for HttpHeadersRoot {
    fn get_type(&self) -> Option<ContextType> {
        Some(ContextType::HttpContext)
    }

    fn create_http_context(&self, context_id: u32) -> Option<Box<dyn HttpContext>> {
        Some(Box::new(HttpHeaders { context_id }))
    }
}

struct HttpHeaders {
    context_id: u32,
}

impl Context for HttpHeaders {}

impl HttpContext for HttpHeaders {
    fn on_http_request_headers(&mut self, _: usize, _: bool) -> Action {
        info!("#{} wasm-rust: on_http_request_headers", self.context_id);

        match self.get_http_request_header("user-name") {
            Some(encoded_user_name) => {
                let mut decoder = DecoderReader::new(encoded_user_name.as_bytes(),  &STANDARD);
                let mut user_name = String::new();
                match decoder.read_to_string(&mut user_name){
                    Ok(_) => {
                        let trimed: &str = user_name.trim();
                        info!("#{} wasm-rust: user-name: {}", self.context_id, trimed);
                        self.set_property(vec!["user-name"], Some(trimed.as_bytes()));
                    },
                    Err(e) => {
                        error!("#{} wasm-rust: error decoding user-name: {}", self.context_id, e);
                    }
                }
                Action::Continue
            }
            _ => {
                info!("#{} wasm-rust: cannot get user-name in request header", self.context_id);
                Action::Continue
            }
        }
    }
}