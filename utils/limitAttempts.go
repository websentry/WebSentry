package utils

import (
	"github.com/patrickmn/go-cache"
	"time"
	"sync"
	)

const (
	limitLoginUsername = uint(1000)
	limitLoginIP = uint(2000)
	loginPurgeDuration = 1 * time.Hour

	// The expiration duration of the login limits
	LoginExpireDuration = 10 * time.Minute
)

type SafeCache struct {
	cache *cache.Cache
	mux sync.Mutex
}

var loginC SafeCache

func init() {
	loginC.cache = cache.New(LoginExpireDuration, loginPurgeDuration)
}

// CheckLoginAvailability checks if the login username or its ip address
// excess the limit of attempts
func CheckLogInAvailability(u, ip string) bool {
	loginC.mux.Lock()
	defer loginC.mux.Unlock()

	uNum, uFound := loginC.cache.Get(u)
	if uFound {
		n := uNum.(uint)

		// check if exceeds
		if n > limitLoginUsername {
			// update expire time
			loginC.cache.Set(u, n, cache.DefaultExpiration)
			return false
		}

		// increment
		loginC.cache.Set(u, uint(n+1), cache.DefaultExpiration)
	} else {
		loginC.cache.Set(u, uint(1), cache.DefaultExpiration)
	}

	ipNum, ipFound := loginC.cache.Get(ip)
	if ipFound {
		n := ipNum.(uint)
		if n > limitLoginIP {
			loginC.cache.Set(ip, n, cache.DefaultExpiration)
			return false
		}
		loginC.cache.Set(ip, uint(n+1), cache.DefaultExpiration)
	} else {
		loginC.cache.Set(ip, uint(1), cache.DefaultExpiration)
	}

	return true
}
