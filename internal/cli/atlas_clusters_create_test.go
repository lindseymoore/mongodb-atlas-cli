// Copyright 2020 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	"github.com/mongodb/mongocli/internal/mocks"
	"github.com/spf13/afero"
)

func TestAtlasClustersCreate_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStore := mocks.NewMockClusterCreator(ctrl)
	defer ctrl.Finish()

	expected := &mongodbatlas.Cluster{}

	t.Run("flags run", func(t *testing.T) {
		createOpts := &atlasClustersCreateOpts{
			name:         "ProjectBar",
			region:       "US",
			instanceSize: atlasM2,
			members:      3,
			diskSizeGB:   10,
			backup:       false,
			mdbVersion:   currentMDBVersion,
			store:        mockStore,
		}

		cluster, _ := createOpts.newCluster()
		mockStore.
			EXPECT().
			CreateCluster(cluster).Return(expected, nil).
			Times(1)

		err := createOpts.Run()
		if err != nil {
			t.Fatalf("Run() unexpected error: %v", err)
		}
	})

	t.Run("file run", func(t *testing.T) {
		appFS := afero.NewMemMapFs()
		// create test file
		fileYML := `
{
  "name": "ProjectBar",
  "diskSizeGB": 10,
  "numShards": 1,
  "providerSettings": {
    "providerName": "AWS",
    "instanceSizeName": "M2",
    "regionName": "US"
  },
  "clusterType" : "REPLICASET",
  "replicationFactor": 3,
  "replicationSpecs": [{
    "numShards": 1,
    "regionsConfig": {
      "US_EAST_1": {
        "analyticsNodes": 0,
        "electableNodes": 3,
        "priority": 7,
        "readOnlyNodes": 0
      }
    },
    "zoneName": "Zone 1"
  }],
  "backupEnabled": false,
  "providerBackupEnabled" : false
}`
		fileName := "test.json"
		_ = afero.WriteFile(appFS, fileName, []byte(fileYML), 0600)

		createOpts := &atlasClustersCreateOpts{
			filename: fileName,
			fs:       appFS,
			store:    mockStore,
		}

		cluster, _ := createOpts.newCluster()
		mockStore.
			EXPECT().
			CreateCluster(cluster).Return(expected, nil).
			Times(1)

		err := createOpts.Run()
		if err != nil {
			t.Fatalf("Run() unexpected error: %v", err)
		}
	})
}
