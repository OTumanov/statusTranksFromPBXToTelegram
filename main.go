package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	noToken    = "Please set TOKEN aenvironment variables\n"
	noChatID   = "Please set CHAT_ID aenvironment variables\n"
	noThredID  = "*OPTIONAL* Set THRED_ID aenvironment variables\n"
	noJsonURL  = "Please set JSON_URL aenvironment variables\n"
	noUsername = "Please set USERNAME aenvironment variables\n"
	noPassword = "Please set PASSWORD aenvironment variables\n"
	Cross      = "❌"
	Check      = "✅"
	Phone      = "📞"
)

var (
	Token    = os.Getenv("BOT_TOKEN")
	ChatID   = os.Getenv("CHAT_ID")
	ThredID  = os.Getenv("CHAT_THREAD_ID")
	JsonURL  = os.Getenv("URL")
	Username = os.Getenv("USER")
	Password = os.Getenv("PASS")
)

var bot *tgbotapi.BotAPI

type Trunk struct {
	Name   string
	Status string
}

func initTelegramBot() error {
	var err error
	bot, err = tgbotapi.NewBotAPI(Token)
	if err != nil {
		return err
	}
	return nil
}

func sendTelegramMessage(message string) {
	msg := tgbotapi.NewMessageToChannel(ChatID, message)
	if ThredID != "" {
		msg.ReplyToMessageID, _ = strconv.Atoi(ThredID)
	}
	bot.Send(msg)
}

func getTrunkStatus() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", JsonURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.SetBasicAuth(Username, Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var data []map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	rtTrunks := make([]Trunk, 0)
	onlineTrunks := make([]Trunk, 0)
	offlineTrunks := make([]Trunk, 0)

	for _, trunk := range data {
		name, ok := trunk["resource"].(string)
		if !ok {
			continue
		}
		state, ok := trunk["state"].(string)
		if !ok {
			continue
		}

		trunk := Trunk{Name: name, Status: strings.ToLower(state)}
		if strings.HasPrefix(name, "rt") {
			rtTrunks = append(rtTrunks, trunk)
		} else if trunk.Status == "online" {
			onlineTrunks = append(onlineTrunks, trunk)
		} else {
			offlineTrunks = append(offlineTrunks, trunk)
		}
	}

	sort.Slice(rtTrunks, func(i, j int) bool {
		return rtTrunks[i].Name < rtTrunks[j].Name
	})

	sort.Slice(onlineTrunks, func(i, j int) bool {
		return onlineTrunks[i].Name < onlineTrunks[j].Name
	})

	sort.Slice(offlineTrunks, func(i, j int) bool {
		return offlineTrunks[i].Name < offlineTrunks[j].Name
	})

	rtMessage := fmt.Sprintf("%vRT транки: %d\n", Phone, len(rtTrunks))

	for _, trunk := range rtTrunks {
		statusSymbol := Cross
		if trunk.Status == "online" && trunk.Name != "rt" {
			statusSymbol = Check
		}
		rtMessage += fmt.Sprintf("- %s: %s %s\n", trunk.Name, statusSymbol, trunk.Status)
	}

	offlineTrunksCount := len(offlineTrunks)
	rtMessage += fmt.Sprintf("\n%vОффлайн транки: %d\n", Cross, offlineTrunksCount)

	for _, trunk := range offlineTrunks {
		rtMessage += fmt.Sprintf("- %s %v\n", trunk.Name, Cross)
	}

	onlineTrunksCount := 0
	for _, trunk := range onlineTrunks {
		if !strings.HasPrefix(trunk.Name, "rt") {
			onlineTrunksCount++
		}
	}

	rtMessage += fmt.Sprintf("\n%vДругие транки: %d\n", Phone, onlineTrunksCount)

	sendTelegramMessage(rtMessage)
}

func main() {
	msg := ""

	if Token == "" {
		msg += fmt.Sprintf(noToken)
	}
	if ChatID == "" {
		msg += fmt.Sprintf(noChatID)
	}
	if ThredID == "" {
		msg += fmt.Sprintf(noThredID)
	}
	if JsonURL == "" {
		msg += fmt.Sprintf(noJsonURL)
	}
	if Username == "" {
		msg += fmt.Sprintf(noUsername)
	}
	if Password == "" {
		msg += fmt.Sprintf(noPassword)
	}

	if msg != "" {
		fmt.Println(msg)
		return
	}

	err := initTelegramBot()
	if err != nil {
		fmt.Println("Error initializing Telegram bot:", err)
		return
	}
	getTrunkStatus()
}
