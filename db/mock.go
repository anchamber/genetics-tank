package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"

	apiModel "github.com/anchamber/genetics-api/model"
	"github.com/anchamber/genetics-tank/db/model"
)

type TankDBMock struct {
	DB *sqlx.DB
}

var MockDataTanks = []*model.Tank{
}

func NewMockDB(initialData []*model.Tank) TankDBMock {
	if initialData == nil {
		initialData = MockDataTanks
	}
	mock := TankDBMock{
		DB: initDB(),
	}
	mock.DB.SetMaxOpenConns(1)
	for _, tank := range initialData {
		err := mock.Insert(tank)
		if err != nil {
			return TankDBMock{}
		}
	}

	return mock
}

func (o *Options) createPaginationClause() string {
	if o.Pageination == nil {
		return ""
	}
	var limit int64 = int64(o.Pageination.Limit)
	if limit <= 0 {
		limit = -1
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, o.Pageination.Offset)
}

func getOperatorAsString(operator apiModel.Operator) string {
	switch operator {
	case apiModel.EQ:
		return "="
	case apiModel.GREATER:
		return ">"
	case apiModel.GREATER_EQ:
		return ">="
	case apiModel.SMALLER:
		return "<"
	case apiModel.SMALLER_EQ:
		return "<="
	case apiModel.CONTAINS:
		return "LIKE"
	default:
		return "="
	}
}

func (o *Options) createFilterClause() string {
	if len(o.Filters) == 0 {
		return ""
	}
	whereClause := "WHERE "

	for index, filter := range o.Filters {
		if index > 0 {
			whereClause += " AND "
		}

		if filter.Operator == apiModel.CONTAINS {
			whereClause += fmt.Sprintf("instr(%s, :%s) > 0", filter.Key, filter.Key)
		} else {
			whereClause += fmt.Sprintf("%s %v :%s", filter.Key, getOperatorAsString(filter.Operator), filter.Key)
		}
	}
	// fmt.Println(whereClause)
	return whereClause
}

func (o *Options) createFilterMap() map[string]interface{} {
	values := make(map[string]interface{})
	for _, filter := range o.Filters {
		values[filter.Key] = filter.Value
	}
	return values
}

func (tankDB TankDBMock) Select(options Options) ([]*model.Tank, error) {
	selectStatement := fmt.Sprintf("SELECT id, tank, number, active, size, fish_count FROM tanks %s %s;", options.createFilterClause(), options.createPaginationClause())
	// fmt.Println(selectStatement)
	filterValues := options.createFilterMap()
	rows, err := tankDB.DB.NamedQuery(selectStatement, filterValues)
	if err != nil {
		fmt.Printf("%v\n", err)
		log.Fatalf(`failed to select all`)
		return nil, err
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf(fmt.Sprintf("failed closing rows %v\n", err))
		}
	}(rows)
	var data []*model.Tank
	for rows.Next() {
		var entry model.Tank
		err = rows.Scan(&entry.ID, &entry.System, &entry.Number, &entry.Active, &entry.Size, &entry.FishCount)
		if err != nil {
			return nil, err
		}
		data = append(data, &entry)
	}

	return data, nil
}

