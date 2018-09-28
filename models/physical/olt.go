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

package physical

import "net"

/*
Represents an arbitrary OLT linecard
*/
type OLT interface {
	GetCLLI() string
	GetHostname() string
	GetAddress() net.TCPAddr
	GetNumber() int
	GetPorts() []PONPort
	GetParent() *Chassis
	GetDataSwitchPort() int
	SetNumber(int)
	activate() error
	Output()
}

/*
A basic representation of an OLT which fulfills the above interface,
and can be used in other OLT implementations
*/
type SimpleOLT struct {
	CLLI           string
	Hostname       string
	Address        net.TCPAddr
	Number         int
	Ports          []PONPort
	Active         bool
	Parent         *Chassis `json:"-"`
	DataSwitchPort int
}

func (s SimpleOLT) GetCLLI() string {
	return s.CLLI
}

func (s SimpleOLT) GetHostname() string {
	return s.Hostname
}

func (s SimpleOLT) GetAddress() net.TCPAddr {
	return s.Address
}

func (s SimpleOLT) GetNumber() int {
	return s.Number
}
func (s SimpleOLT) SetNumber(num int) {
	s.Number = num
}

func (s SimpleOLT) GetPorts() []PONPort {
	return s.Ports
}

func (s SimpleOLT) GetParent() *Chassis {
	return s.Parent
}

func (s SimpleOLT) GetDataSwitchPort() int {
	return s.DataSwitchPort
}
func (s SimpleOLT) activate() error {
	s.Active = true
	//TODO make call to XOS to activate phyiscal OLT
	return nil
}
func (s SimpleOLT) Output() error {
	//TODO make call to XOS to activate phyiscal OLT
	return nil
}
