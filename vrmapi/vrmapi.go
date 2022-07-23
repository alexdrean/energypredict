package vrmapi

import (
	"encoding/json"
	"energypredict/cfg"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Attr [][2]float64

type rawVrmResponse struct {
	Records struct {
		Data struct {
			Soc          Attr `json:"144"`
			BatteryPower Attr `json:"243"`
			SolarPower   Attr `json:"107"`
		} `json:"data"`
	} `json:"records"`
}

type VrmResponse struct {
	SOC      float64
	PowerAvg float64
}

func Get(config cfg.Vrm, channel chan *VrmResponse) {
	// TODO
	onehourago := time.Now().Add(-1 * time.Hour)
	req, err := http.NewRequest(http.MethodGet,
		"https://vrmapi.victronenergy.com/v2/installations/"+config.InstallationId+"/widgets/Graph?attributeCodes[]=bp&attributeCodes[]=bs&attributeCodes[]=ScW&start="+strconv.FormatInt(onehourago.Unix(), 10),
		nil)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	req.Header.Set("X-Authorization", "Bearer "+config.ApiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	defer res.Body.Close()
	data := new(rawVrmResponse)
	jsonResponse, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	err = json.Unmarshal(jsonResponse, data)
	if err != nil {
		log.Println(err)
		channel <- nil
		return
	}
	result := VrmResponse{
		PowerAvg: data.Records.Data.SolarPower.avg() - data.Records.Data.BatteryPower.avg(),
		SOC:      data.Records.Data.Soc.last(),
	}
	channel <- &result
}

func (a Attr) last() float64 {
	return a[len(a)-1][1]
}

func (a Attr) avg() float64 {
	total := 0.0
	for _, item := range a {
		total += item[1]
	}
	return total / float64(len(a))
}
