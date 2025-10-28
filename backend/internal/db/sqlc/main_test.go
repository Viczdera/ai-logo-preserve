package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Viczdera/ai-logo-preserve/backend/internal/utils"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(t *testing.M) {

	config, err := utils.LoadConfig("../../../")
	if err != nil {
		log.Fatal("Failed to load config 💿", err)
	}

	dbSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName, config.Database.SSLMode)
	fmt.Println(dbSource)
	testDB, err = sql.Open(utils.DBDriver, dbSource)

	if err != nil {
		log.Fatal("Could not connect to DB", err)
	}

	//use connection to create new test queries object
	testQueries = New(testDB)

	//run and report back to test runner via the o.exit command
	os.Exit(t.Run())

}
