package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Item struct {
	Id_str string
	Text   string
	User   struct {
		Name        string
		Screen_name string
	}
}

type Response struct {
	Statuses []Item
}

func GetTag(hashtag string) {
	url := "https://api.twitter.com/1.1/search/tweets.json?q=%23" + hashtag

	oauth := ``

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", oauth)

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}

	r := new(Response)
	err = json.NewDecoder(resp.Body).Decode(r)

	for _, child := range r.Statuses {
		fmt.Printf("@%s -> %s\n", child.User.Screen_name, child.Text)
	}

}

func main() {
	GetTag("io14extended")

}
