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

package models

import (
	"sync"

	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

var once sync.Once
var absOnce sync.Once
var chassisMap map[string]*physical.Chassis
var aChassisMap map[string]*abstract.Chassis

/*
GetPhyChassisMap return the chassis map singleton
*/
func GetPhyChassisMap() *map[string]*physical.Chassis {
	// the go singleton pattern
	once.Do(func() {
		chassisMap = make(map[string]*physical.Chassis)
	})
	return &chassisMap
}

/*
GetAbstractChassisMap return the chassis map singleton
*/
func GetAbstractChassisMap() *map[string]*abstract.Chassis {
	// the go singleton pattern
	absOnce.Do(func() {
		aChassisMap = make(map[string]*abstract.Chassis)
	})
	return &aChassisMap
}
