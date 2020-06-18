package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/helpers"
	"github.com/harsh-98/witnetBOT/log"
	"github.com/spf13/viper"
)

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
		log.Logger.Fatalf("Err in loading config file %s", err)
	}
	return v
}
func main() {

	defaults := make(map[string]interface{})
	defaults["timer"] = 60
	defaults["ticker"] = 10
	v := readConfig(defaults)

	err := helpers.DB.Init(v)
	if err != nil {
		log.Logger.Fatal("Unable to connect to database")
	}
	defer helpers.DB.Close()
	// helpers.GenerateGraph(676523999)

	helpers.TgBot, _ = tgbotapi.NewBotAPI(v.GetString("tgToken"))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	go helpers.QueryWorker(v)

	updates, _ := helpers.TgBot.GetUpdatesChan(u)

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
