package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	CreateAccount(Account) error
	UpdateAccount(Account) (*Account, error)
	DeleteAccount(int) error
	TransferToAccount(TransferRequest) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	// do not do this, working with locally
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createTableAccount()
}

func (s *PostgresStore) createTableAccount() error {
	query := `create table if not exists account(
      id serial primary key,
      first_name varchar(50),
      last_name varchar(50),
      number serial,
      balance serial,
      created_at timestamp
  )`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) CreateAccount(account Account) error {
	query := `insert into
  account (
      first_name,
      last_name,
      number,
      balance,
      created_at)
  values
  ($1,$2,$3,$4,$5)`

	_, err := s.db.Query(query,
		account.FirstName,
		account.LastName,
		account.Number,
		account.Balance,
		account.CreateAt)

	return err
}

func (s *PostgresStore) UpdateAccount(account Account) (*Account, error) {
  _, err := s.db.Query("update account set first_name = $2, last_name=$3 where id = $1", account.ID, account.FirstName, account.LastName)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query("select * from account where id = $1", account.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", account.ID)
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)
	return err
}

func (s *PostgresStore) TransferToAccount(detail TransferRequest) (*Account, error) {
  _, err := s.db.Query("update account set balance = $1 where id = $2", detail.Amount, detail.ToAccount)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query("select * from account where id = $1", detail.ToAccount)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", detail.ToAccount)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreateAt)

	return account, err

}
