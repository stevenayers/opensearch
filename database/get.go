package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"go-clamber/page"
)

func (store *DbStore) GetPageByUid(ctx context.Context, txn *dgo.Txn, uid string) (currentPage *page.Page, err error) {
	variables := map[string]string{"$id": uid}
	q := `query withvar($id: string){
			result(func: uid($id)) {
				url
				body
				timestamp
			}
		}`
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		fmt.Print(err)
		return
	}
	currentPage, err = ConvertJsonToPage(resp.Json)
	return
}

func (store *DbStore) PredicateExists(ctx context.Context, parentUid string, childUid string) (exists bool, err error) {
	txn := store.NewTxn()
	defer txn.Discard(ctx)
	variables := map[string]string{"$parentUid": parentUid, "$childUid": childUid}
	q := `query withvar($parentUid: string, $childUid: string){
			edges(func: uid($parentUid)) {
				matching: count(links) @filter(uid($childUid))
			  }
			}`
	var resp *api.Response
	resp, err = txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		return
	}
	jsonMap := make(map[string][]PredicateCount)
	err = json.Unmarshal(resp.Json, &jsonMap)
	if err != nil {
		return
	}
	edges := jsonMap["edges"]
	if len(edges) > 0 {
		exists = edges[0].Matching > 0
	} else {
		exists = false
	}
	return
}

func (store *DbStore) GetPageByUrl(ctx context.Context, txn *dgo.Txn, Url string) (currentPage *page.Page, err error) {
	variables := map[string]string{"$url": Url}
	q := `query withvar($url: string){
			result(func: eq(url, $url)) {
				url
				body
				timestamp
				links {
					url
				}
			}
		}`
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		fmt.Print(err)
		return
	}
	currentPage, err = ConvertJsonToPage(resp.Json)
	return
}

func (store *DbStore) GetUidByUrl(ctx *context.Context, txn *dgo.Txn, Url string) (uid string, err error) {
	variables := map[string]string{"$url": Url}
	q := `query withvar($url: string){
			result(func: eq(url, $url)) {
				uid
			}
		}`
	resp, err := txn.QueryWithVars(*ctx, q, variables)
	if err != nil {
		fmt.Print(err)
		return
	}
	jsonMap := make(map[string][]JSONPage)
	err = json.Unmarshal(resp.Json, &jsonMap)
	jsonPage := jsonMap["result"]
	if len(jsonPage) > 0 {
		uid = jsonPage[0].Uid
	}
	return
}
