package main

import (
	"time"

	utils "github.com/verfio/governor-clerk-utils"
)

func main() {
	for {
		tickets := utils.GetJiraTickets()
		for _, ticket := range *tickets {
			utils.ChangeJiraStatus(&ticket)
		}
		time.Sleep(5 * time.Second)
	}
}
