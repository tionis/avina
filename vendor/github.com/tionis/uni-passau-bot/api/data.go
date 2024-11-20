package uni_passau_bot

import (
	"time"
)

var tmpData map[string]map[string]tmpDataObject

type tmpDataObject struct {
	data       string
	validUntil time.Time
}

func dbInit() {
	// Init RAM Store
	tmpData = make(map[string]map[string]tmpDataObject)
}

func setTmp(bucket string, key string, value string, duration time.Duration) {
	var dataToSave tmpDataObject
	dataToSave.data = value
	dataToSave.validUntil = time.Now().Add(duration)
	if tmpData[bucket] == nil {
		tmpData[bucket] = make(map[string]tmpDataObject)
	}
	tmpData[bucket][key] = dataToSave
	// TODO init job to delete old values
}

func getTmp(bucket string, key string) string {
	if tmpData[bucket] == nil {
		return ""
	}
	dataToLoad := tmpData[bucket][key]
	if dataToLoad.validUntil.Before(time.Now()) {
		delete(tmpData[bucket], key)
		return ""
	}
	return dataToLoad.data
}

func delTmp(bucket string, key string) {
	if tmpData[bucket] == nil {
		return
	}
	delete(tmpData[bucket], key)
}

