package ledger

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"go.uber.org/zap"
)

type dbStore struct {
	pgConfig *pgx.ConnConfig
	conn     *pgx.Conn
	context  context.Context
}

func InitDB(logger *zap.Logger, pgConnection string) error {
	s := &dbStore{}
	var err error

	s.context = context.Background()

	s.pgConfig, err = pgx.ParseConfig(pgConnection)
	if err != nil {
		return err
	}
	s.pgConfig.Logger = zapadapter.NewLogger(logger)

	s.conn, err = pgx.ConnectConfig(s.context, s.pgConfig)
	if err != nil {
		return err
	}

	err = createTable(s)
	if err != nil {
		return err
	}

	RegisterStore(s)

	return nil
}

func createTable(s *dbStore) error {
	_, err := s.conn.Exec(s.context, `
	  CREATE TABLE IF NOT EXISTS transactions (
		  id INT DEFAULT unique_rowid(),
		  amount FLOAT NOT NULL,
		  description STRING NOT NULL,
		  from_acct INT NOT NULL,
		  to_acct INT NOT NULL,
		  hash STRING NOT NULL,
		  type STRING NOT NULL
	  )`)

	return err
}

func (s *dbStore) GetAll() ([]*recordedTransaction, error) {
	rows, err := s.conn.Query(s.context, "SELECT id, amount, description, from_acct, to_acct, hash, type FROM transactions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txns := make([]*recordedTransaction, 0)
	for rows.Next() {
		txn := recordedTransaction{}
		var id int
		err := rows.Scan(
			&id,
			&txn.Amount,
			&txn.Description,
			&txn.FromAcct,
			&txn.ToAcct,
			&txn.Hash,
			&txn.Type)
		if err != nil {
			return nil, err
		}

		txn.Id = strconv.Itoa(id)

		txns = append(txns, &txn)
	}

	return txns, nil
}

func (s *dbStore) Get(id string) (*recordedTransaction, error) {
	row := s.conn.QueryRow(s.context, "SELECT id, amount, description, from_acct, to_acct, hash, type FROM transactions WHERE id = $1", id)

	txn := recordedTransaction{}
	var retId int
	err := row.Scan(
		&retId,
		&txn.Amount,
		&txn.Description,
		&txn.FromAcct,
		&txn.ToAcct,
		&txn.Hash,
		&txn.Type)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	txn.Id = strconv.Itoa(retId)

	return &txn, nil
}

func (s *dbStore) Last() (*recordedTransaction, error) {
	row := s.conn.QueryRow(s.context, "SELECT id, amount, description, from_acct, to_acct, hash, type FROM transactions ORDER BY id DESC LIMIT 1")

	txn := recordedTransaction{}
	var id int
	err := row.Scan(
		&id,
		&txn.Amount,
		&txn.Description,
		&txn.FromAcct,
		&txn.ToAcct,
		&txn.Hash,
		&txn.Type)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	txn.Id = strconv.Itoa(id)

	return &txn, nil
}

func (s *dbStore) Add(txn *recordedTransaction) (string, error) {
	row := s.conn.QueryRow(s.context, "INSERT INTO transactions (amount, description, from_acct, to_acct, hash, type) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		txn.Amount,
		txn.Description,
		txn.FromAcct,
		txn.ToAcct,
		txn.Hash,
		txn.Type)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(id), nil
}
