package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type JSONData map[string]interface{}

type testEngine struct {
	db  *gorm.DB
	rdb *redis.Client
}

func TestCreateRecord(t *testing.T) {
	engine, flush := setup()
	defer flush()

	logger := log.Default()
	handler := NewRecordHandler(engine.db, engine.rdb, logger)

	recorder := httptest.NewRecorder()
	reqData := JSONData{
		"data":       "ABC - new",
		"order_date": "2005-04-01T15:04:05Z",
		"forward":    true,
		"location":   "Port 1",
	}
	content, err := json.Marshal(reqData)
	if err != nil {
		t.Fatal(err)
	}
	data := bytes.NewBuffer(content)
	request, err := http.NewRequest("POST", "/records", data)
	if err != nil {
		t.Fatal(err)
	}

	handler.Create(recorder, request)
	res := recorder.Result()
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	var count int64
	err = handler.db.Model(&Record{}).Count(&count).Error
	assert.Nil(t, err)
	assert.EqualValues(t, count, 1)
}

func TestGetAllRecords(t *testing.T) {
	engine, flush := setup()
	defer flush()

	count := 5
	records, err := populateRecord(engine, count)
	if err != nil {
		t.Fatal(err)
	}

	logger := log.Default()
	handler := NewRecordHandler(engine.db, engine.rdb, logger)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/records", nil)
	handler.GetAll(recorder, request)

	result := recorder.Result()
	assert.Equal(t, result.StatusCode, http.StatusOK)

	var response []JSONData
	if err := json.NewDecoder(result.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if err := result.Body.Close(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(response), count)
	assert.EqualValues(t, records[0].Data, response[0]["data"])
	assert.EqualValues(t, records[count-1].Data, response[count-1]["data"])
}

func TestUpdateRecord(t *testing.T) {
	engine, flush := setup()
	defer flush()

	count := 1
	records, err := populateRecord(engine, count)
	if err != nil {
		t.Fatal(err)
	}
	record := records[0]

	logger := log.Default()
	handler := NewRecordHandler(engine.db, engine.rdb, logger)

	requestData := JSONData{
		"data": "record - updated",
	}
	content, err := json.Marshal(requestData)
	if err != nil {
		t.Fatal(err)
	}
	data := bytes.NewBuffer(content)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/records", data)
	request.SetPathValue("id", record.ID.String())
	handler.Update(recorder, request)

	result := recorder.Result()

	assert.Equal(t, result.StatusCode, http.StatusOK)

	var retrieved Record
	err = engine.db.First(&retrieved, &Record{ID: record.ID}).Error
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, retrieved.Data, requestData["data"])
}

func TestDeleteRecord(t *testing.T) {
	engine, flush := setup()
	defer flush()

	records, err := populateRecord(engine, 1)
	if err != nil {
		t.Fatal(err)
	}
	record := records[0]

	handler := NewRecordHandler(engine.db, engine.rdb, log.Default())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/records", nil)
	request.SetPathValue("id", record.ID.String())
	handler.Delete(recorder, request)

	result := recorder.Result()
	assert.Equal(t, result.StatusCode, http.StatusOK)

	var retrieved []Record
	err = engine.db.Find(&retrieved).Error
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(retrieved), 0)
}

func populateRecord(engine *testEngine, count int) ([]Record, error) {
	var records []Record
	for range count {
		id := uuid.New()
		records = append(records, Record{
			ID:        id,
			Data:      faker.Word(),
			OrderDate: time.Now(),
			Forward:   true,
			Location:  faker.GetCountryInfo().Name,
		})
	}
	return records, engine.db.CreateInBatches(records, 30).Error
}

func setup() (engine *testEngine, callback func()) {
	if err := godotenv.Load("test.env"); err != nil {
		log.Fatal(err)
	}

	engine = &testEngine{
		db:  setupDB(),
		rdb: setupRedis(),
	}

	// flush databases
	return engine, func() {
		err := flushDB(engine.db)
		if err != nil {
			log.Fatal(err)
		}

		err = flushRDB(engine.rdb)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setupDB() *gorm.DB {
	db := SQL(os.Getenv("PG_DSN"))
	return db
}

func setupRedis() *redis.Client {
	redis_db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("redis db parsing error: %v", err)
	}
	rdb := Redis(redis_db)
	return rdb
}

func flushDB(db *gorm.DB) error {
	return db.Migrator().DropTable(&Record{})
}

func flushRDB(rdb *redis.Client) error {
	return rdb.FlushDB(context.Background()).Err()
}
