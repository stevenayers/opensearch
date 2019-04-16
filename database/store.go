package database

import (
	"context"
	"fmt"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"go-clamber/page"
	"google.golang.org/grpc"
)

type (
	Results []page.Page

	Store interface {
		Create(currentPage *page.Page) (err error)
		GetOrCreatePage(ctx *context.Context, currentPage *page.Page) (uid string, err error)
		DeleteAll() (err error)
	}

	DbStore struct {
		*dgo.Dgraph
		Connection *grpc.ClientConn
	}
)

func (store *DbStore) SetSchema() (err error) {
	op := &api.Operation{}
	op.Schema = `
	url: string @index(exact) @upsert .
	timestamp: int .
	body: string .
    links: uid @count .
	`
	ctx := context.TODO()
	err = store.Alter(ctx, op)
	if err != nil {
		fmt.Print(err)
	}
	return
}

func (store *DbStore) DeleteAll() (err error) {
	err = store.Alter(context.Background(), &api.Operation{DropAll: true})
	return
}

func InitStore(s *DbStore) {
	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		fmt.Print(err)
	}
	s.Connection = conn
	s.Dgraph = dgo.NewDgraphClient(api.NewDgraphClient(conn))
	DB = s
}

var DB Store
