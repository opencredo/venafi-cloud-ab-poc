package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/ledgerserver"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/swaggerui"
)

var pgConnection string
var store transactionStore = &memoryStore{}

func RegisterStore(s transactionStore) {
	store = s
}

type transactionStore interface {
	GetAll() ([]*recordedTransaction, error)
	Get(string) (*recordedTransaction, error)
	Last() (*recordedTransaction, error)
	Add(t *recordedTransaction) (string, error)
}

type recordedTransaction struct {
	ledgerserver.Transaction
	ledgerserver.TransactionIdentifiers
}

type transactionsImpl struct {
	transactions transactionStore
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
}

func (t *transactionsImpl) GetTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := t.transactions.GetAll()
	if err != nil {
		writeError(w, err)
		return
	}
	buf, err := json.Marshal(transactions)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Write(buf)
}

func (t *transactionsImpl) PostTransactions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sum := sha256.New()
	last, err := t.transactions.Last()
	if err != nil {
		writeError(w, err)
		return
	}

	if last != nil {
		sum.Write([]byte(last.Hash))
	}

	teer := io.TeeReader(r.Body, sum)
	d := json.NewDecoder(teer)
	txn := &recordedTransaction{}
	err = d.Decode(txn)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
		w.WriteHeader(http.StatusBadRequest)
	}
	txn.Hash = hex.EncodeToString(sum.Sum(nil))

	id, err := t.transactions.Add(txn)
	if err != nil {
		writeError(w, err)
		return
	}

	requestURI := r.Header.Get("X-Original-Uri")
	if requestURI == "" {
		requestURI = r.RequestURI
	}
	w.Header().Add("Location", fmt.Sprintf("%s/%s", requestURI, id))
	w.WriteHeader(http.StatusCreated)
}

func (t *transactionsImpl) GetTransactionsTransactionId(w http.ResponseWriter, r *http.Request) {
	transactionId := r.Context().Value("transactionId").(string)
	txn, err := t.transactions.Get(transactionId)
	if err != nil {
		writeError(w, err)
		return
	}

	if txn == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buf, err := json.Marshal(txn)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Write(buf)
}

func Handler() http.Handler {
	api := transactionsImpl{transactions: store}

	r := chi.NewMux()
	r.Mount("/", ledgerserver.Handler(&api))
	r.Mount("/swaggerui", swaggerui.Handler())

	return r
}
