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

package settings

var debug = false
var dummy = false
var mongo = false
var grpc = true
var mongodb = ""
var mongoUser = ""
var mongoPasswd = ""

/*
SetDebug - sets debug setting
*/
func SetDebug(logDebug bool) {
	debug = logDebug
}

/*
GetDebug returns debug setting
*/
func GetDebug() bool {
	return debug
}

/*
SetDummy sets dummy mode
*/
func SetDummy(dummyMode bool) {
	dummy = dummyMode
}

/*
GetDummy - returns the current value of dummy
*/
func GetDummy() bool {
	return dummy
}

/*
SetMongo - sets useMongo mode
*/
func SetMongo(useMongo bool) {
	mongo = useMongo
}

/*GetMongo - returns the value of mongo
 */
func GetMongo() bool {
	return mongo
}

/*
SetMongodb - sets the connection string for mongo db used for backups
*/
func SetMongodb(connectString string) {
	mongodb = connectString
}

/*
GetMongodb - returns the connection string used for connecting to mongo db
*/
func GetMongodb() string {
	return mongodb
}

/*
SetMongoUser - sets the connection string for mongo db used for backups
*/
func SetMongoUser(user string) {
	mongoUser = user
}

/*
GetMongoUser - returns the connection string used for connecting to mongo db
*/
func GetMongoUser() string {
	return mongoUser
}

/*
SetMongoPassword - sets the connection string for mongo db used for backups
*/
func SetMongoPassword(password string) {
	mongoPasswd = password
}

/*
GetMongoPassword - returns the connection string used for connecting to mongo db
*/
func GetMongoPassword() string {
	return mongoPasswd
}

/*
SetGrpc - sets useGrpc mode
*/
func SetGrpc(useGrpc bool) {
	grpc = useGrpc
}

/*
GetGrpc - returns the value of grpc
*/
func GetGrpc() bool {
	return grpc
}
