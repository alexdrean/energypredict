package solarapi

import (
	"encoding/json"
	"energypredict/cfg"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

type solarResponse struct {
	Result struct {
		Watts        map[string]int `json:"watts"`
		WattHours    map[string]int `json:"watt_hours"`
		WattHoursDay map[string]int `json:"watt_hours_day"`
	} `json:"result"`
}

type Line struct {
	Datetime time.Time
	Power    int
}
type SolarPrediction []Line

func Get(cfg cfg.Solar, channel chan *SolarPrediction) {
	// TODO
	url := fmt.Sprintf("https://api.forecast.solar/%s/estimate/%.6f/%.6f/%d/%d/%.2f",
		cfg.ApiKey,
		cfg.Lat,
		cfg.Lon,
		cfg.Heading-180,
		cfg.Slope,
		cfg.Power,
	)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	defer res.Body.Close()
	jsonResponse, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	data := new(solarResponse)
	err = json.Unmarshal(jsonResponse, data)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	keys := make([]time.Time, len(data.Result.Watts))
	keyValues := make(map[time.Time]int, len(keys))
	n := 0
	for k, v := range data.Result.Watts {
		keys[n], err = time.ParseInLocation("2006-01-02 15:04:05", k, time.Local)
		keyValues[keys[n]] = v
		n++
		if err != nil {
			log.Println(err)
			channel <- nil
			return
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	prediction := make(SolarPrediction, len(keys))
	n = 0
	for _, k := range keys {
		prediction[n] = Line{k, keyValues[k]}
		n++
	}
	total := int64(0)
	for n, p := range prediction {
		if n == 0 {
			continue
		}
		if p.Datetime.Day() == 23 {
			total += int64(prediction[n-1].Power+p.Power) * int64((p.Datetime.Sub(prediction[n-1].Datetime))/time.Second) / 2
		}
	}
	channel <- &prediction
}

func (p SolarPrediction) Get(t time.Time) float64 {
	if t.Before(p[0].Datetime) {
		return 0
	}
	if t.After(p.last().Datetime) {
		return 0
	}
	n := -1
	for i, l := range p {
		if t.Before(l.Datetime) {
			n = i
			break
		}
	}
	start := p[n-1].Datetime
	end := p[n].Datetime
	duration := end.Sub(start)
	timeFromStart := t.Sub(start)
	ratio := float64(timeFromStart) / float64(duration)
	estimatedPower := ratio*float64(p[n].Power) + (1-ratio)*float64(p[n-1].Power)
	return estimatedPower
}

func (p SolarPrediction) last() Line {
	return p[len(p)-1]
}
