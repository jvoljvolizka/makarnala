package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mmcdole/gofeed"
)

type update struct {
	UpdateID int     `json:"update_id"`
	Message  message `json:"message"`
}

type message struct {
	Text string `json:"text"`
	Chat chat   `json:"chat"`
}

type chat struct {
	ID int `json:"id"`
}

func parseRss(url string) string {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)
	body := feed.Items[0].Content
	z := strings.Replace(body, "</p>", "\n</p>", -1)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(z))
	if err != nil {
		log.Fatal(err)
	}
	out := ""
	doc.Find(".md").Each(func(i int, s *goquery.Selection) {
		// Get the pasta
		out = s.Find("p").Text()
		out = feed.Items[0].Link + "\n" + out
	})
	return out
}

func getLink(url string) string {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)
	return feed.Items[0].Link
}

func sendReq(botToken string, text string, ID int) *http.Response {
	clientID := strconv.Itoa(ID)
	req, err := http.NewRequest("GET", "https://api.telegram.org/bot"+botToken+"/sendMessage", nil)
	client := &http.Client{}
	if err != nil {
		fmt.Println(err.Error())
	}
	q := req.URL.Query()
	q.Add("chat_id", clientID)
	q.Add("text", text)

	req.URL.RawQuery = q.Encode()
	fmt.Println(q.Encode())
	res, _ := client.Do(req)
	return res
}

func clean(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	botToken := os.Getenv("TOKEN")
	b := []byte(req.Body)
	var f update
	json.Unmarshal(b, &f)
	chatID := f.Message.Chat.ID

	if f.Message.Text == "/makarna" || f.Message.Text == "/makarna@kopyamakarnabot" {
		text := parseRss("https://www.reddit.com/r/kopyamakarna/rising.rss")
		res := sendReq(botToken, text, chatID)
		if res.StatusCode != 200 {
			sendReq(botToken, getLink("https://www.reddit.com/r/kopyamakarna/rising.rss"), chatID)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       `{"result" : true}`,
		}, nil
	}

	if f.Message.Text == "/makarna2" || f.Message.Text == "/makarna2@kopyamakarnabot" {
		text := parseRss("https://www.reddit.com/r/kopyamakarna2/rising.rss")
		res := sendReq(botToken, text, chatID)
		if res.StatusCode != 200 {
			sendReq(botToken, getLink("https://www.reddit.com/r/kopyamakarna2/rising.rss"), chatID)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       `{"result" : true}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"result" : true}`,
	}, nil
}

func main() {
	lambda.Start(clean)
}
