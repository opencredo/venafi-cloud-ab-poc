package ledger

import "strconv"

type memoryStore []*recordedTransaction

func (m *memoryStore) GetAll() ([]*recordedTransaction, error) {
	return *m, nil
}

func (m *memoryStore) Get(id string) (*recordedTransaction, error) {
	idx, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	if len(*m) == 0 || idx < 0 || idx >= len(*m) {
		// Not found
		return nil, nil
	}

	return (*m)[idx], nil
}

func (m *memoryStore) Last() (*recordedTransaction, error) {
	if len(*m) == 0 {
		return nil, nil
	}
	return (*m)[len(*m)-1], nil
}

func (m *memoryStore) Add(txn *recordedTransaction) (string, error) {
	txn.Id = strconv.Itoa(len(*m))

	*m = append(*m, txn)

	return txn.Id, nil
}
