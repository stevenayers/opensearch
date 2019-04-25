package service

import (
	"context"
	"fmt"
	"github.com/dgraph-io/dgo"
	dapi "github.com/dgraph-io/dgo/protos/api"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"strconv"
	"strings"
)

type (
	// Store is interface for database method
	Store interface {
		SetSchema() (err error)
		DeleteAll() (err error)
		Create(currentPage *Page) (err error)
		FindNode(ctx *context.Context, txn *dgo.Txn, url string, depth int) (currentPage *Page, err error)
		FindOrCreateNode(ctx *context.Context, currentPage *Page) (uid string, err error)
		CheckPredicate(ctx *context.Context, txn *dgo.Txn, parentUid string, childUid string) (exists bool, err error)
		CheckOrCreatePredicate(ctx *context.Context, parentUid string, childUid string) (err error)
	}

	// DbStore holds dgraph client and connections
	DbStore struct {
		*dgo.Dgraph
		Connection []*grpc.ClientConn
	}
)

// DB makes a global store
var DB Store

// Connect function initiates connections to database
func Connect(s *DbStore, dbConfig DatabaseConfig) {
	var clients []dapi.DgraphClient
	for _, connConfig := range dbConfig.Connections {
		connString := fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port)
		conn, err := grpc.Dial(connString, grpc.WithInsecure())
		if err != nil {
			fmt.Print(err)
		}
		clients = append(clients, dapi.NewDgraphClient(conn))
	}
	s.Dgraph = dgo.NewDgraphClient(clients...)
	DB = s
}

// SetSchema function sets the schema for dgraph (mainly for tests)
func (store *DbStore) SetSchema() (err error) {
	op := &dapi.Operation{}
	op.Schema = `
	url: string @index(hash) @upsert .
	timestamp: int .
    links: uid @count @reverse .
	`
	ctx := context.TODO()
	err = store.Alter(ctx, op)
	if err != nil {
		fmt.Print(err)
	}
	return
}

// DeleteAll function deletes all data in database
func (store *DbStore) DeleteAll() (err error) {
	err = store.Alter(context.Background(), &dapi.Operation{DropAll: true})
	return
}

// Create function checks for current page, creates if doesn't exist. Checks for parent page, creates if doesn't exist. Checks for edge
// between them, creates if doesn't exist.
func (store *DbStore) Create(currentPage *Page) (err error) {
	uid := uuid.New().String()
	var currentUid string
	ctx := context.Background()
	currentUid, err = store.FindOrCreateNode(&ctx, currentPage)
	if err != nil {
		APILogger.LogError(
			"msg", err.Error(),
			"context", "create current page",
			"url", currentPage.Url,
			"uid", uid,
		)
		return
	}
	if currentPage.Parent != nil {
		var parentUid string
		parentUid, err = store.FindOrCreateNode(&ctx, currentPage.Parent)
		if err != nil {
			APILogger.LogError(
				"msg", err.Error(),
				"context", "create parent page",
				"url", currentPage.Parent.Url,
				"uid", uid,
			)
			return
		}
		err = store.CheckOrCreatePredicate(&ctx, parentUid, currentUid)
		if err != nil {
			APILogger.LogError(
				"context", "create predicate",
				"msg", err.Error(),
				"parentUid", parentUid,
				"childUid", currentUid,
				"uid", uid,
			)
			if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
				!strings.Contains(err.Error(), "Transaction is too old") {
				return
			}
		}
	}
	return
}

// FindNode function finds Page by URL and depth
func (store *DbStore) FindNode(ctx *context.Context, txn *dgo.Txn, Url string, depth int) (currentPage *Page, err error) {
	queryDepth := strconv.Itoa(depth + 1)
	variables := map[string]string{"$url": Url}
	q := `query withvar($url: string, $depth: int){
			result(func: eq(url, $url)) @recurse(depth: ` + queryDepth + `, loop: false){
 				uid
				url
				timestamp
    			links
			}
		}`
	resp, err := txn.QueryWithVars(*ctx, q, variables)
	if err != nil {
		fmt.Print(err)
		return
	}
	currentPage, err = deserializePage(resp.Json)

	if currentPage != nil {
		if currentPage.MaxDepth() < depth {
			return nil, nil
		}
	}
	return
}

// FindOrCreateNode function checks for page, creates if doesn't exist.
func (store *DbStore) FindOrCreateNode(ctx *context.Context, currentPage *Page) (uid string, err error) {
	for uid == "" {
		var assigned *dapi.Assigned
		var p []byte
		var resultPage *Page
		txn := store.NewTxn()
		resultPage, err = store.FindNode(ctx, txn, currentPage.Url, 0)
		if err != nil {
			return
		} else if resultPage != nil {
			uid = resultPage.Uid
		}
		if uid == "" {
			p, err = serializePage(currentPage)
			if err != nil {
				return
			}
			mu := &dapi.Mutation{}
			mu.SetJson = p
			assigned, err = txn.Mutate(*ctx, mu)
			if err != nil {
				return
			}
		}
		err = txn.Commit(*ctx)
		txn.Discard(*ctx)
		if uid == "" && err == nil {
			uid = assigned.Uids["blank-0"]
		}
		if uid != "" {
			currentPage.Uid = uid
		}

	}
	return
}

// CheckPredicate function checks to see if edge exists
func (store *DbStore) CheckPredicate(ctx *context.Context, txn *dgo.Txn, parentUid string, childUid string) (exists bool, err error) {
	variables := map[string]string{"$parentUid": parentUid, "$childUid": childUid}
	q := `query withvar($parentUid: string, $childUid: string){
			edges(func: uid($parentUid)) {
				matching: count(links) @filter(uid($childUid))
			  }
			}`
	var resp *dapi.Response
	resp, err = txn.QueryWithVars(*ctx, q, variables)
	if err != nil {
		return
	}
	exists, err = deserializePredicate(resp.Json)
	return
}

// CheckOrCreatePredicate function checks for edge, creates if doesn't exist.
func (store *DbStore) CheckOrCreatePredicate(ctx *context.Context, parentUid string, childUid string) (err error) {
	attempts := 10
	exists := false
	for !exists && attempts > 0 {
		attempts--
		txn := store.NewTxn()
		exists, err = store.CheckPredicate(ctx, txn, parentUid, childUid)
		if err != nil {
			return
		}
		if !exists {
			_, err = txn.Mutate(*ctx, &dapi.Mutation{
				Set: []*dapi.NQuad{{
					Subject:   parentUid,
					Predicate: "links",
					ObjectId:  childUid,
				}}})
			if err != nil && attempts <= 0 {
				txn.Discard(*ctx)
				return
			}
			txn.Commit(*ctx)
			txn.Discard(*ctx)
		}
	}
	return
}