func (tankDB TankDBMock) SelectByNumber(number uint32) (*model.Tank, error) {
	//goland:noinspection ALL
	selectStatement := `
		SELECT tank, number, active, size, fish_count
		FROM tanks
		WHERE number = $1;
	`
	rows, err := tankDB.DB.Query(selectStatement, number)
	if err != nil {
		log.Fatalf(`failed to select all`)
		return nil, err
	}

	var entry model.Tank
	if !rows.Next() {
		return nil, nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf(fmt.Sprintf("failed closing rows %v\n", err))
		}
	}(rows)
	err = rows.Scan(&entry.ID, &entry.System, &entry.Number, &entry.Active, &entry.Size, &entry.FishCount)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (tankDB TankDBMock) Insert(tank *model.Tank) error {
	var errorString = ""
	//goland:noinspection ALL
	insertStatement := `
		INSERT INTO tanks (tank, number, active, size, fish_count)
			VALUES (?, ?, ?, ?, ?, ?);
	`
	tx, err := tankDB.DB.Begin()
	if err != nil {
		fmt.Printf("failed to begin transaction\n")
		return err
	}

	statement, err := tx.Prepare(insertStatement)
	if err != nil {
		fmt.Printf("failed to prepare statement\n")
		return err
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			fmt.Printf(fmt.Sprintf("failed closing statement %v\n", err))
		}
	}(statement)

	_, err = statement.Exec(tank.ID, tank.System, tank.Number, tank.Active, tank.Size, tank.FishCount)
	if err != nil {
		fmt.Printf("failed to execute statement\n")
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			switch sqliteErr.Code {
			case sqlite3.ErrConstraint:
				errorString = string(TankAlreadyExists)
			default:
				fmt.Printf("%v\n", sqliteErr)
				errorString = string(Unknown)
			}
		} else {
			fmt.Printf("%v\n", err.Error())
			errorString = string(Unknown)
		}
		err := tx.Rollback()
		if err != nil {
			return err
		}
	} else {
		// numberCreated, _ := result.RowsAffected()
		// fmt.Printf("created %d entries\n", numberCreated)
		err := tx.Commit()
		if err != nil {
			return err
		}
	}
	if errorString == "" {
		return nil
	}
	return errors.New(errorString)
}

func (tankDB TankDBMock) Update(tank *model.Tank) error {
	//goland:noinspection ALL
	insertStatement := `
		UPDATE tanks 
			SET tank = $1, number = $2, active = $3, size = $4, fish_count = $5
			WHERE name = $1;
	`
	tx, err := tankDB.DB.Begin()
	if err != nil {
		fmt.Printf("failed to begin transaction\n")
		return err
	}

	statement, err := tx.Prepare(insertStatement)
	if err != nil {
		fmt.Printf("failed to prepare statement\n")
		return err
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			fmt.Printf(fmt.Sprintf("failed closing statement %v\n", err))
		}
	}(statement)

	_, err = statement.Exec(tank.ID, tank.System, tank.Number, tank.Active, tank.Size, tank.FishCount)
	if err != nil {
		fmt.Printf("failed to execute statement\n")
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (tankDB TankDBMock) Delete(number uint32) error {
	//goland:noinspection ALL
	statementString := `
		DELETE FROM tanks WHERE number = ?;
	`
	statement, err := tankDB.DB.Prepare(statementString)
	if err != nil {
		fmt.Printf("failed to prepare statement\n")
		return err
	}
	defer func(statement *sql.Stmt) {
		err := statement.Close()
		if err != nil {
			fmt.Printf(fmt.Sprintf("failed closing statement %v\n", err))
		}
	}(statement)

	_, err = statement.Exec(number)
	if err != nil {
		fmt.Printf("failed to execute statement\n")
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			switch sqliteErr.Code {
			case sqlite3.ErrConstraint:
			default:
			}
		} else {
			fmt.Printf("%v\n", err.Error())
			return errors.New(string(Unknown))
		}
	}
	return nil
}

func initDB() *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	err = CreateTables(db)
	if err != nil {
		return nil
	}
	if err != nil {
		return nil
	}

	return db
}

func CreateTables(db *sqlx.DB) error {
	//goland:noinspection ALL
	tankTable := `
		CREATE TABLE IF NOT EXISTS tanks(
			id					INTEGER	PRIMARY KEY AUTOINCREMENT,
			number				INT UNIQUE,
			tank				string,
			active				bit ,
			size				INT,
			fish_count 			INT
		);
	`

	_, err := db.Exec(tankTable)
	if err != nil {
		log.Fatalf("failed to create table\n")
		return err
	}
	return nil
}
