// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"github.com/datacratic/goklog/klog"
	"github.com/datacratic/gorest/rest"

	"net/http"
	"net/url"
	"sync"
	"time"
)

type HTTPHandler struct {
	URL    string
	Method string

	HTTPClient *http.Client

	initialize sync.Once
	client     *rest.Client
}

func (handler *HTTPHandler) Init() {
	handler.initialize.Do(handler.init)
}

func (handler *HTTPHandler) init() {
	if handler.URL == "" {
		klog.KFatal("meter.http.init.error", "no URL configured")
	}

	if handler.Method == "" {
		klog.KFatal("meter.http.init.error", "no HTTP method configured")
	}

	if _, err := url.Parse(handler.URL); err != nil {
		klog.KFatalf("meter.http.init.error", "invalid URL '%s': %s", handler.URL, err)
	}

	if handler.HTTPClient == nil {
		handler.HTTPClient = http.DefaultClient
	}

	handler.client = &rest.Client{
		Client: handler.HTTPClient,
		Root:   handler.URL,
	}
}

func (handler *HTTPHandler) HandleMeters(values map[string]float64) {
	handler.Init()

	var body struct {
		Timestamp time.Time          `json:"timestamp"`
		Values    map[string]float64 `json:"values"`
	}

	body.Timestamp = time.Now()
	body.Values = values

	resp := handler.client.NewRequest(handler.Method).SetBody(body).Send()

	if err := resp.GetBody(nil); err != nil {
		klog.KPrintf("meter.http.send.error", "unable to send metrics: %s", err)
	}
}
