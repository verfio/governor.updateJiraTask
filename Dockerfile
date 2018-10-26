FROM golang:1.11-stretch

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 

RUN go get -u   "github.com/andygrunwald/go-jira"

RUN GOARCH=amd64 GOOS=linux go build -o updateJiraTask main.go

EXPOSE 3000

CMD ["./updateJiraTask"]