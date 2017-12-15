package counter_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/csouls/counter_test/sharding_counter"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

func ShardingDatastoreHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	if err := sharding_counter.Reset(ctx, scname()); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	bc, err := sharding_counter.Count(ctx, scname())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Write([]byte(fmt.Sprintf("Before count: %d\n", bc)))

	result := testing.Benchmark(func(b *testing.B) {
		rw.Write([]byte(fmt.Sprintf("benchmark count: %v\n", b.N)))
		b.ResetTimer()

		if err := shardingDatastoreIncrement(ctx, b.N); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	rw.Write([]byte(fmt.Sprintf("%v\n", result)))
}

func ShardingDatastoreInfoHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	bc, err := sharding_counter.Count(ctx, scname())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Write([]byte(fmt.Sprintf("count: %d\n", bc)))
}

func shardingDatastoreIncrement(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		if err := sharding_counter.Increment(ctx, scname()); err != nil {
			return err
		}
	}

	return nil
}

func scname() string {
	return "ShardingCounter"
}
