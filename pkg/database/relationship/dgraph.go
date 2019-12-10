package relationship

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/page"
	"google.golang.org/grpc"
	"strconv"
)

type (

	// Store holds dgraph client and connections
	Store struct {
		DB         *dgo.Dgraph
		Connection []*grpc.ClientConn
	}
)

// Connect function initiates connections to database
func (store *Store) Connect() {
	var clients []api.DgraphClient
	for _, connConfig := range config.AppConfig.Database.Connections {
		var conn *grpc.ClientConn
		connString := fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port)
		conn, _ = grpc.Dial(connString, grpc.WithInsecure())
		clients = append(clients, api.NewDgraphClient(conn))
	}
	store.DB = dgo.NewDgraphClient(clients...)
	return
}

// SetSchema function sets the schema for dgraph (mainly for tests)
func (store *Store) SetSchema() (err error) {
	op := &api.Operation{}
	op.Schema = `
	url: string @index(hash) @upsert .
	timestamp: int .
	deoth: int .
	status_code: int .
    links: [uid] @count @reverse .
	`
	ctx := context.TODO()
	err = store.DB.Alter(ctx, op)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// DeleteAll function deletes all data in database
func (store *Store) DeleteAll() (err error) {
	err = store.DB.Alter(context.Background(), &api.Operation{DropAll: true})
	return
}

// FindNode function finds Page by URL and depth
func (store *Store) FindNode(ctx *context.Context, Url string, depth int) (currentPage *page.Page, err error) {
	txn := store.DB.NewTxn()
	defer txn.Discard(*ctx)
	var resp *api.Response
	queryDepth := strconv.Itoa(depth + 1)
	v := map[string]string{"$url": Url}
	q := `query withvar($url: string, $depth: int){
			result(func: eq(url, $url)) @recurse(depth: ` + queryDepth + `, loop: false){
 				uid
				url
				timestamp
    			links
			}
		}`

	req := &api.Request{
		Query: q,
		Vars:  v,
	}
	resp, err = txn.Do(*ctx, req)
	if err != nil {
		return
	}
	currentPage, err = page.DeserializeJsonPage(resp.Json)
	if currentPage != nil {
		if currentPage.MaxDepth() < depth {
			return nil, errors.New("Depth does not match dgraph result.")
		}
	}
	return
}

// FindOrCreateNode function checks for page, creates if doesn't exist.F
func (store *Store) FindOrCreateNode(ctx *context.Context, currentPage *page.Page) (uid string, err error) {
	txn := store.DB.NewTxn()
	defer txn.Discard(*ctx)
	var resp *api.Response
	v := map[string]string{"$url": currentPage.Url}
	q := `query withvar($url: string){
			page as var(func: eq(url, $url))
		}`
	currentPage.Uid = "_:cp"
	p, _ := page.SerializeJsonPage(currentPage)
	req := &api.Request{
		Query:     q,
		Vars:      v,
		Mutations: []*api.Mutation{{SetJson: p, Cond: `@if(eq(len(page), 0))`}},
		CommitNow: true,
	}
	resp, err = txn.Do(*ctx, req)
	if err != nil {
		return
	}
	if resp.Uids != nil {
		uid = resp.Uids["cp"]
	} else {
		var resultPage *page.Page
		txn = store.DB.NewTxn()
		resultPage, err = store.FindNode(ctx, currentPage.Url, 0)
		uid = resultPage.Uid
		if err != nil {
			return
		}
	}
	return
}

// CheckPredicate function checks to see if edge exists
func (store *Store) CheckPredicate(ctx *context.Context, parentUid string, childUid string) (exists bool, err error) {
	txn := store.DB.NewTxn()
	variables := map[string]string{"$parentUid": parentUid, "$childUid": childUid}
	q := `query withvar($parentUid: string, $childUid: string){
			edges(func: uid($parentUid)) {
				matching: count(links) @filter(uid($childUid))
			  }
			}`
	var resp *api.Response
	resp, err = txn.QueryWithVars(*ctx, q, variables)
	if err != nil {
		return
	}
	exists, err = page.DeserializePredicate(resp.Json)
	return
}

// CheckOrCreatePredicate function checks for edge, creates if doesn't exist.
func (store *Store) CheckOrCreatePredicate(ctx *context.Context, parentUid string, childUid string) (exists bool, err error) {
	txn := store.DB.NewTxn()
	defer txn.Discard(*ctx)
	v := map[string]string{"$parentUid": parentUid, "$childUid": childUid}
	q := `query withvar($parentUid: string, $childUid: string){
			link as var(func: uid($parentUid)) @filter(uid_in(links, $childUid))
		}`
	req := &api.Request{
		Query: q,
		Vars:  v,
		Mutations: []*api.Mutation{{
			Cond: `@if(eq(len(link), 0))`,
			Set:  []*api.NQuad{{Subject: parentUid, Predicate: "links", ObjectId: childUid}},
		}},
		CommitNow: true,
	}
	_, err = txn.Do(*ctx, req)
	if err != nil {
		return
	}
	exists, err = store.CheckPredicate(ctx, parentUid, childUid)
	if err != nil {
		return
	}
	return
}
