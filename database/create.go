package database

import (
	"context"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"go-clamber/page"
	"log"
)

func (store *DbStore) Create(currentPage *page.Page) (err error) {
	ctx := context.Background()
	txn := store.NewTxn()
	parentUid, err := store.GetUidByUrl(ctx, txn, currentPage.Url)
	if parentUid == "" {
		parentUid, err = store.CreatePage(ctx, txn, currentPage)
	}
	err = txn.Commit(ctx)
	if err != nil {
		txn.Discard(ctx)
	}
	for _, childPage := range currentPage.Children {
		txn := store.NewTxn()
		childUid, err := store.GetUidByUrl(ctx, txn, childPage.Url)
		if err != nil {
			log.Fatal(err)
			return err
		}
		if childUid == "" {
			childUid, err = store.CreatePage(ctx, txn, childPage)
		}
		if err != nil {
			log.Fatal(err)
			return err
		}
		err = txn.Commit(ctx)
		if err != nil {
			txn.Discard(ctx)
		}
		txn = store.NewTxn()

		err = store.CreatePredicate(ctx, txn, parentUid, childUid)
		if err != nil {
			log.Fatal(err)
			return err
		}
		err = txn.Commit(ctx)
		if err != nil {
			txn.Discard(ctx)
		}
	}
	return
}

func (store *DbStore) CreatePage(ctx context.Context, txn *dgo.Txn, currentPage *page.Page) (uid string, err error) {
	p, err := ConvertPageToJson(currentPage)
	mu := &api.Mutation{}
	mu.SetJson = p
	assigned, err := txn.Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
		return
	}
	uid = assigned.Uids["blank-0"]
	return
}

func (store *DbStore) CreatePredicate(ctx context.Context, txn *dgo.Txn, parentUid string, childUid string) (err error) {
	_, err = txn.Mutate(ctx, &api.Mutation{
		Set: []*api.NQuad{
			{
				Subject:   parentUid,
				Predicate: "links.To",
				ObjectId:  childUid,
			},
		}})
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}
