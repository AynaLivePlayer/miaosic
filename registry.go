package miaosic

import (
	"sort"

	"github.com/aynakeya/deepcolor"
	"github.com/aynakeya/deepcolor/dphttp"
)

var Requester dphttp.IRequester = deepcolor.NewRestyRequester()

func init() {
	Requester.Config().Timeout = 3
	deepcolor.SetDefaultRequester(Requester)
}

type Registry struct {
	providers map[string]MediaProvider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]MediaProvider),
	}
}

func (r *Registry) RegisterProvider(provider MediaProvider) {
	if _, ok := r.providers[provider.GetName()]; ok {
		panic("provider " + provider.GetName() + " already exists")
	}
	r.providers[provider.GetName()] = provider
}

func (r *Registry) UnregisterProvider(name string) {
	_, ok := r.providers[name]
	if ok {
		delete(r.providers, name)
	}
}

func (r *Registry) UnregisterAllProvider() {
	r.providers = make(map[string]MediaProvider)
}

func (r *Registry) GetProvider(name string) (MediaProvider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

func (r *Registry) ListAvailableProviders() []string {
	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}
