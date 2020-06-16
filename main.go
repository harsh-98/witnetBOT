package main

import (
	"flag"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetbot/helpers"
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
func loadConfig() {
	viper.SetConfigName(".witnetbot")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.ExpandEnv("$HOME"))
	viper.AddConfigPath(os.ExpandEnv("$PWD"))
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatal("Err in loading config file %s", err)
	}
}
func main() {
	initLoad()

	err := helpers.DB.Init()
	if err != nil {
		log.Fatal("Unable to connect to database")
	}
	defer helpers.DB.Close()

	helpers.TgBot, _ = tgbotapi.NewBotAPI(helpers.TgBotToken)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	go helpers.QueryWorker()

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
