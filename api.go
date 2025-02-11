package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHandleFunc(s.handleAccountWithId))
	router.HandleFunc("/transfer", makeHandleFunc(s.handleTransfer))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return fmt.Errorf("Method not allowed %s", r.Method)
	}
}
func (s *APIServer) handleAccountWithId(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccountByID(w, r)
	case "PUT":
		return s.handleUpdateAccount(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
	default:
		return fmt.Errorf("Method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "PUT":
		return s.handleTransferToAccount(w, r)
	default:
		return fmt.Errorf("Method not allowed %s", r.Method)
	}
}

// GET
func (s *APIServer) handleGetAccount(w http.ResponseWriter, _ *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

// CREATE
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccount := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccount); err != nil {
		return err
	}

	account := NewAccount(createAccount.FirstName, createAccount.LastName)
	if err := s.store.CreateAccount(*account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, createAccount)
}

// UPDATE
func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	reqAccount := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(reqAccount); err != nil {
		return err
	}
	account := Account{ID: id, FirstName: reqAccount.FirstName, LastName: reqAccount.LastName}

	updatedAccount, err := s.store.UpdateAccount(account)
	if err != nil {
		return nil
	}

	return WriteJSON(w, http.StatusOK, updatedAccount)
}

// DELETE
func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

// Transfer
func (s *APIServer) handleTransferToAccount(w http.ResponseWriter, r *http.Request) error {
	transfer := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transfer); err != nil {
		return err
	}

	account, err := s.store.TransferToAccount(*transfer)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

type apiFunc func(http.ResponseWriter, *http.Request) error
type ApiError struct {
	Error string
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("Invalid id given %s", idStr)
	}
	return id, nil
}
