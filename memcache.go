package counter_test

import (
	"fmt"
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

func MemcachedHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	if err := memcacheReset(ctx); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	result := testing.Benchmark(func(b *testing.B) {
		rw.Write([]byte(fmt.Sprintf("benchmark count: %v\n", b.N)))
		b.ResetTimer()

		if err := memcacheIncrement(ctx, b.N); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	rw.Write([]byte(fmt.Sprintf("%v\n", result)))
	rw.Write([]byte(fmt.Sprintf("After count: %d\n", memcacheCount(ctx))))
}

func memcacheCount(ctx context.Context) int {
	var total int
	if _, err := memcache.JSON.Get(ctx, mkey(), &total); err == nil {
		return total
	} else {
		log.Warningf(ctx, "%v", err)
	}
	return 0
}

func memcacheIncrement(ctx context.Context, c int) error {
	for i := 0; i < c; i++ {
		_, err := memcache.IncrementExisting(ctx, mkey(), 1)
		if err != nil {
			return err
		}
	}

	return nil
}

func memcacheReset(ctx context.Context) error {
	return memcache.JSON.Set(ctx, &memcache.Item{
		Key:    mkey(),
		Object: 0,
	})
}

func mkey() string {
	return "counter"
}
