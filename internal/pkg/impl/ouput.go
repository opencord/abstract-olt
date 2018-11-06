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
package impl

import (
	"fmt"
	"log"
	"os"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/options"
	context "golang.org/x/net/context"
)

/*
DoOutput - creates a backup and stores it to disk/mongodb
*/
func DoOutput() (bool, error) {
	if isDirty {
		myChan := getSyncChannel()
		<-myChan
		defer done(myChan, true)
		chassisMap := models.GetChassisMap()
		if settings.GetMongo() {
			client, err := mongo.NewClient(settings.GetMongodb())
			client.Connect(context.Background())
			if err != nil {
				log.Printf("client connect to mongo db @%s failed with %v\n", settings.GetMongodb(), err)
			}
			defer client.Disconnect(context.Background())
			for clli, chassisHolder := range *chassisMap {
				json, _ := (chassisHolder).Serialize()
				collection := client.Database("AbstractOLT").Collection("backup")
				doc := bson.NewDocument(bson.EC.String("_id", clli))
				filter := bson.NewDocument(bson.EC.String("_id", clli))
				doc.Append(bson.EC.Binary("body", json))

				updateDoc := bson.NewDocument(bson.EC.SubDocument("$set", doc))
				//update or insert if not existent
				upsert := true
				res, err := collection.UpdateOne(context.Background(), filter, updateDoc, &options.UpdateOptions{Upsert: &upsert})
				if err != nil {
					log.Printf("collection.UpdateOne failed with %v\n", err)
				} else {
					id := res.UpsertedID
					if settings.GetDebug() {
						log.Printf("Update Succeeded with id %v\n", id)
					}
				}
			}
		} else {
			for clli, chassisHolder := range *chassisMap {

				json, _ := (chassisHolder).Serialize()
				if settings.GetMongo() {

				} else {
					//TODO parameterize dump location
					backupFile := fmt.Sprintf("backup/%s", clli)
					f, _ := os.Create(backupFile)

					defer f.Close()

					_, _ = f.WriteString(string(json))
					f.Sync()
				}
			}
		}
		isDirty = false
	} else {
		if settings.GetDebug() {
			log.Print("Not dirty not dumping config")
		}

	}
	return true, nil

}
