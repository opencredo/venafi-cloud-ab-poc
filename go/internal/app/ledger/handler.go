package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/ledgerserver"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/swaggerui"
)

type recordedTransaction struct {
	ledgerserver.Transaction
	ledgerserver.TransactionIdentifiers
}

type transactionsImpl struct {
	transactions []*recordedTransaction
}

func (t *transactionsImpl) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if len(t.transactions) == 0 {
		w.Write([]byte("[]"))
		return
	}

	buf, err := json.Marshal(t.transactions)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func (t *transactionsImpl) PostTransactions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sum := sha256.New()
	if len(t.transactions) > 0 {
		sum.Write([]byte(t.transactions[len(t.transactions)-1].Hash))
	}
	teer := io.TeeReader(r.Body, sum)
	d := json.NewDecoder(teer)
	txn := &recordedTransaction{}
	err := d.Decode(txn)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
		w.WriteHeader(http.StatusBadRequest)
	}
	txn.Hash = hex.EncodeToString(sum.Sum(nil))
	txn.Id = strconv.Itoa(len(t.transactions))

	t.transactions = append(t.transactions, txn)

	w.Header().Add("Location", fmt.Sprintf("/transactions/%s", txn.Id))
	w.WriteHeader(http.StatusCreated)
}

func (t *transactionsImpl) GetTransactionsTransactionId(w http.ResponseWriter, r *http.Request) {
	transactionId := r.Context().Value("transactionId").(string)
	idx, err := strconv.Atoi(transactionId)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(t.transactions) == 0 || idx < 0 || idx >= len(t.transactions) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buf, err := json.Marshal(t.transactions[idx])
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, err)))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func Handler() http.Handler {
	var api transactionsImpl

	r := chi.NewMux()
	r.Mount("/", ledgerserver.Handler(&api))
	r.Mount("/swaggerui", swaggerui.Handler())

	return r
}
