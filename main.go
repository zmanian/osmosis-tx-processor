package main

import (
	"container/heap"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type Item struct {
	value    string
	priority int
	index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

type ConcurrentPriorityQueue struct {
	pq  PriorityQueue
	mtx sync.Mutex
}

func (cpq *ConcurrentPriorityQueue) Push(item *Item) {
	cpq.mtx.Lock()
	defer cpq.mtx.Unlock()
	heap.Push(&cpq.pq, item)
}

func (cpq *ConcurrentPriorityQueue) Pop() *Item {
	cpq.mtx.Lock()
	defer cpq.mtx.Unlock()
	if cpq.pq.Len() == 0 {
		return nil
	}
	return heap.Pop(&cpq.pq).(*Item)
}

var cpq = ConcurrentPriorityQueue{pq: make(PriorityQueue, 0)}

func handlePost(w http.ResponseWriter, r *http.Request, txHandler *TxHandler) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
			return
		}
		cpq.Push(&Item{value: string(body), priority: len(body)})
		fmt.Fprint(w, "Received POST request and added to queue")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {

	txHandler := NewTxHandler()
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		go handlePost(w, r, txHandler)
	})

	http.ListenAndServe(":8080", nil)
}

type TxHandler struct {
	txConfig client.TxConfig
}

func NewTxHandler() *TxHandler {
	reg := codectypes.NewInterfaceRegistry()

	authtypes.RegisterInterfaces(reg)
	banktypes.RegisterInterfaces(reg)
	govtypes.RegisterInterfaces(reg)

	cdc := codec.NewProtoCodec(reg)

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	return &TxHandler{txConfig: txConfig}

}

func (t *TxHandler) HandleTx(txBytes []byte) types.Tx {
	tx, err := t.txConfig.TxDecoder()(txBytes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tx)
	return tx
}
