#!/bin/bash -eux

pushd dp-files-api
  make build
  cp build/dp-files-api Dockerfile.concourse ../build
popd
