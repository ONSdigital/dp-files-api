#!/bin/bash -eux

pushd dp-files-api
  make docker-test-component
popd
