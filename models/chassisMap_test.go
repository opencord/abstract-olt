/*
 Copyright 2017 the original author or authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package models_test

import (
	"testing"

	"gerrit.opencord.org/abstract-olt/models"
)

func TestChassisMap_GetPhyChassisMap(t *testing.T) {
	firstChassisMap := models.GetPhyChassisMap()
	secondChassisMap := models.GetPhyChassisMap()

	if firstChassisMap != secondChassisMap {
		t.Fatalf("GetPhyChassisMap should always return pointer to same map")
	}
}
func TestChassisMap_GetAbstractChassisMap(t *testing.T) {
	firstChassisMap := models.GetAbstractChassisMap()
	secondChassisMap := models.GetAbstractChassisMap()

	if firstChassisMap != secondChassisMap {
		t.Fatalf("GetPhyChassisMap should always return pointer to same map")
	}
}
