package miaosic

import (
	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
	"sort"
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

func GetProvider(name string) (MediaProvider, bool) {
	provider, ok := _providers[name]
	return provider, ok
}

func ListAvailableProviders() []string {
	var names []string
	for name := range _providers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}
