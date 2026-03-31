#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-files-api
  make lint
popd