#!/bin/bash

# Copyright 2021 MongoDB Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# shellcheck disable=SC1091
if [[ -f project.sh ]]; then
  source project.sh
fi

if [[ "${MUST_CLEANUP_CM-}" != "true" ]]; then
  echo "skipping cloud manager clean up post task"
  exit 0
fi

if [[ -z "${MCLI_PROJECT_ID-}" ]]; then
  echo "MCLI_PROJECT_ID not set"
  exit 1
fi

delete_project() {
  mongocli iam projects delete "${MCLI_PROJECT_ID-}" --force
} 

for _ in {1..10}
do
  echo "attempting to delete cloud manager project ${MCLI_PROJECT_ID-}"
  if [[ $(delete_project) ]]; then
    echo "cloud manager project ${MCLI_PROJECT_ID-} was deleted successfully"
    exit 0
  fi
  sleep 30
done

echo "cloud manager project ${MCLI_PROJECT_ID-} was unable to be deleted"
exit 1
