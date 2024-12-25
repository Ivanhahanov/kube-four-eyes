package storage

import (
	"context"
	"log"
	"time"
	"webhook/pkg/helpers"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
)

type Store struct {
	cli *clientv3.Client
}

func DB() *Store {
	return &Store{
		cli: NewClient(),
	}
}

func (s *Store) Close() {
	s.cli.Close()
}

func NewClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{helpers.GetEnv("ETCD_ENDPOINT", "localhost:2379")},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	return cli

}
func (s *Store) Old(k, v string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := s.cli.Put(ctx, k, v)
	cancel()
	return err
}

func (s *Store) Put(k, v string, leaseId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := s.cli.Put(ctx, k, v, clientv3.WithLease(clientv3.LeaseID(leaseId)))
	cancel()
	return err
}

func (s *Store) PutTemporary(k, v string, timeout int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	lease, err := s.cli.Grant(ctx, timeout)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	_, err = s.cli.Put(ctx, k, v, clientv3.WithLease(lease.ID))
	cancel()
	return int64(lease.ID), err
}

func (s *Store) Get(k string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := s.cli.Get(ctx, k)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	if len(resp.Kvs) == 1 {
		return resp.Kvs[0].Value
	}
	return nil
}

func (s *Store) GetMany(k string) []*mvccpb.KeyValue {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := s.cli.Get(ctx, k, clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	return resp.Kvs
}

// func (s *Store) GetMany(k string) []*mvccpb.KeyValue {
// 	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
// 	resp, err := s.cli.Get(ctx, k, clientv3.WithPrefix())
// 	cancel()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var leasedKV []*mvccpb.KeyValue
// 	//filter by leaseID
// 	for _, r := range resp.Kvs {
// 		fmt.Println(r)
// 		if r.Lease == leaseId {
// 			leasedKV = append(leasedKV, r)
// 		}
// 	}
// 	return leasedKV
// }

func (s *Store) Delete(k string) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := s.cli.Delete(ctx, k)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
}
