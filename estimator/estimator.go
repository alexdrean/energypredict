package estimator

import (
	"energypredict/cfg"
	"energypredict/solarapi"
	"energypredict/vrmapi"
	"math"
	"time"
)

func (e Estimate) GetTimeRemaining() time.Duration {
	for i, line := range e.Battery {
		if line <= 0.0 {
			return time.Duration(i) * e.Resolution
		}
	}
	return -1
}

type Estimate struct {
	Time       time.Time
	Resolution time.Duration
	Battery    []float64
}

func (e Estimate) GetOutages() [][2]time.Time {
	outages := make([][2]time.Time, 0)
	outageStart := -1
	for i, batteryLevel := range e.Battery {
		if batteryLevel <= 0 {
			if outageStart == -1 {
				outageStart = i
			}
		} else {
			if outageStart != -1 {
				outages = append(outages, [2]time.Time{
					e.Time.Add(time.Duration(outageStart) * e.Resolution),
					e.Time.Add(time.Duration(i) * e.Resolution),
				})
				outageStart = -1
			}
		}
	}
	if outageStart != -1 {
		outages = append(outages, [2]time.Time{
			e.Time.Add(time.Duration(outageStart) * e.Resolution),
			time.Unix(0, 0),
		})
	}
	return outages
}

func (e Estimate) TimeAt(i int) time.Time {
	return e.Time.Add(time.Duration(i) * e.Resolution)
}

func GetEstimate(config *cfg.Config, solar *solarapi.SolarPrediction, vrm *vrmapi.VrmResponse) (*Estimate, error) {
	resolution := time.Minute
	e := Estimate{
		Time:       time.Now(),
		Resolution: resolution,
		Battery:    make([]float64, config.Lookahead*int(time.Hour/resolution)),
	}
	batteryWh := vrm.SOC * float64(config.Vrm.BatteryWh) * 0.01
	for i := 0; i < len(e.Battery); i++ {
		t := e.TimeAt(i)
		solarOutput := adjust(e.Time, solar.Get(t), 52)
		batteryWh += solarOutput / 60.0
		batteryWh -= vrm.PowerAvg / 60.0
		batteryWh = clamp(0, batteryWh, float64(config.Vrm.BatteryWh))
		e.Battery[i] = batteryWh
	}
	return &e, nil
}

func clamp(min float64, value float64, max float64) float64 {
	if min >= value {
		return min
	}
	if value >= max {
		return max
	}
	return value
}

func adjust(time time.Time, power float64, voltage float64) float64 {
	if (time.Hour() == 16 && time.Minute() > 30) || time.Hour() >= 17 {
		power = power * 0.6
	}
	return math.Min(power, voltage*90.0)
}
