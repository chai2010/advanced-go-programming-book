#!/bin/sh -eu

bazel \
  --batch \
  --output_base=$HOME/.cache/_grpc_gateway_bazel \
  --host_jvm_args=-Xmx500m \
  --host_jvm_args=-Xms500m \
  build \
  --local_resources=400,1,1.0 \
  //...
