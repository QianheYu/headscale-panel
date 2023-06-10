package config

import (
	"fmt"
	"headscale-panel/model"
	"sync/atomic"
	"testing"
)

var valueTest = &atomic.Value{}

func BenchmarkAtomic(b *testing.B) {
	b.Run("store", store)
	b.Run("load", load)
	fmt.Println(valueTest.Load().(*model.HeadscaleConfig))
}

func store(b *testing.B) {
	fmt.Printf("b.N is %d\n", b.N)
	data := &model.HeadscaleConfig{
		GRPCListenAddr: "localhost:50443",
		ApiKey:         "adfadf",
		Insecure:       false,
		AccessControl:  "/tmp/headscale/acl/",
		DNS: model.DNSConfig{
			BaseDomain: "localhost",
		},
		OIDC: model.OIDC{
			OnlyStartIfOIDCIsAvailable: true,
			Issuer:                     "https://localhost:50443/",
			ClientID:                   "client-id",
			ClientSecret:               "client-secret",
			//RedirectURL:                "https://localhost:50443/callback",
			//Scope:                      "openid profile email",
		},
	}
	for i := 0; i < b.N; i++ {
		valueTest.Store(data)
	}
}

func load(b *testing.B) {
	fmt.Printf("b.N is %d\n", b.N)
	for i := 0; i < b.N; i++ {
		_ = valueTest.Load().(*model.HeadscaleConfig)
	}
}
