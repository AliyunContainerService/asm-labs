docker run --rm -it \
      -v $(pwd)/test/envoy-demo.yaml:/envoy-demo.yaml \
      -v $(pwd)/build/plugin.wasm:/main.wasm \
      -p 18000:18000 \
      -p 9902:9902 \
      envoyproxy/envoy:v1.28.0 \
          -c envoy-demo.yaml