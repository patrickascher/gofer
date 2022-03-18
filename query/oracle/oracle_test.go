// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package oracle

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/guregu/null"
	"github.com/patrickascher/gofer/logger/mocks"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/query/types"
	"github.com/stretchr/testify/assert"
)

type Config struct {
	DB query.Config
}

// testConfig helper returns the DB config.
func testConfig() Config {
	return Config{DB: query.Config{Username: "root", Password: "root", Host: "127.0.0.1", Port: 3306}}
}

// createDatabase helper for the tests.
func createDatabase(asserts *assert.Assertions) {

	m := mysql{}
	m.Base.Config = testConfig().DB
	err := m.Open()
	asserts.NoError(err)

	_, err = m.DB().Exec("DROP DATABASE IF EXISTS `tests`")
	asserts.NoError(err)

	_, err = m.DB().Exec("CREATE DATABASE `tests` DEFAULT CHARACTER SET = `utf8`")
	asserts.NoError(err)

}

// createTable helper for the tests.
func createTable(b query.Builder, asserts *assert.Assertions) {

	_, err := b.Query().DB().Exec("DROP TABLE IF EXISTS `query`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("CREATE TABLE `query` (\n`id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n`int` int(11) DEFAULT NULL,\n`varchar` varchar(250) DEFAULT NULL,\n`tinyint` tinyint(4) DEFAULT '4',\n`smallint` smallint(6) DEFAULT NULL,\n`mediumint` mediumint(9) DEFAULT NULL,\n`bigint` bigint(20) DEFAULT NULL,\n`float` float DEFAULT NULL,\n`double` double DEFAULT NULL,\n`char` char(1) DEFAULT NULL,\n`tinytext` tinytext,\n`text` text,\n`mediumtext` mediumtext,\n`longtext` longtext,\n`enum` enum('JOHN','DOE') DEFAULT NULL,\n`set` set('FOO','BAR') DEFAULT NULL,\n`date` date DEFAULT NULL,\n`datetime` datetime DEFAULT NULL,\n`timestamp` timestamp NULL DEFAULT NULL,\n`bool` tinyint(1) DEFAULT NULL,`utinyint` tinyint(3) unsigned DEFAULT NULL,\n  `usmallint` smallint(5) unsigned DEFAULT NULL,\n  `umediumint` mediumint(8) unsigned DEFAULT NULL,\n  `ubigint` bigint(20) unsigned DEFAULT NULL,`time` time DEFAULT NULL,`geometry` geometry DEFAULT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("CREATE TABLE `query_fk` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, PRIMARY KEY (`id`), CONSTRAINT `query_fk_ibfk_1` FOREIGN KEY (`id`) REFERENCES `query` (`id`)) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8")
	asserts.NoError(err)
}

// truncateTestTable is a helper to delete the test table.
func truncateTestTable(b query.Builder, asserts *assert.Assertions) {
	_, err := b.Query().DB().Exec("Delete FROM query")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("ALTER TABLE query AUTO_INCREMENT=0;")
	asserts.NoError(err)
}

type InsertTest struct {
	Int     null.Int
	Varchar null.String
}

// checkInsertResult helper to check if the result was added correctly
func checkInsertResult(b query.Builder, asserts *assert.Assertions) []InsertTest {

	rows, err := b.Query().DB().Query("SELECT `int`,`varchar` FROM `query`")
	asserts.NoError(err)

	var rv []InsertTest
	for rows.Next() {
		val := InsertTest{}
		err = rows.Scan(&val.Int, &val.Varchar)
		asserts.NoError(err)
		rv = append(rv, val)
	}

	return rv
}

// createTestEntries helper to add some data.
func createTestEntries(b query.Builder, asserts *assert.Assertions) {
	truncateTestTable(b, asserts)
	_, err := b.Query().DB().Exec("INSERT INTO query (`int`,`varchar`) VALUES (?,?),(?,?),(?,?)", 1, "a", 2, "b", 3, "c")
	asserts.NoError(err)
}

// TestBase_Without_DB checks if an error returns if on the base.Open function when no sql.DB is set.
func TestBase_Without_DB(t *testing.T) {
	asserts := assert.New(t)
	mysql := mysql{}
	err := mysql.Base.Open()
	asserts.Error(err)
	asserts.Equal(query.ErrDbNotSet.Error(), err.Error())
}

// TestBase_CommitRollback_Without_TX checks if an error returns if Commit or Rollback is called on a nil sql.TX.
func TestBase_CommitRollback_Without_TX(t *testing.T) {
	asserts := assert.New(t)
	mysql := mysql{}
	err := mysql.Commit()
	asserts.Error(err)
	asserts.Equal(query.ErrNoTx.Error(), err.Error())

	err = mysql.Rollback()
	asserts.Error(err)
	asserts.Equal(query.ErrNoTx.Error(), err.Error())
}

// TestMysql_Timeout_Config checks the mysql timeout dns param.
func TestMysql_Timeout_Config(t *testing.T) {
	asserts := assert.New(t)

	// set a database and create a builder.
	cfg := testConfig().DB
	cfg.Host = "127.0.0.50"
	cfg.Timeout = "1s"
	b, err := query.New("mysql", cfg)
	asserts.Error(err)
	asserts.Nil(b)
	//asserts.True(strings.Contains(err.Error(), "timeout")) //TODO better solution to test a timeout.
}

// TestMysql_PreQuery_Config checks if the config pre-queries are executed.
func TestMysql_PreQuery_Config(t *testing.T) {
	asserts := assert.New(t)
	createDatabase(asserts)

	// ok: set a database and create a builder.
	cfg := testConfig().DB
	cfg.PreQuery = append(cfg.PreQuery, "DROP DATABASE IF EXISTS `tests`")
	b, err := query.New("mysql", cfg)
	asserts.NoError(err)
	asserts.NotNil(b)
	// check if db was dropped.
	row := b.Query().DB().QueryRow("select schema_name from information_schema.schemata where schema_name = 'tests';")
	err = row.Scan()
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows.Error(), err.Error())

	// error: wrong syntax.
	cfg = testConfig().DB
	cfg.PreQuery = append(cfg.PreQuery, "DROP DATA `tests`")
	b, err = query.New("mysql", cfg)
	asserts.Error(err)
	asserts.Nil(b)
}

// TestMysql_Query tests:
// - if a new instance will be creates (tx must be different)
func TestMysql_Query(t *testing.T) {
	asserts := assert.New(t)
	createDatabase(asserts)

	cfg := testConfig().DB
	cfg.Database = "tests"
	b, err := query.New("mysql", cfg)
	asserts.NoError(err)
	createTable(b, asserts)

	// transaction #1
	tx, err := b.Query().Tx()
	asserts.NoError(err)
	row, err := tx.Select("tests").Columns("id").First()
	asserts.NoError(err)
	var id int
	err = row.Scan(&id)

	// transaction #2
	tx2, err := b.Query().Tx()
	asserts.NoError(err)
	row, err = tx2.Select("tests").Columns("id").First()
	asserts.NoError(err)
	err = row.Scan(&id)

	// check if tx differs
	asserts.NotEqual(fmt.Sprintf("%p", tx.(*mysql).TransactionBase.Tx), fmt.Sprintf("%p", tx2.(*mysql).TransactionBase.Tx))
	asserts.NoError(tx.Commit())
	asserts.NoError(tx2.Commit())

	// should have the same db connection
	asserts.Equal(tx.DB(), tx2.DB())
}

// TestMysql tests:
// - logger
// - insert
// - select
// - update
// - delete
// - information
func TestMysql(t *testing.T) {
	asserts := assert.New(t)
	// create database.
	createDatabase(asserts)
	// set a database and create a builder.
	cfg := testConfig().DB
	cfg.Database = "tests"
	b, err := query.New("mysql", cfg)
	asserts.NoError(err)
	asserts.NotNil(b)
	// create test table
	createTable(b, asserts)

	testLogger(b, t)
	testInsert(b, asserts)
	testSelect(b, asserts)
	testUpdate(b, asserts)
	testDelete(b, asserts)
	testInformationDescribe(b, t, asserts)
	testInformationForeignKey(b, asserts)
}

// testLogger tests:
// - if the logger is called on First
// - if the logger is called on All
// - if the logger is called on Exec
func testLogger(b query.Builder, t *testing.T) {
	logger := new(mocks.Manager)
	b.SetLogger(logger)

	logger.On("WithTimer").Once().Return(logger)
	logger.On("Debug", "SELECT `*` FROM `query`").Once().Return()
	_, _ = b.Query().Select("query").First()

	logger.On("WithTimer").Once().Return(logger)
	logger.On("Debug", "SELECT `*` FROM `query`").Once().Return()
	_, _ = b.Query().Select("query").All()

	logger.On("WithTimer").Once().Return(logger)
	logger.On("Debug", "DELETE FROM `query`").Once().Return()
	_, _ = b.Query().Delete("query").Exec()

	b.SetLogger(nil)
	logger.AssertExpectations(t)
}

// testInsert tests:
// - insert of a single value.
// - insert of multiple values.
// - insert with a batch.
// - error: insert with batch - rollback (wrong column type).
// - error: no values are set.
// - insert only defined columns.
// - insert only defined columns with batch.
// - error: column does not exist.
// - last inserted id.
// - last inserted id with no ptr.
// - manually added tx commit.
// - manually added tx rollback.
func testInsert(b query.Builder, asserts *assert.Assertions) {

	// ok: insert single value
	truncateTestTable(b, asserts)
	v := []map[string]interface{}{{"int": 1, "varchar": "a"}}
	res, err := b.Query().Insert("query").Values(v).Exec()
	asserts.NoError(err)
	asserts.Equal(1, len(res))
	rowsAffected, err := res[0].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), rowsAffected)
	// test string
	stmt, args, err := b.Query().Insert("query").Columns("int", "varchar").Values(v).String()
	asserts.NoError(err)
	asserts.Equal([]string{"INSERT INTO `query`(`int`, `varchar`) VALUES (?, ?)"}, stmt)
	asserts.Equal([][]interface{}{{1, "a"}}, args)
	// check result
	asserts.Equal(1, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(1), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal(true, checkInsertResult(b, asserts)[0].Varchar.Valid)

	// ok: insert multiple values
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "b"}, {"int": 3, "varchar": "c"}, {"int": 4, "varchar": "d"}}
	res, err = b.Query().Insert("query").Values(v).Exec()
	asserts.NoError(err)
	asserts.Equal(1, len(res))
	rowsAffected, err = res[0].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(3), rowsAffected)
	// test string
	stmt, args, err = b.Query().Insert("query").Columns("int", "varchar").Values(v).String()
	asserts.NoError(err)
	asserts.Equal([]string{"INSERT INTO `query`(`int`, `varchar`) VALUES (?, ?), (?, ?), (?, ?)"}, stmt)
	asserts.Equal([][]interface{}{{2, "b", 3, "c", 4, "d"}}, args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(true, checkInsertResult(b, asserts)[0].Varchar.Valid)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(true, checkInsertResult(b, asserts)[1].Varchar.Valid)
	asserts.Equal(int64(4), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("d", checkInsertResult(b, asserts)[2].Varchar.String)
	asserts.Equal(true, checkInsertResult(b, asserts)[2].Varchar.Valid)

	// ok: insert multiple values with batch
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2}, {"int": 3}, {"int": 4}}
	res, err = b.Query().Insert("query").Batch(2).Values(v).Exec()
	asserts.NoError(err)
	asserts.Equal(2, len(res))
	rowsAffected, err = res[0].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(2), rowsAffected)
	rowsAffected, err = res[1].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), rowsAffected)
	// test string
	stmt, args, err = b.Query().Insert("query").Batch(2).Values(v).String()
	asserts.NoError(err)
	asserts.Equal([]string{"INSERT INTO `query`(`int`) VALUES (?), (?)", "INSERT INTO `query`(`int`) VALUES (?)"}, stmt)
	asserts.Equal([][]interface{}{{2, 3}, {4}}, args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal(false, checkInsertResult(b, asserts)[0].Varchar.Valid)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal(false, checkInsertResult(b, asserts)[1].Varchar.Valid)
	asserts.Equal(int64(4), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal(false, checkInsertResult(b, asserts)[2].Varchar.Valid)

	// error: insert multiple values with batch - rollback - wrong column type
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2}, {"int": "a"}, {"int": 4}}
	res, err = b.Query().Insert("tests").Batch(2).Values(v).Exec()
	asserts.Error(err)
	asserts.Equal(0, len(res))
	// test string
	stmt, args, err = b.Query().Insert("query").Batch(2).Values(v).String()
	asserts.NoError(err)
	asserts.Equal([]string{"INSERT INTO `query`(`int`) VALUES (?), (?)", "INSERT INTO `query`(`int`) VALUES (?)"}, stmt)
	asserts.Equal([][]interface{}{{2, "a"}, {4}}, args)
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))

	// error: no values are set
	truncateTestTable(b, asserts)
	res, err = b.Query().Insert("query").Exec()
	asserts.Error(err)
	asserts.Nil(res)
	asserts.Equal(fmt.Sprintf(query.ErrValueMissing, "insert", "tests.query"), err.Error())
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))

	// ok: only insert defined columns
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}, {"int": 3, "varchar": "b"}, {"int": 4, "varchar": "c"}}
	res, err = b.Query().Insert("query").Columns("varchar").Values(v).Exec()
	asserts.NoError(err)
	asserts.Equal(1, len(res))
	rowsAffected, err = res[0].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(3), rowsAffected)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(false, checkInsertResult(b, asserts)[0].Int.Valid)
	asserts.Equal(false, checkInsertResult(b, asserts)[1].Int.Valid)
	asserts.Equal(false, checkInsertResult(b, asserts)[2].Int.Valid)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// ok: only insert defined columns with batch
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}, {"int": 3, "varchar": "b"}, {"int": 4, "varchar": "c"}}
	res, err = b.Query().Insert("query").Columns("varchar").Batch(2).Values(v).Exec()
	asserts.NoError(err)
	asserts.Equal(2, len(res))
	rowsAffected, err = res[0].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(2), rowsAffected)
	rowsAffected, err = res[1].RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), rowsAffected)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(false, checkInsertResult(b, asserts)[0].Int.Valid)
	asserts.Equal(false, checkInsertResult(b, asserts)[1].Int.Valid)
	asserts.Equal(false, checkInsertResult(b, asserts)[2].Int.Valid)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// error: column does not exist
	truncateTestTable(b, asserts)
	res, err = b.Query().Insert("query").Columns("notExisting").Values(v).Exec()
	asserts.Error(err)
	asserts.Nil(res)
	asserts.Equal(fmt.Sprintf(query.ErrColumn, "notExisting", "tests.query"), err.Error())
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))

	// ok: last inserted ID
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}}
	var lastID int64
	res, err = b.Query().Insert("query").Columns("varchar").Values(v).LastInsertedID(&lastID).Exec()
	asserts.NoError(err)
	asserts.Equal(int64(1), lastID)
	// check result
	asserts.Equal(1, len(checkInsertResult(b, asserts)))
	asserts.Equal(false, checkInsertResult(b, asserts)[0].Int.Valid)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)

	// error: last inserted ID is no ptr value
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}}
	res, err = b.Query().Insert("query").Columns("varchar").Values(v).LastInsertedID(lastID).Exec()
	asserts.Error(err)
	asserts.Equal(query.ErrLastID.Error(), err.Error())
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))

	// ok: manually added TX with Commit
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}}
	tx, err := b.Query().Tx()
	asserts.NoError(err)
	res, err = tx.Insert("query").Columns("varchar").Values(v).LastInsertedID(&lastID).Exec()
	asserts.NoError(err)
	asserts.Equal(int64(1), lastID)
	err = tx.Commit()
	asserts.NoError(err)
	// check result
	asserts.Equal(1, len(checkInsertResult(b, asserts)))
	asserts.Equal(false, checkInsertResult(b, asserts)[0].Int.Valid)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)

	// ok: manually added TX with Rollback
	truncateTestTable(b, asserts)
	v = []map[string]interface{}{{"int": 2, "varchar": "a"}}
	tx, err = b.Query().Tx()
	asserts.NoError(err)
	res, err = tx.Insert("query").Columns("varchar").Values(v).LastInsertedID(&lastID).Exec()
	asserts.NoError(err)
	asserts.Equal(int64(1), lastID)
	err = tx.Rollback()
	asserts.NoError(err)
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))
}

// testSelect tests:
// - First normal select
// - First with tx
// - All with tx
// - First with manually added condition
// - error: First render (condition,argument mismatch)
// - error: All render (condition,argument mismatch)
// - test column wild card (*).
// - testing all with limit, offset
// - set condition directly
// _ testing join conditions
func testSelect(b query.Builder, asserts *assert.Assertions) {

	// ok: First test
	createTestEntries(b, asserts)
	stmt := b.Query().Select("query").Columns("int", "varchar")
	row, err := stmt.First()
	asserts.NoError(err)
	var intVal int
	var varcharVal string
	err = row.Scan(&intVal, &varcharVal)
	asserts.NoError(err)
	asserts.Equal(1, intVal)
	asserts.Equal("a", varcharVal)
	// string
	sqlStmt, args, err := stmt.String()
	asserts.NoError(err)
	asserts.Equal("SELECT `int`, `varchar` FROM `query`", sqlStmt)
	asserts.Equal([]interface{}(nil), args)

	// ok: First with TX
	createTestEntries(b, asserts)
	tx, err := b.Query().Tx()
	stmt = tx.Select("query").Columns("int", "varchar")
	row, err = stmt.First()
	asserts.NoError(err)
	err = row.Scan(&intVal, &varcharVal)
	asserts.NoError(err)
	asserts.Equal(1, intVal)
	asserts.Equal("a", varcharVal)
	// string
	sqlStmt, args, err = stmt.String()
	asserts.NoError(err)
	asserts.Equal("SELECT `int`, `varchar` FROM `query`", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	err = tx.Rollback()
	asserts.NoError(err)

	// ok: All with TX
	createTestEntries(b, asserts)
	tx, err = b.Query().Tx()
	stmt = tx.Select("query").Columns("int", "varchar")
	rows, err := stmt.All()
	i := 0
	for rows.Next() {
		var Int int
		var Varchar string
		err = rows.Scan(&Int, &Varchar)
		asserts.NoError(err)
		asserts.Equal(i+1, Int)
		switch i {
		case 0:
			asserts.Equal(1, Int)
			asserts.Equal("a", Varchar)
		case 1:
			asserts.Equal(2, Int)
			asserts.Equal("b", Varchar)
		case 2:
			asserts.Equal(3, Int)
			asserts.Equal("c", Varchar)
		}
		i++
	}
	asserts.Equal(3, i)
	err = tx.Rollback()
	asserts.NoError(err)

	// ok: First with manually added conditions.
	// LIMIT and OFFSET must be removed (First)
	createTestEntries(b, asserts)
	stmt = b.Query().Select("query").Columns("int", "varchar").Where("id = ?", 1).Group("id").Order("-id").Limit(1).Offset(0).Having("id=1")
	row, err = stmt.First()
	asserts.NoError(err)
	err = row.Scan(&intVal, &varcharVal)
	asserts.NoError(err)
	asserts.Equal(1, intVal)
	asserts.Equal("a", varcharVal)
	// string
	sqlStmt, args, err = stmt.String()
	asserts.NoError(err)
	asserts.Equal("SELECT `int`, `varchar` FROM `query` WHERE id = ? GROUP BY id HAVING id=1 ORDER BY id DESC", sqlStmt)
	asserts.Equal([]interface{}{1}, args)

	// error: First render error (condition,argument mismatch)
	createTestEntries(b, asserts)
	row, err = b.Query().Select("query").Columns("int", "varchar").Where("id = ? ?", 1).First()
	asserts.Nil(row)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrPlaceholderMismatch, "id = ? ?", 2, 1), err.Error())

	// error: all with render error
	rows, err = b.Query().Select("query").Where("id = ? ?", 1).All()
	asserts.Nil(rows)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrPlaceholderMismatch, "id = ? ?", 2, 1), err.Error())

	// ok: test column wild card.
	createTestEntries(b, asserts)
	stmt = b.Query().Select("query").Columns("int", "varchar")
	rows, err = stmt.All()
	asserts.NoError(err)
	i = 0
	for rows.Next() {
		var Int int
		var Varchar string
		err = rows.Scan(&Int, &Varchar)
		asserts.NoError(err)
		asserts.Equal(i+1, Int)
		switch i {
		case 0:
			asserts.Equal(1, Int)
			asserts.Equal("a", Varchar)
		case 1:
			asserts.Equal(2, Int)
			asserts.Equal("b", Varchar)
		case 2:
			asserts.Equal(3, Int)
			asserts.Equal("c", Varchar)
		}
		i++
	}
	asserts.Equal(3, i)
	// string
	sqlStmt, args, err = stmt.Columns().String()
	asserts.NoError(err)
	asserts.Equal("SELECT `*` FROM `query`", sqlStmt)
	asserts.Equal([]interface{}(nil), args)

	// ok: testing all with limit, offset
	sqlStmt, args, err = b.Query().Select("query").Limit(1).Offset(2).String()
	asserts.NoError(err)
	asserts.Equal("SELECT `*` FROM `query` LIMIT 1 OFFSET 2", sqlStmt)
	asserts.Equal([]interface{}(nil), args)

	// ok: set condition directly
	sqlStmt, args, err = b.Query().Select("query").Condition(condition.New().SetWhere("id = ?", 1)).Limit(10).String()
	asserts.NoError(err)
	asserts.Equal("SELECT `*` FROM `query` WHERE id = ? LIMIT 10", sqlStmt)
	asserts.Equal([]interface{}{1}, args)

	// ok testing join conditions
	sqlStmt, args, err = b.Query().Select("query").Where("id = ?", 1).
		Join(condition.LEFT, "query2", "a=?", 10).Order("-id").
		Join(condition.INNER, "query4", "c=?", 30).
		Join(condition.RIGHT, "query3", "b=?", 20).
		Join(condition.CROSS, "query5", "").
		Offset(10).String()
	asserts.NoError(err)
	asserts.Equal("SELECT `*` FROM `query` LEFT JOIN `query2` ON a=? INNER JOIN `query4` ON c=? RIGHT JOIN `query3` ON b=? CROSS JOIN `query5` WHERE id = ? ORDER BY id DESC OFFSET 10", sqlStmt)
	asserts.Equal([]interface{}{10, 30, 20, 1}, args)
}

// testUpdate tests:
// - normal update.
// - update without where.
// - update with where.
// - error: column does not exist in db.
// - error: defined column does not exist in value map.
// - error: value is not set.
// - error: condition render error.
// - set manually condition - everything gets deleted except where.
func testUpdate(b query.Builder, asserts *assert.Assertions) {

	// ok: Update only int
	createTestEntries(b, asserts)
	sqlQuery := b.Query().Update("query").Columns("int").Set(map[string]interface{}{"int": 10, "varchar": "aa"}).Where("id = ?", 1)
	res, err := sqlQuery.Exec()
	asserts.NoError(err)
	affected, err := res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), affected)
	// string
	sqlStmt, args, err := sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("UPDATE `query` SET `int` = ? WHERE id = ?", sqlStmt)
	asserts.Equal([]interface{}{10, 1}, args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// ok: Update with where
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query").Set(map[string]interface{}{"int": 10, "varchar": "aa"}).Where("id = ?", 1)
	res, err = sqlQuery.Exec()
	asserts.NoError(err)
	affected, err = res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), affected)
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.True(sqlStmt == "UPDATE `query` SET `int` = ?, `varchar` = ? WHERE id = ?" || sqlStmt == "UPDATE `query` SET `varchar` = ?, `int` = ? WHERE id = ?")
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("aa", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// ok: Update without where
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query").Set(map[string]interface{}{"int": 10, "varchar": "aa"})
	res, err = sqlQuery.Exec()
	asserts.NoError(err)
	affected, err = res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(3), affected)
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.True(sqlStmt == "UPDATE `query` SET `int` = ?, `varchar` = ?" || sqlStmt == "UPDATE `query` SET `varchar` = ?, `int` = ?")
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("aa", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("aa", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("aa", checkInsertResult(b, asserts)[2].Varchar.String)

	// error: column does not exist in db
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query").Set(map[string]interface{}{"notExisting": 10})
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal("Error 1054: Unknown column 'notExisting' in 'field list'", err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("UPDATE `query` SET `notExisting` = ?", sqlStmt)
	asserts.Equal([]interface{}{10}, args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(1), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// error: defined column does not exist in map
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query").Columns("notExisting").Set(map[string]interface{}{"int": 10})
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(query.ErrColumn, "notExisting", "query"), err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.Error(err)
	asserts.Equal("", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(1), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// error: value is not set
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query")
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(query.ErrValueMissing, "update", "query"), err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.Error(err)
	asserts.Equal("", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(1), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// error: condition render error
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Update("query").Set(map[string]interface{}{"int": 10}).Where("a=??", 1)
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrPlaceholderMismatch, "a=??", 2, 1), err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.Error(err)
	asserts.Equal("", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(1), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

	// ok: set condition, check if everything except where gets deleted.
	createTestEntries(b, asserts)
	c := condition.New().SetWhere("id = ?", 1).SetOffset(10).SetLimit(1).SetOrder("id").SetGroup("id").SetHaving("id").SetJoin(condition.LEFT, "a", "a=b")
	sqlQuery = b.Query().Update("query").Set(map[string]interface{}{"int": 10}).Condition(c)
	res, err = sqlQuery.Exec()
	asserts.NoError(err)
	affected, err = res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), affected)
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("UPDATE `query` SET `int` = ? WHERE id = ?", sqlStmt)
	asserts.Equal([]interface{}{10, 1}, args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
	asserts.Equal(int64(10), checkInsertResult(b, asserts)[0].Int.Int64)
	asserts.Equal("a", checkInsertResult(b, asserts)[0].Varchar.String)
	asserts.Equal(int64(2), checkInsertResult(b, asserts)[1].Int.Int64)
	asserts.Equal("b", checkInsertResult(b, asserts)[1].Varchar.String)
	asserts.Equal(int64(3), checkInsertResult(b, asserts)[2].Int.Int64)
	asserts.Equal("c", checkInsertResult(b, asserts)[2].Varchar.String)

}

// testDelete tests:
// - delete with where.
// - delete without where.
// - delete with condition.
// - error: table does not exist.
// - error: render mismatch.
func testDelete(b query.Builder, asserts *assert.Assertions) {

	// ok: delete with where
	createTestEntries(b, asserts)
	sqlQuery := b.Query().Delete("query").Where("id = ?", 1)
	res, err := sqlQuery.Exec()
	asserts.NoError(err)
	affected, err := res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), affected)
	// string
	sqlStmt, args, err := sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("DELETE FROM `query` WHERE id = ?", sqlStmt)
	asserts.Equal([]interface{}{1}, args)
	// check result
	asserts.Equal(2, len(checkInsertResult(b, asserts)))

	// ok: delete without where
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Delete("query")
	res, err = sqlQuery.Exec()
	asserts.NoError(err)
	affected, err = res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(3), affected)
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("DELETE FROM `query`", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(0, len(checkInsertResult(b, asserts)))

	// ok: delete with condition
	createTestEntries(b, asserts)
	c := condition.New().SetWhere("id = ?", 1).SetOffset(10).SetLimit(1).SetOrder("id").SetGroup("id").SetHaving("id").SetJoin(condition.LEFT, "a", "a=b")
	sqlQuery = b.Query().Delete("query").Condition(c)
	res, err = sqlQuery.Exec()
	asserts.NoError(err)
	affected, err = res.RowsAffected()
	asserts.NoError(err)
	asserts.Equal(int64(1), affected)
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("DELETE FROM `query` WHERE id = ?", sqlStmt)
	asserts.Equal([]interface{}{1}, args)
	// check result
	asserts.Equal(2, len(checkInsertResult(b, asserts)))

	// error: table does not exist
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Delete("query2")
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal("Error 1146: Table 'tests.query2' doesn't exist", err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.NoError(err)
	asserts.Equal("DELETE FROM `query2`", sqlStmt)
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))

	// error: render mismatch
	createTestEntries(b, asserts)
	sqlQuery = b.Query().Delete("query").Where("id=??", 1)
	res, err = sqlQuery.Exec()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrPlaceholderMismatch, "id=??", 2, 1), err.Error())
	// string
	sqlStmt, args, err = sqlQuery.String()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrPlaceholderMismatch, "id=??", 2, 1), err.Error())
	asserts.Equal([]interface{}(nil), args)
	// check result
	asserts.Equal(3, len(checkInsertResult(b, asserts)))
}

// testInformationDescribe tests:
// - desc only one column.
// - desc all columns.
// - error: column does not exist.
// - error: one requested column does not exist of list.
// - checking all column types.
func testInformationDescribe(b query.Builder, t *testing.T, asserts *assert.Assertions) {

	// ok: getting only one column
	cols, err := b.Query().Information("query").Describe("id")
	asserts.NoError(err)
	asserts.Equal(1, len(cols))
	expectedCol := query.Column{Table: "query", Type: cols[0].Type, Name: "id", Position: 1, NullAble: false, PrimaryKey: true, Unique: false, DefaultValue: query.NewNullString("", false), Length: query.NewNullInt(0, false), Autoincrement: true}
	asserts.Equal(expectedCol, cols[0])
	asserts.Equal("Integer", cols[0].Type.Kind())
	asserts.Equal("int(11) unsigned", cols[0].Type.Raw())

	// error: column does not exist
	cols, err = b.Query().Information("query").Describe("notExisting")
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrTableDoesNotExist, "tests.query", []string{"notExisting"}), err.Error())
	asserts.Nil(cols)

	// ok: getting all types
	cols, err = b.Query().Information("query").Describe()
	asserts.NoError(err)
	asserts.Equal(26, len(cols))
	var tests = []struct {
		Table         string
		TypeKind      string
		TypeRaw       string
		Name          string
		Position      int
		NullAble      bool
		PrimaryKey    bool
		Unique        bool
		DefaultValue  query.NullString
		Length        query.NullInt
		Autoincrement bool
	}{
		{Table: "query", Name: "id", TypeKind: "Integer", TypeRaw: "int(11) unsigned", Position: 1, NullAble: false, PrimaryKey: true, Unique: false, Autoincrement: true},
		{Table: "query", Name: "int", TypeKind: "Integer", TypeRaw: "int(11)", Position: 2, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "varchar", TypeKind: "Text", TypeRaw: "varchar(250)", Position: 3, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(250, true), Autoincrement: false},
		{Table: "query", Name: "tinyint", TypeKind: "Integer", TypeRaw: "tinyint(4)", Position: 4, NullAble: true, PrimaryKey: false, Unique: false, DefaultValue: query.NewNullString("4", true), Autoincrement: false},
		{Table: "query", Name: "smallint", TypeKind: "Integer", TypeRaw: "smallint(6)", Position: 5, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "mediumint", TypeKind: "Integer", TypeRaw: "mediumint(9)", Position: 6, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "bigint", TypeKind: "Integer", TypeRaw: "bigint(20)", Position: 7, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "float", TypeKind: "Float", TypeRaw: "float", Position: 8, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "double", TypeKind: "Float", TypeRaw: "double", Position: 9, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "char", TypeKind: "Text", TypeRaw: "char(1)", Position: 10, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(1, true), Autoincrement: false},
		{Table: "query", Name: "tinytext", TypeKind: "TextArea", TypeRaw: "tinytext", Position: 11, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(255, true), Autoincrement: false},
		{Table: "query", Name: "text", TypeKind: "TextArea", TypeRaw: "text", Position: 12, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(65535, true), Autoincrement: false},
		{Table: "query", Name: "mediumtext", TypeKind: "TextArea", TypeRaw: "mediumtext", Position: 13, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(16777215, true), Autoincrement: false},
		{Table: "query", Name: "longtext", TypeKind: "TextArea", TypeRaw: "longtext", Position: 14, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(4294967295, true), Autoincrement: false},
		{Table: "query", Name: "enum", TypeKind: "Select", TypeRaw: "enum('JOHN','DOE')", Position: 15, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(4, true), Autoincrement: false},
		{Table: "query", Name: "set", TypeKind: "MultiSelect", TypeRaw: "set('FOO','BAR')", Position: 16, NullAble: true, PrimaryKey: false, Unique: false, Length: query.NewNullInt(7, true), Autoincrement: false},
		{Table: "query", Name: "date", TypeKind: "Date", TypeRaw: "date", Position: 17, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "datetime", TypeKind: "DateTime", TypeRaw: "datetime", Position: 18, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "timestamp", TypeKind: "DateTime", TypeRaw: "timestamp", Position: 19, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "bool", TypeKind: "Bool", TypeRaw: "tinyint(1)", Position: 20, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},

		{Table: "query", Name: "utinyint", TypeKind: "Integer", TypeRaw: "tinyint(3) unsigned", Position: 21, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "usmallint", TypeKind: "Integer", TypeRaw: "smallint(5) unsigned", Position: 22, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "umediumint", TypeKind: "Integer", TypeRaw: "mediumint(8) unsigned", Position: 23, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "ubigint", TypeKind: "Integer", TypeRaw: "bigint(20) unsigned", Position: 24, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "time", TypeKind: "Time", TypeRaw: "time", Position: 25, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
		{Table: "query", Name: "geometry", TypeKind: "", TypeRaw: "", Position: 26, NullAble: true, PrimaryKey: false, Unique: false, Autoincrement: false},
	}

	for i, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			asserts.Equal(test.Table, cols[i].Table)
			asserts.Equal(test.Name, cols[i].Name)
			if test.Name == "geometry" { // undefined typeMapping type
				asserts.Nil(cols[i].Type)
			} else {
				if test.Name == "enum" {
					items := cols[i].Type.(types.Items)
					asserts.Equal([]string{"JOHN", "DOE"}, items.Items())
				}
				if test.Name == "set" {
					items := cols[i].Type.(types.Items)
					asserts.Equal([]string{"FOO", "BAR"}, items.Items())
				}
				asserts.Equal(test.TypeKind, cols[i].Type.Kind())
				asserts.Equal(test.TypeRaw, cols[i].Type.Raw())
			}

			asserts.Equal(test.Position, cols[i].Position)
			asserts.Equal(test.NullAble, cols[i].NullAble)
			asserts.Equal(test.PrimaryKey, cols[i].PrimaryKey)
			asserts.Equal(test.Unique, cols[i].Unique)
			asserts.Equal(test.DefaultValue, cols[i].DefaultValue)
			asserts.Equal(test.Length, cols[i].Length)
			asserts.Equal(test.Autoincrement, cols[i].Autoincrement)
		})
	}
}

// testInformationForeignKey tests:
// - error: relation does not exist
// - error: table does not exist
// - getting the fks
func testInformationForeignKey(b query.Builder, asserts *assert.Assertions) {
	// error: relation does not exist
	fk, err := b.Query().Information("query").ForeignKey()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrTableRelation, "query"), err.Error())

	// error: table does not exist
	fk, err = b.Query().Information("query2").ForeignKey()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrTableRelation, "query2"), err.Error())

	// ok: getting fks
	fk, err = b.Query().Information("query_fk").ForeignKey()
	asserts.NoError(err)
	asserts.Equal(1, len(fk))
	asserts.Equal("query_fk_ibfk_1", fk[0].Name)
	asserts.Equal(query.Relation{Table: "query_fk", Column: "id"}, fk[0].Primary)
	asserts.Equal(query.Relation{Table: "query", Column: "id"}, fk[0].Secondary)

}
