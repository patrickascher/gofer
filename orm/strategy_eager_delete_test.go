// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"database/sql"
	"testing"

	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
)

// TestEager_Delete_SelfRef tests:
// - If only the main model gets deleted and everything else by reference.
func TestEager_Delete_SelfRef(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// init orm model
	role := Role{}
	err := role.Init(&role)
	asserts.NoError(err)

	var roles []Role
	// check db roles (id 1,2,3)
	err = role.All(&roles, condition.New().SetWhere("id IN (?)", []int{1, 2, 3}))
	asserts.NoError(err)
	asserts.Equal(3, len(roles))

	// delete id 1 which has the relation to id 2 and 3.
	err = role.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	err = role.Delete()
	asserts.NoError(err)

	// check db roles (id 1,2,3) again.
	// Logic at the moment is delete first ID, and m2m relations will only references in the junction table deleted.
	err = role.All(&roles, condition.New().SetWhere("id IN (?)", []int{1, 2, 3}))
	asserts.NoError(err)
	asserts.Equal(2, len(roles))
}

// TestEager_Delete_SoftDelete tests:
// - If a orm is getting soft deleted, if defined.
func TestEager_Delete_SoftDelete(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// init orm model
	animal := Animal{}
	err := animal.Init(&animal)
	asserts.NoError(err)

	// fetch id 1
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)

	// delete entry.
	err = animal.Delete()
	asserts.NoError(err)

	// no result because its soft deleted and not shown by default.
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows, err)

	// soft deletion is included in scope.
	scope, err := animal.Scope()
	asserts.NoError(err)
	scope.SetConfig(orm.NewConfig().SetShowDeletedRows(true))
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	helperTestResults(asserts, err, helperTestCases()[0], animal, false)
}

// TestEager_Delete tests:
// - If the orm gets completely deleted if no soft_delete exists.
// - If all relations are getting deleted correctly (hasOne, hasMany - all, m2m, belongsTo only reference)
func TestEager_Delete(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// drop deleted_at column
	_, err := builder.Query().DB().Exec("ALTER TABLE `animals` DROP `deleted_at`;")
	asserts.NoError(err)

	// delete existing cache because of the saved field (deleted_at).
	err = c.Delete("orm_", "orm_test.Animal")
	asserts.NoError(err)

	// init orm model
	animal := Animal{}
	err = animal.Init(&animal)
	asserts.NoError(err)

	// fetch entry
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)

	// delete entry
	err = animal.Delete()
	asserts.NoError(err)

	// check if deleted
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows, err)

	// test if all hasOne relations are deleted
	adr := Address{}
	err = adr.Init(&adr)
	asserts.NoError(err)
	err = adr.First(condition.New().SetWhere("id = ?", 1))
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows, err)

	adrPoly := AddressPoly{}
	err = adrPoly.Init(&adrPoly)
	asserts.NoError(err)
	err = adrPoly.First(condition.New().SetWhere("id = ?", 1))
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows, err)

	// test if all hasMany relations are deleted
	toy := Toy{}
	var toys []Toy
	err = toy.Init(&toy)
	asserts.NoError(err)
	err = toy.All(&toys, condition.New().SetWhere("animal_id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(0, len(toys))

	toyPoly := ToyPoly{}
	var toyPolies []ToyPoly
	err = toyPoly.Init(&toyPoly)
	asserts.NoError(err)
	err = toyPoly.All(&toyPolies, condition.New().SetWhere("animal_id = ? AND toy_type = ?", 1, "Animal"))
	asserts.NoError(err)
	asserts.Equal(0, len(toyPolies))

	// belongsTo and m2m only deletes the references.
	res, err := builder.Query().Select("animal_walkers").Columns("animal_id").Where("animal_id = ?", 1).All()
	asserts.NoError(err)
	rows := 0
	for res.Next() {
		var id int
		err = res.Scan(&id)
		asserts.NoError(err)
		rows++
	}
	err = res.Close()
	asserts.Equal(0, rows)

	res, err = builder.Query().Select("animal_walker_polies").Columns("animal_id", "animal_type").Where("animal_id = ? AND animal_type=?", 1, "Fast").All()
	asserts.NoError(err)
	rows = 0
	for res.Next() {
		var id int
		err = res.Scan(&id)
		asserts.NoError(err)
		rows++
	}
	err = res.Close()
	asserts.Equal(0, rows)
}
