package database

import (
	"context"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/dgraph-io/dgo/y"
	"go-clamber/page"
	"log"
)

func (store *DbStore) Create(currentPage *page.Page) (err error) {
	var currentUid string
	ctx := context.Background()
	currentUid, err = store.GetOrCreatePage(&ctx, currentPage)
	if err != nil {
		log.Printf("[ERROR] context: create current page (%s) - message: %s\n", currentPage.Url, err.Error())
		return
	}
	if currentPage.Parent != nil {
		var parentUid string
		parentUid, err = store.GetOrCreatePage(&ctx, currentPage.Parent)
		if err != nil {
			log.Printf("[ERROR] context: create parent page (%s) - message: %s\n", currentPage.Parent.Url, err.Error())
			return
		}
		err = store.CreatePredicate(&ctx, parentUid, currentUid)
		if err != nil {
			log.Printf("[ERROR] create predicate (%s -> %s) - message: %s\n", parentUid, currentUid, err.Error())
			return
		}
	}
	return
}

func (store *DbStore) GetOrCreatePage(ctx *context.Context, currentPage *page.Page) (uid string, err error) {
	for uid == "" {
		var assigned *api.Assigned
		var p []byte
		txn := store.NewTxn()
		uid, err = store.GetUidByUrl(ctx, txn, currentPage.Url)
		if err != nil {
			return
		}
		if uid == "" {
			p, err = ConvertPageToJson(currentPage)
			if err != nil {
				return
			}
			mu := &api.Mutation{}
			mu.SetJson = p
			assigned, err = txn.Mutate(*ctx, mu)
			if err != nil {
				return
			}
		}
		err = txn.Commit(*ctx)
		if uid == "" && err == nil {
			uid = assigned.Uids["blank-0"]
		}
		txn.Discard(*ctx)
	}
	return
}

func (store *DbStore) CreatePredicate(ctx *context.Context, parentUid string, childUid string) (err error) {

	exists, err := store.PredicateExists(*ctx, parentUid, childUid) // false
	if err != nil {
		return
	}
	for !exists {
		txn := store.NewTxn()
		_, err = txn.Mutate(*ctx, &api.Mutation{
			Set: []*api.NQuad{{
				Subject:   parentUid,
				Predicate: "links",
				ObjectId:  childUid,
			}}})
		if err != nil {
			return
		}
		err = txn.Commit(*ctx)
		txn.Discard(*ctx)

		switch err {
		case y.ErrAborted:

		case y.ErrConflict:
			return
		default:
			exists = true
		}
	}
	return
}
