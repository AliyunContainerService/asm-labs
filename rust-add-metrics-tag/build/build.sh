cargo build --target wasm32-wasip1 --release
cp target/wasm32-wasip1/release/rust_add_metrics_tag.wasm build/plugin.wasm