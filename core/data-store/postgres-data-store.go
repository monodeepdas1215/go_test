package data_store

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/monodeepdas1215/go_test/core/config"
	"github.com/monodeepdas1215/go_test/core/logger"
	"github.com/monodeepdas1215/go_test/core/models"
	"sync"
)

var once sync.Once
var initPgDb *PostgresConnection

type PostgresConnection struct {
	db *gorm.DB
}

// singleton implementation of getting a database connection
func GetNewPgDatabaseConnection() *PostgresConnection {
	once.Do(func() {

		logger.Logger.Infoln("connecting to postgres database...")

		conn := &PostgresConnection{db: nil}

		if err := conn.Connect(); err != nil {
			panic("application database is nil")
		} else {
			initPgDb = conn
			logger.Logger.Infoln("connection to database successful")
		}
	})
	return initPgDb
}

func (pc *PostgresConnection) Connect() error {

	var err error

	DbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", config.DbHost, config.DbPort, config.DbUser, config.DbName, config.DbPass)

	pc.db, err = gorm.Open(config.DbDriver, DbURI)
	if err != nil {
		logger.Logger.Errorln("error occurred while connecting to postgres database: ", err)
		return err
	}

	pc.db.LogMode(config.DbDebugStatus)

	if config.DbRecreate {
		pc.db.DropTableIfExists(&models.Transactions{})
	}
	pc.db.AutoMigrate(&models.Transactions{})

	return nil
}

func (pc *PostgresConnection) Disconnect() {
	if err := pc.db.Close(); err != nil {
		logger.Logger.Errorln("error occurred while closing a connection: ", err)
	} else {
		logger.Logger.Infoln("database connection disconnected successfully...")
	}
}

