// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"github.com/datacratic/goklog/klog"

	"encoding/json"
	"log"
)

type KlogHandler struct{}

func (KlogHandler) HandleMeters(values map[string]float64) {
	body, err := json.Marshal(values)
	if err != nil {
		log.Panicf("unable to marshal meter values: %s", err)
	}

	klog.KPrint("meter.klog.info", string(body))
}
