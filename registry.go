package miaosic

import (
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
)

var Requester dphttp.IRequester = deepcolor.NewRestyRequester()

func init() {
	deepcolor.SetDefaultRequester(Requester)
}

var _providers map[string]MediaProvider = make(map[string]MediaProvider)

func RegisterProvider(provider MediaProvider) {
	if _, ok := _providers[provider.GetName()]; ok {
		panic("provider " + provider.GetName() + " already exists")
		return
	}
	_providers[provider.GetName()] = provider
}

func ListAvailableProviders() []string {
	var names []string
	for name := range _providers {
		names = append(names, name)
	}
	return names
}
