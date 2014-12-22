// Copyright (c) 2014 Datacratic. All rights reserved.

package meter

import (
	"github.com/datacratic/gorest/rest"

	"strings"
	"sync"
)

const DefaultPathREST = "/debug/meter"

type RESTHandler struct {
	PathPrefix string

	mutex sync.Mutex
	last  map[string]float64
}

func (handler *RESTHandler) RESTRoutes() rest.Routes {
	prefix := handler.PathPrefix
	if prefix == "" {
		prefix = DefaultPathREST + "/keys"
	}

	return []*rest.Route{
		rest.NewRoute(prefix, "GET", handler.Get),
		rest.NewRoute(prefix+"/filter/:substr", "GET", handler.GetSubstr),
	}
}

func (handler *RESTHandler) HandleMeters(meters map[string]float64) {
	handler.mutex.Lock()

	handler.last = meters

	handler.mutex.Unlock()
}

func (handler *RESTHandler) Get() map[string]float64 {
	handler.mutex.Lock()

	result := handler.last

	handler.mutex.Unlock()

	return result
}

func (handler *RESTHandler) GetSubstr(substr string) map[string]float64 {
	handler.mutex.Lock()

	result := make(map[string]float64)

	for key, value := range handler.last {
		if strings.Index(key, substr) >= 0 {
			result[key] = value
		}
	}

	handler.mutex.Unlock()

	return result
}
