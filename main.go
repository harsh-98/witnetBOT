package main

import (
	"flag"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initLoad() {
	debug := flag.Bool("debug", false, "for debugging")
	flag.Parse()
	log.SetOutput(os.Stdout)
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func readConfig(defaults map[string]interface{}) *viper.Viper {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME")
	v.SetConfigName(".witnet")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("Err in loading config file %s", err)
	}
	return v
}
func main() {
	initLoad()
	var defaults map[string]interface{}
	v := readConfig(defaults)

	err := helpers.DB.Init(v)
	if err != nil {
		log.Fatal("Unable to connect to database")
	}
	defer helpers.DB.Close()

	helpers.TgBot, _ = tgbotapi.NewBotAPI(helpers.TgBotToken)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	go helpers.QueryWorker(v)

	updates, _ := helpers.TgBot.GetUpdatesChan(u)
	// go helpers.GetHeartBeat()

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				helpers.CommandReceived(update)
			}
			if update.Message.ReplyToMessage != nil {
				helpers.ReplyReceived(update.Message)
			}
		}
		if update.CallbackQuery != nil {
			helpers.CallbackQueryReceived(update.CallbackQuery)
		}
	}
}
