package database

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/dgo"
	"go-clamber/page"
	"log"
)

func (store *DbStore) GetPageByUid(ctx context.Context, txn *dgo.Txn, uid string) (currentPage *page.Page, err error) {
	variables := map[string]string{"$id": uid}
	q := `query withvar($id: string){
			result(func: uid($id)) {
				url
				body
				timestamp
				links.To {
					url
				}
			}
		}`
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
		return
	}
	currentPage, err = ConvertJsonToPage(resp.Json)
	return
}

func (store *DbStore) GetPageByUrl(ctx context.Context, txn *dgo.Txn, Url string) (currentPage *page.Page, err error) {
	variables := map[string]string{"$url": Url}
	q := `query withvar($url: string){
			result(func: eq(url, $url)) {
				url
				body
				timestamp
				links.To {
					url
				}
			}
		}`
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
		return
	}
	currentPage, err = ConvertJsonToPage(resp.Json)
	return
}

func (store *DbStore) GetUidByUrl(ctx context.Context, txn *dgo.Txn, Url string) (uid string, err error) {
	variables := map[string]string{"$url": Url}
	q := `query withvar($url: string){
			result(func: eq(url, $url)) {
				uid
			}
		}`
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
		return
	}
	jsonMap := make(map[string][]JSONPage)
	err = json.Unmarshal(resp.Json, &jsonMap)
	if jsonPage := jsonMap["result"]; len(jsonPage) > 0 {
		uid = jsonPage[0].Uid
	}
	return
}
