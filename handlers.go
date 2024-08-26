package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RecordHandler struct {
	db     *gorm.DB
	rdb    *redis.Client
	logger *log.Logger
}

const RecordsKey = "records"

type ResponseMsg map[string]string

func NewRecordHandler(db *gorm.DB, rdb *redis.Client, logger *log.Logger) *RecordHandler {
	return &RecordHandler{
		db:     db,
		rdb:    rdb,
		logger: logger,
	}
}

func (rh *RecordHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var records []Record
	err := rh.db.Find(&records).Error
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	ToJSON(w, records, http.StatusOK)
}

func (rh *RecordHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRecordRequest
	if err := FromJSON(r.Body, &req); err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	record := Record{
		Data:      req.Data,
		Location:  req.Location,
		OrderDate: req.OrderDate,
		Forward:   req.Forward,
	}

	// database operation
	err := rh.db.Create(&record).Error
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}
	// add count to redis
	err = rh.rdb.Incr(r.Context(), RecordsKey).Err()
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	// return data
	ToJSON(w, record, http.StatusCreated)
}

func (rh *RecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	recordId, err := uuid.Parse(id)
	if err != nil {
		ToJSON(w, ResponseMsg{"error": "id invalid: " + err.Error()}, http.StatusBadRequest)
		return
	}

	var req createRecordRequest
	if err := FromJSON(r.Body, &req); err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	record := Record{
		ID:        recordId,
		Data:      req.Data,
		Location:  req.Location,
		OrderDate: req.OrderDate,
		Forward:   req.Forward,
	}

	// database operation
	err = rh.db.Updates(&record).Error
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	ToJSON(w, record, http.StatusOK)
}

func (rh *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	recordId, err := uuid.Parse(id)
	if err != nil {
		ToJSON(w, ResponseMsg{"error": "id invalid: " + err.Error()}, http.StatusBadRequest)
		return
	}

	err = rh.db.Delete(&Record{ID: recordId}).Error
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	err = rh.rdb.Decr(r.Context(), RecordsKey).Err()
	if err != nil {
		ToJSON(w, ResponseMsg{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	ToJSON(w, ResponseMsg{"message": "record deleted"}, http.StatusOK)
}
