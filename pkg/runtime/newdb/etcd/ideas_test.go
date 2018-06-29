package etcd

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/stretchr/testify/assert"
)

//func TestGenerate(t *testing.T) {
//	client, err := etcd.New(etcd.Config{
//		Endpoints:   []string{"localhost:2379"},
//		DialTimeout: 5 * time.Second,
//	})
//	assert.NoError(t, err)
//	assert.NotNil(t, client)
//
//	for i := 0; i < 10000; i++ {
//		_, err := client.Put(context.TODO(), fmt.Sprintf("val_%d", i), "value")
//		assert.NoError(t, err)
//	}
//}

func BenchmarkIdeasSortedGet(b *testing.B) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(b, err)
	assert.NotNil(b, client)
	defer client.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(context.TODO(), "val_", etcd.WithPrefix(), etcd.WithSort(etcd.SortByKey, etcd.SortDescend), etcd.WithLimit(1))
		assert.NoError(b, err)
		assert.Equal(b, "val_9999", string(resp.Kvs[0].Key))
	}
}

func BenchmarkIdeasIndexedGet(b *testing.B) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(b, err)
	assert.NotNil(b, client)
	defer client.Close()

	client.Put(context.TODO(), "index", "val_9999")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(context.TODO(), "index")
		assert.NoError(b, err)
		latestVersionKey := string(resp.Kvs[0].Value)
		assert.Equal(b, "val_9999", latestVersionKey)

		resp, err = client.Get(context.TODO(), latestVersionKey)
		assert.NoError(b, err)
		assert.Equal(b, "val_9999", string(resp.Kvs[0].Key))
	}
}

func BenchmarkIdeasTransactionIndexedGet(b *testing.B) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(b, err)
	assert.NotNil(b, client)
	defer client.Close()

	client.Put(context.TODO(), "index", "val_9999")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(context.TODO(), "index")
		assert.NoError(b, err)
		latestVersionKey := string(resp.Kvs[0].Value)
		assert.Equal(b, "val_9999", latestVersionKey)

		txResp, err := client.Txn(context.TODO()).If(
			etcd.Compare(etcd.ModRevision("index"), "=", resp.Kvs[0].ModRevision),
		).Then(
			etcd.OpGet(latestVersionKey),
		).Commit()

		assert.NoError(b, err)
		assert.Equal(b, "val_9999", string(txResp.Responses[0].GetResponseRange().Kvs[0].Key))
	}
}

func BenchmarkIdeasSTM(b *testing.B) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	assert.NoError(b, err)
	assert.NotNil(b, client)
	defer client.Close()

	// set up "accounts"
	totalAccounts := 5
	accountBalance := 100
	for i := 0; i < totalAccounts; i++ {
		k := fmt.Sprintf("accts/%d", i)
		_, err = client.Put(context.TODO(), k, fmt.Sprintf("%d", accountBalance))
		assert.NoError(b, err)
	}

	b.ResetTimer()

	exchange := func(stm concurrency.STM) error {
		from, to := rand.Intn(totalAccounts), rand.Intn(totalAccounts)
		if from == to {
			// nothing to do
			return nil
		}
		// read values
		fromK, toK := fmt.Sprintf("accts/%d", from), fmt.Sprintf("accts/%d", to)
		fromV, toV := stm.Get(fromK), stm.Get(toK)
		fromInt, toInt := 0, 0
		fmt.Sscanf(fromV, "%d", &fromInt)
		fmt.Sscanf(toV, "%d", &toInt)

		// transfer amount
		xfer := fromInt / 2
		fromInt, toInt = fromInt-xfer, toInt+xfer

		// write back
		stm.Put(fromK, fmt.Sprintf("%d", fromInt))
		stm.Put(toK, fmt.Sprintf("%d", toInt))
		return nil
	}

	// concurrently exchange values between accounts
	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			_, serr := concurrency.NewSTM(client, exchange)
			assert.NoError(b, serr)
		}()
	}
	wg.Wait()

	// confirm account sum matches sum from beginning.
	sum := 0
	accts, err := client.Get(context.TODO(), "accts/", etcd.WithPrefix())
	assert.NoError(b, err)
	for _, kv := range accts.Kvs {
		v := 0
		fmt.Sscanf(string(kv.Value), "%d", &v)
		sum += v
	}

	//fmt.Println("account sum is", sum)

	assert.Equal(b, totalAccounts*accountBalance, sum)
}
