package service

import (
	"context"
	"errors"
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
		FindOrCreateNode(ctx *context.Context, txn *dgo.Txn, currentPage *Page) (uid string, err error)
		CheckPredicate(ctx *context.Context, txn *dgo.Txn, parentUid string, childUid string) (exists bool, err error)
		CheckOrCreatePredicate(ctx *context.Context, txn *dgo.Txn, parentUid string, childUid string) (exists bool, err error)
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
		var conn *grpc.ClientConn
		connString := fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port)
		conn, _ = grpc.Dial(connString, grpc.WithInsecure())
		clients = append(clients, dapi.NewDgraphClient(conn))
	}
	s.Dgraph = dgo.NewDgraphClient(clients...)
	DB = s
	return
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
		fmt.Println(err)
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
	txnUid := uuid.New().String()
	ctx := context.Background()
	var currentUid string
	if currentPage.Url != "" {
		for currentUid == "" {
			txn := store.NewTxn()
			currentUid, err = store.FindOrCreateNode(&ctx, txn, currentPage)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = APILogger.LogError(
						"msg", err.Error(),
						"context", "create current page",
						"url", currentPage.Url,
						"txnUid", txnUid,
					)
					return
				}
				continue
			}
		}
	}
	if currentPage.Parent != nil {
		var parentUid string
		for parentUid == "" {
			txn := store.NewTxn()
			parentUid, err = store.FindOrCreateNode(&ctx, txn, currentPage.Parent)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = APILogger.LogError(
						"msg", err.Error(),
						"context", "create parent page",
						"url", currentPage.Parent.Url,
						"txnUid", txnUid,
					)
					return
				}
				continue
			}
		}
		attempts := 10
		for attempts > 0 {
			attempts--
			txn := store.NewTxn()
			success, err := store.CheckOrCreatePredicate(&ctx, txn, parentUid, currentUid)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = APILogger.LogError(
						"context", "create predicate",
						"msg", err.Error(),
						"parentUid", parentUid,
						"childUid", currentUid,
						"txnUid", txnUid,
					)
					break
				}
				continue
			}
			if success {
				break
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
		return
	}
	currentPage, err = deserializePage(resp.Json)

	if currentPage != nil {
		if currentPage.MaxDepth() < depth {
			return nil, errors.New("Depth does not match dgraph result.")
		}
	}
	return
}

// FindOrCreateNode function checks for page, creates if doesn't exist.
func (store *DbStore) FindOrCreateNode(ctx *context.Context, txn *dgo.Txn, currentPage *Page) (uid string, err error) {
	var assigned *dapi.Assigned
	var p []byte
	var resultPage *Page
	resultPage, err = store.FindNode(ctx, txn, currentPage.Url, 0)
	if err != nil {
		if !strings.Contains(err.Error(), "Depth does not match dgraph result.") {
			return
		}
	} else if resultPage != nil {
		uid = resultPage.Uid
	}
	if uid == "" {
		p, _ = serializePage(currentPage)
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
	exists, err = DeserializePredicate(resp.Json)
	return
}

// CheckOrCreatePredicate function checks for edge, creates if doesn't exist.
func (store *DbStore) CheckOrCreatePredicate(ctx *context.Context, txn *dgo.Txn, parentUid string, childUid string) (exists bool, err error) {
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
		if err != nil {
			txn.Discard(*ctx)
			return
		}
		txn.Commit(*ctx)
		txn.Discard(*ctx)
	}
	return
}
