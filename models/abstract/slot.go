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

package abstract

/*
Slot models a collection of PON ports (likely a single chassis) as if it is a
line card within a chassis
*/
type Slot struct {
	// DeviceID string
	// Hostname string
	// Address  net.TCPAddr
	Number int
	Ports  [16]Port
	Parent *Chassis `json:"-"`
}
