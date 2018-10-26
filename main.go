package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"

	jira "github.com/andygrunwald/go-jira"
)

//MongoDb const is a Mongo DB address
const MongoDb string = "35.232.89.65"

type task struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Number   int64         `json:"number" bson:"number"`
	Source   string        `json:"source" bson:"source"`
	SourceID string        `json:"sourceid" bson:"sourceid"`
	User     string        `json:"user" bson:"user"`
	Action   string        `json:"action" bson:"action"`
	State    string        `json:"state" bson:"state"`
	Email    string        `json:"email" bson:"email"`
}

func main() {
	jiraURL := "http://jira.verf.io:8080"
	jiraUsername := "brian"
	jiraPassword := "P@ssw0rd"

	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(jiraUsername),
		Password: strings.TrimSpace(jiraPassword),
	}

	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(jiraURL))
	if err != nil {
		fmt.Printf("\nerror with Jira connection: %v\n", err)
		return
	}

	for {

		tickets := getdataMongo("done")

		for _, ticket := range tickets {
			changeJiraStatus(ticket, client)
		}

		time.Sleep(5 * time.Second)
	}
}

func getdataMongo(status string) []task {

	session, err := mgo.Dial(MongoDb)

	if err != nil {
		println("Error: Could not connect  DB ")
	}
	var tasks []task

	c := session.DB("governor").C("tasks")

	err = c.Find(bson.M{"state": status, "source": "jira"}).Sort("-timestamp").All(&tasks)

	if err != nil {
		println("Error: Could not find data in DB ")
		log.WithFields(log.Fields{
			"Try to find data": "Error",
		}).Info(err)
	}
	fmt.Println("Results All: ", tasks)

	defer session.Close()

	return tasks

}

func changeStatus(ticket task, state string) {

	println("changeStatus :", state)

	session, err := mgo.Dial(MongoDb)

	if err != nil {
		println("Error: Could not connect on MongoDB ")
	}

	c := session.DB("governor").C("tasks")

	// Update
	colQuerier := bson.M{"_id": ticket.ID}
	change := bson.M{"$set": bson.M{"state": state}}
	err = c.Update(colQuerier, change)
	if err != nil {
		println("Error: Could not update DB ")
	}
}

func changeJiraStatus(ticket task, client *jira.Client) {

	clientIssue := client.Issue

	issue, _, err := clientIssue.Get(ticket.SourceID, nil)
	fmt.Println("Update status in Jira to \"DONE\" for Issue:", issue.ID)
	//    11 - from TO DO to IN PROGRESS
	//    21 - from TO DO to DONE

	//    31 - from IN PROGRESS to TO DO
	//    41 - from IN PROGRESS to DONE

	//    51 - from DONE to TO DO
	//    61 - from DONE to IN PROGRESS
	//update status to "DONE"
	_, err = clientIssue.DoTransition(issue.ID, "21")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	changeStatus(ticket, "Closed")
	issue, _, err = clientIssue.Get(ticket.SourceID, nil)
	fmt.Println("Status of Issue", issue.ID, "was successfully updated to:", issue.Fields.Status.Name)
}