func (pc *PostgresConnection) UpsertTransactionDetails(details map[string]interface{}) error {

	logger.Logger.Debugf("details: %+v\n", details)

	// starting a transaction
	tx := pc.db.Begin()

	var parent *models.Transactions = nil
	var present *models.Transactions = nil

	// finding the record
	var tmp models.Transactions
	if err := tx.Where(map[string]interface{}{"id": details["id"]}).First(&tmp).Error; err != nil {

		if !gorm.IsRecordNotFoundError(err) {

			logger.Logger.Errorln("some error occurred while fetching the record: ", err)
			tx.RollbackUnlessCommitted()
			return err
		}

	} else {
		// if record found then simply point the present to it
		present = &tmp
	}

	/*
	case 1: for root level transactions no parent_id is specified
	if present record not found then simply create the present record
	else
	update the present record with the details and also update the difference amount in total sum
	*/
	if _, ok := details["parent_id"]; !ok {

		if present == nil {
			tmp := models.Transactions{
				Id:          details["id"].(int64),
				ParentId:    0,
				Amount:      details["amount"].(float64),
				Type:        details["type"].(string),
				TotalAmount: details["amount"].(float64),
			}

			tx.Create(&tmp)

		} else {

			present.TotalAmount = present.TotalAmount - present.Amount + details["amount"].(float64)
			present.Amount = details["amount"].(float64)
			present.Type = details["type"].(string)

			tx.Save(present)
		}

		tx.Commit()
		return nil

	} else {

		/*
		case 2:

		raise error if the parent_id record is not found

		for non root level transactions
		if present record is not found then
			create the present record
		else
			update the present record and fix the total amount difference

		recursively update the total amount in the parent nodes until root is reached
		*/

		var tmp models.Transactions

		if err := tx.Where(map[string]interface{}{"id": details["parent_id"]}).First(&tmp).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				logger.Logger.Errorln("parent record does not exists")
				tx.RollbackUnlessCommitted()
				return errors.New("parent_id is invalid")
			} else {
				logger.Logger.Errorln("error occurred while fetching the parent record: ", err)
				tx.RollbackUnlessCommitted()
				return err
			}
		}

		parent = &tmp

		var differenceAmt float64
		var lastParentId int64 = -1

		// handling the present record
		if present == nil {

			logger.Logger.Debugln("parent with no present")

			tmp := models.Transactions{
				Id:          details["id"].(int64),
				ParentId:    details["parent_id"].(int64),
				Amount:      details["amount"].(float64),
				Type:        details["type"].(string),
				TotalAmount: details["amount"].(float64),
			}

			differenceAmt = tmp.Amount
			tx.Create(&tmp)

			// recursively change the parent total amount
			for {

				parent.TotalAmount += differenceAmt
				tx.Save(parent)

				// exit when the parent record is modified
				if parent.ParentId == 0 {
					break
				}

				var tmp models.Transactions
				if err := tx.Where(map[string]interface{}{"id": parent.ParentId}).First(&tmp).Error; err != nil {

					if gorm.IsRecordNotFoundError(err) {
						break
					} else {
						logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
						tx.RollbackUnlessCommitted()
						return err
					}

				} else {
					parent = &tmp
				}
			}

			tx.Commit()
			return nil

		} else {

			present.Type = details["type"].(string)

			// the parent record is same
			if present.ParentId == details["parent_id"].(int64) {

				// if the amount has changed
				if present.Amount != details["amount"].(float64) {

					differenceAmt = details["amount"].(float64) - present.Amount
					present.TotalAmount = present.TotalAmount + differenceAmt
					present.Amount = details["amount"].(float64)

				}

				// save the present record
				tx.Save(present)

				// recursively change the parent total amount
				for {

					parent.TotalAmount += differenceAmt
					tx.Save(parent)

					// exit when the parent record is modified
					if parent.ParentId == 0 {
						break
					}

					var tmp models.Transactions
					if err := tx.Where(map[string]interface{}{"id": parent.ParentId}).First(&tmp).Error; err != nil {

						if gorm.IsRecordNotFoundError(err) {
							break
						} else {
							logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
							tx.RollbackUnlessCommitted()
							return err
						}

					} else {
						parent = &tmp
					}
				}

				tx.Commit()
				return nil

			} else {
				// the parent record has changed

				lastParentId = present.ParentId
				present.ParentId = details["parent_id"].(int64)

				if present.Amount != details["amount"].(float64) {
					// amount has changed

					differenceAmt = present.TotalAmount
					present.TotalAmount = present.TotalAmount + details["amount"].(float64) - present.Amount
					present.Amount = details["amount"].(float64)

					addAmount := present.TotalAmount

					tx.Save(present)

					// adding the difference amount to the new parent_id recursively until root
					for {

						parent.TotalAmount += addAmount
						tx.Save(parent)

						// exit when the parent record is modified
						if parent.ParentId == 0 {
							break
						}

						var tmp models.Transactions
						if err := tx.Where(map[string]interface{}{"id": parent.ParentId}).First(&tmp).Error; err != nil {

							if gorm.IsRecordNotFoundError(err) {
								break
							} else {
								logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
								tx.RollbackUnlessCommitted()
								return err
							}

						} else {
							parent = &tmp
						}
					}

					// subtracting the difference amount to the new parent_id recursively until root
					for {

						var tmp models.Transactions
						if err := tx.Where(map[string]interface{}{"id": lastParentId}).First(&tmp).Error; err != nil {

							if gorm.IsRecordNotFoundError(err) {
								break
							} else {
								logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
								tx.RollbackUnlessCommitted()
								return err
							}

						} else {
							parent = &tmp
							lastParentId = parent.ParentId
						}

						parent.TotalAmount -= differenceAmt
						tx.Save(parent)

						// exit when the parent record is modified
						if lastParentId == 0 {
							break
						}
					}

					tx.Commit()
					return nil
				} else {
					// amount has not changed

					tx.Save(present)

					differenceAmt = present.TotalAmount

					// adding the difference amount to the new parent_id recursively until root
					for {

						parent.TotalAmount += differenceAmt
						tx.Save(parent)

						// exit when the parent record is modified
						if parent.ParentId == 0 {
							break
						}

						var tmp models.Transactions
						if err := tx.Where(map[string]interface{}{"id": parent.ParentId}).First(&tmp).Error; err != nil {

							if gorm.IsRecordNotFoundError(err) {
								break
							} else {
								logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
								tx.RollbackUnlessCommitted()
								return err
							}

						} else {
							parent = &tmp
						}
					}

					// subtracting the difference amount to the new parent_id recursively until root
					for {

						var tmp models.Transactions
						if err := tx.Where(map[string]interface{}{"id": lastParentId}).First(&tmp).Error; err != nil {

							if gorm.IsRecordNotFoundError(err) {
								break
							} else {
								logger.Logger.Errorln("some error occurred while finding parent in loop: ", err)
								tx.RollbackUnlessCommitted()
								return err
							}

						} else {
							parent = &tmp
							lastParentId = parent.ParentId
						}

						parent.TotalAmount -= differenceAmt
						tx.Save(parent)

						// exit when the parent record is modified
						if lastParentId == 0 {
							break
						}
					}

					tx.Commit()
					return nil
				}
			}
		}
	}
}

func (pc *PostgresConnection) GetTransactionDetails(query map[string]interface{}) ([]models.Transactions, error) {

	var transactions []models.Transactions

	if err := pc.db.Model(&models.Transactions{}).Where(query).Scan(&transactions).Error; err != nil {
		logger.Logger.Errorln("error occurred while fetching records: ", err)
		return transactions, err
	}

	return transactions, nil
}