package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

const urlUsers string = "http://governor.verf.io/api/users/"

var myClient = &http.Client{Timeout: 10 * time.Second}

type task struct {
	ID       string `json:"id"`
	Number   int64  `json:"number"`
	Source   string `json:"source"`
	SourceID string `json:"sourceid"`
	User     string `json:"user"`
	Action   string `json:"action"`
	State    string `json:"state"`
	Email    string `json:"email"`
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

		tickets := getTickets("done")

		for _, ticket := range *tickets {
			changeJiraStatus(&ticket, client)
		}

		time.Sleep(5 * time.Second)
	}
}

func getTickets(status string) *[]task {

	resp, err := myClient.Get(urlUsers)
	if err != nil {
		println("Error:", err)
	}
	defer resp.Body.Close()

	var ticketsTemp []task
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respByte := buf.Bytes()
	err = json.Unmarshal(respByte, &ticketsTemp)
	if err != nil {
		println("Error:", err)
	}

	var tickets []task
	for _, t := range ticketsTemp {
		if t.State == status && t.Source == "jira" {
			tickets = append(tickets, t)
		}
	}

	fmt.Println("Results All: ", tickets)

	return &tickets
}

func changeStatus(ticket *task, state string) {
	println("changeStatus :", state)
	ticket.State = state

	var urlUser = urlUsers + ticket.ID
	j, err := json.Marshal(ticket)
	if err != nil {
		fmt.Println("Error marshaling ticket into JSON")
	}

	t := bytes.NewReader(j)
	resp, err := myClient.Post(urlUser, "application/json", t)
	if err != nil {
		fmt.Println("Error with POST request")
	}
	defer resp.Body.Close()
}

func changeJiraStatus(ticket *task, client *jira.Client) {

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
