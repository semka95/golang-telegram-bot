package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var rss = map[string]string{
	"Habr": "https://habrahabr.ru/rss/best/",
}

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	URL   string `xml:"guid"`
	Title string `xml:"title"`
}

func getNews(url string) (*RSS, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	rss := new(RSS)
	err = xml.Unmarshal(body, rss)
	if err != nil {
		return nil, err
	}

	return rss, nil
}

func main() {
	var (
		BotToken = os.Getenv("TOKEN")
		AppURL   = os.Getenv("APP_URL")
		Port     = os.Getenv("PORT")
	)

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	botURL := AppURL + ":443/" + BotToken
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(botURL))
	if err != nil {
		panic(err)
	}

	info, err := bot.GetWebhookInfo()
	log.Println(info)
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	go http.ListenAndServe(":"+Port, nil)
	log.Printf("start listen :%s", Port)

	// получаем все обновления из канала updates
	for update := range updates {
		if url, ok := rss[update.Message.Text]; ok {
			rss, err := getNews(url)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"sorry, error happend",
				))
			}
			for _, item := range rss.Items {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					item.URL+"\n"+item.Title,
				))
			}
		} else {
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`there is only Habr feed availible`,
			))
		}

	}
}
