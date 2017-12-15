package counter_test

import (
	"fmt"
	"net/http"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func DatastoreHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	if err := datastoreReset(ctx); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	result := testing.Benchmark(func(b *testing.B) {
		rw.Write([]byte(fmt.Sprintf("benchmark count: %v\n", b.N)))
		b.ResetTimer()

		if err := datastoreIncrement(ctx, b.N); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	rw.Write([]byte(fmt.Sprintf("%v\n", result)))
}

func DatastoreInfoHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	rw.Write([]byte(fmt.Sprintf("Before count: %d\n", datastoreCount(ctx))))
}

type countEntity struct {
	Count int
}

func datastoreIncrement(ctx context.Context, n int) error {
	var count countEntity

	for i := 0; i < n; i++ {
		err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			key := dkey(ctx)
			err := datastore.Get(ctx, key, &count)
			if err != nil && err != datastore.ErrNoSuchEntity {
				return err
			}
			count.Count++
			_, err = datastore.Put(ctx, key, &count)
			return err
		}, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func datastoreCount(ctx context.Context) int {
	count := countEntity{Count: 0}
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		err := datastore.Get(ctx, dkey(ctx), &count)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		log.Warningf(ctx, "%v", err)
	}
	return count.Count
}

func datastoreReset(ctx context.Context) error {
	count := countEntity{Count: 0}
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		_, err := datastore.Put(ctx, dkey(ctx), &count)
		return err
	}, nil)
}

func dkey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, "SimpleCounter", "SimpleCounter", 0, nil)
}
