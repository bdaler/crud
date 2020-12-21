package app

import (
	"encoding/json"
	"errors"
	"github.com/bdaler/crud/cmd/app/middleawre"
	"github.com/bdaler/crud/pkg/types"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const ADMIN = "ADMIN"

func (s *Server) handleManagerRegistration(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}
	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	if !s.managerSvc.IsAdmin(r.Context(), id) {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	var regItem struct {
		ID    int64    `json:"id"`
		Name  string   `json:"name"`
		Phone string   `json:"phone"`
		Roles []string `json:"roles"`
	}

	err = json.NewDecoder(r.Body).Decode(&regItem)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	item := &types.Manager{
		ID:    regItem.ID,
		Name:  regItem.Name,
		Phone: regItem.Phone,
	}

	for _, role := range regItem.Roles {
		if role == ADMIN {
			item.IsAdmin = true
			break
		}
	}

	token, err := s.managerSvc.Create(r.Context(), item)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	responseJSON(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerGetToken(w http.ResponseWriter, r *http.Request) {
	var manager *types.Manager
	err := json.NewDecoder(r.Body).Decode(&manager)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	token, err := s.managerSvc.Token(r.Context(), manager.Phone, manager.Password)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerChangeProducts(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}
	product := &types.Product{}
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	product, err = s.managerSvc.SaveProduct(r.Context(), product)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	responseJSON(w, product)
}

func (s *Server) handleManagerMakeSales(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	sale := &types.Sale{}
	sale.ManagerID = id
	err = json.NewDecoder(r.Body).Decode(&sale)

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	sale, err = s.managerSvc.MakeSale(r.Context(), sale)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, sale)
}

func (s *Server) handleManagerGetSales(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	total, err := s.managerSvc.GetSales(r.Context(), id)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, map[string]interface{}{"manager_id": id, "total": total})
}

func (s *Server) handleManagerGetProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.managerSvc.Products(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, items)
}

func (s *Server) handleManagerRemoveProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		errorWriter(w, http.StatusBadRequest, errors.New("id required"))
		return
	}
	productID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	err = s.managerSvc.RemoveProductByID(r.Context(), productID)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}
}

func (s *Server) handleManagerRemoveCustomerByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		errorWriter(w, http.StatusBadRequest, errors.New("Missing id"))
		return
	}

	customerID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	err = s.managerSvc.RemoveCustomerByID(r.Context(), customerID)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}
}

func (s *Server) handleManagerGetCustomers(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	items, err := s.managerSvc.Customers(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, items)
}

func (s *Server) handleManagerChangeCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errorWriter(w, http.StatusForbidden, err)
		return
	}

	customer := &types.Customer{}
	err = json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	customer, err = s.managerSvc.ChangeCustomer(r.Context(), customer)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	responseJSON(w, customer)
}
