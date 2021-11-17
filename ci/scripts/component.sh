#!/bin/bash -eux

pushd dp-files-api
  make test-component
popd
