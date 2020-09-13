package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/monodeepdas1215/go_test/core/logger"
	"github.com/monodeepdas1215/go_test/core/models"
	"github.com/monodeepdas1215/go_test/core/services"
	"net/http"
)

func AddTransactionController(ctx *gin.Context) {

	transactionId := ctx.Param("transaction_id")

	var reqBody models.AddTransactionRequestJson

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		logger.Logger.Errorln(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong happened while trying to decoding input json"})
		return
	}

	// some small validations
	if transactionId == "0" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id must be > 0"})
		return
	}

	if transactionId == fmt.Sprintf("%d", reqBody.ParentId) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id and parent_id are the same, keep parent_id = 0 for all root level transactions"})
		return
	}

	err := services.UpsertTransactionService(transactionId, &reqBody)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func GetTransactionTypesController(ctx *gin.Context) {

	queryType := ctx.Param("queryType")
	val := ctx.Param("val")

	data, err := services.GetRequestedDetails(queryType, val)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func GetTransactionSumController(ctx *gin.Context) {
}