package services

import (
	"errors"
	data_store "github.com/monodeepdas1215/go_test/core/data-store"
	"github.com/monodeepdas1215/go_test/core/logger"
	"github.com/monodeepdas1215/go_test/core/models"
	"strconv"
)

func UpsertTransactionService(trId string, reqJson *models.AddTransactionRequestJson) error {

	transactionId, err := strconv.ParseInt(trId, 10, 64)
	if err != nil {
		logger.Logger.Errorln("error occurred while converting transaction_id from string to int64: ", err)
		return err
	}

	details := make(map[string]interface{})
	details["id"] = transactionId
	details["amount"] = reqJson.Amount
	details["type"] = reqJson.Type

	if reqJson.ParentId > 0 {
		details["parent_id"] = reqJson.ParentId
	}

	return data_store.AppDb.UpsertTransactionDetails(details)
}

func GetTransactionDetailsService(queryType string, val interface{}) (interface{}, error) {

	filter := make(map[string]interface{})

	if queryType == "types" {
		filter["type"] = val.(string)
		return data_store.AppDb.GetTransactionDetails(filter)
	}

	id, err := strconv.ParseInt(val.(string), 10, 64)
	if err != nil {
		logger.Logger.Errorln("error occurred while converting transaction_id from string to int64: ", err)
	}

	if queryType == "transaction" {
		filter["id"] = id
		return data_store.AppDb.GetTransactionDetails(filter)
	}

	if queryType == "sum" {
		filter["id"] = id
		res, err := data_store.AppDb.GetTransactionDetails(filter)
		if err != nil {
			return nil, err
		}

		if len(res) == 0 {
			x := make(map[string]int)
			x["sum"] = 0
			return x, nil
		}

		tmp := res[0]
		data := make(map[string]interface{})
		data["sum"] = tmp.TotalAmount
		return data, nil
	}

	return nil, errors.New("query type not supported")
}