#!/bin/bash -eux

# Run component tests in docker compose defined in features/steps/compose folder
pushd dis-files-api/features/compose
  COMPONENT_TEST_USE_LOG_FILE=true docker-compose up --attach dp-files-api --abort-on-container-exit
  e=$?
popd

# Cat the component-test output file and remove it so log output can
# be seen in Concourse
pushd dp-files-api
  cat component-output.txt && rm component-output.txt
popd

# exit with the same code returned by docker compose
exit $e