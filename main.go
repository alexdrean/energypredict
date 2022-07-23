package main

import (
	"energypredict/cfg"
	"energypredict/estimator"
	"energypredict/solarapi"
	"energypredict/vrmapi"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"time"
)

func main() {
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Panicln(err)
		return
	}
	ticker := time.NewTicker(time.Duration(config.Frequency) * time.Minute)
	defer ticker.Stop()
	for {
		solarChan := make(chan *solarapi.SolarPrediction)
		vrmChan := make(chan *vrmapi.VrmResponse)
		go solarapi.Get(config.Solar, solarChan)
		go vrmapi.Get(config.Vrm, vrmChan)
		if solar, vrm := <-solarChan, <-vrmChan; solar != nil && vrm != nil {
			estimate, err := estimator.GetEstimate(config, solar, vrm)
			if err != nil {
				log.Println(err)
				continue
			}
			for i, line := range estimate.Battery {
				println(estimate.TimeAt(i).Format("2006-01-02 15:04") + "," + strconv.Itoa(int(line)))
			}
			if estimate.GetTimeRemaining() >= 0 {
				sendWarning(config, estimate, vrm)
			}
		}
		<-ticker.C
	}
}

func sendWarning(config *cfg.Config, estimate *estimator.Estimate, vrm *vrmapi.VrmResponse) {
	outages := estimate.GetOutages()
	s := fmt.Sprintf("SOC = %.1f%% (%.2fkWh), Power = %.0fW\n", vrm.SOC, estimate.Battery[0]/1000.0, vrm.PowerAvg)
	format := "Mon 15:04"
	for _, outage := range outages {
		s += outage[0].Format(format) + " - "
		if outage[1].Unix() != 0 {
			s += outage[1].Format(format)
		} else {
			s += "<...>"
		}
		s += "\n"
	}
	bot, err := tgbotapi.NewBotAPI(config.Telegram.ApiKey)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = bot.Send(tgbotapi.NewMessage(config.Telegram.ChatId, s))
	if err != nil {
		log.Println(err)
		return
	}
}
