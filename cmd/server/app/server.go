package app

import (
	"encoding/json"
	"github.com/bdaler/crud/pkg/customers"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	mux         *http.ServeMux
	customerSvc *customers.Service
}

//NewServer construct
func NewServer(mux *http.ServeMux, customerSvc *customers.Service) *Server {
	return &Server{mux: mux, customerSvc: customerSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) Init() {
	log.Println("Init method")
	s.mux.HandleFunc("/banners.getById", s.handleGetCustomerById)
}

func (s *Server) handleGetCustomerById(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.customerSvc.ByID(request.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	jsonResponse(writer, data)
}

func jsonResponse(writer http.ResponseWriter, data []byte) {
	writer.Header().Set("Content-Type", "application/json")
	_, err := writer.Write(data)
	if err != nil {
		log.Println("Error write response: ", err)
	}
}
