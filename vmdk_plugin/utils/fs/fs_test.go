// Copyright 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import (
	"github.com/stretchr/testify/assert"
	"github.com/vmware/docker-volume-vsphere/vmdk_plugin/utils/fs"
	"testing"
)

const funnyfs = "funnyfs"

func TestVerifyFSSupport(t *testing.T) {
	err := fs.VerifyFSSupport(fs.FstypeDefault)
	assert.Nil(t, err, "Fstype %s should be supported", fs.FstypeDefault)
}

func TestVerifyFSSupportError(t *testing.T) {
	err := fs.VerifyFSSupport(funnyfs)
	assert.NotNil(t, err, "Fstype %s shouldn't be supported", funnyfs)
}
