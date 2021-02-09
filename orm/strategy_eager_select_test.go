// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
)

// TestEager_First_LoopDetection tests:
// - If self referencing models return the correct result.
// - If an error returns if a db loop is set.
func TestEager_First_DBLoopDetection(t *testing.T) {
	asserts := assert.New(t)

	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// Test preCache.
	err := orm.PreInit(&Role{})
	asserts.NoError(err)

	// Init user model.
	role := Role{}
	err = role.Init(&role)
	asserts.NoError(err)

	// ok - no loop exists RoleA has Role B has Role C.
	err = role.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal("RoleA", role.Name)
	asserts.Equal(1, len(role.Roles))
	asserts.Equal("RoleB", role.Roles[0].Name)
	asserts.Equal(1, len(role.Roles[0].Roles))
	asserts.Equal("RoleC", role.Roles[0].Roles[0].Name)
	asserts.Equal(0, len(role.Roles[0].Roles[0].Roles))

	// infinity loop Loop1 has Loop2 has Loop1 = infinity loop.
	err = role.First(condition.New().SetWhere("id = ?", 4))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrInfinityLoop, "orm_test.Role"), err.Error())
}

// TestEager_All_DBLoopDetection tests:
// - If self referencing models return the correct result.
// - If an error returns if a db loop is set.
func TestEager_All_DBLoopDetection(t *testing.T) {
	asserts := assert.New(t)

	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// Test preCache.
	err := orm.PreInit(&Role{})
	asserts.NoError(err)

	// Init user model.
	role := Role{}
	err = role.Init(&role)
	asserts.NoError(err)

	// ok - no loop for id 1,2
	var roles []Role
	err = role.All(&roles, condition.New().SetWhere("id IN (?)", []int{1, 2}))
	asserts.NoError(err)
	asserts.Equal(2, len(roles))
	asserts.Equal(1, roles[0].ID)
	asserts.Equal("RoleA", roles[0].Name)
	asserts.Equal(1, len(roles[0].Roles))
	asserts.Equal(1, len(roles[0].Roles))
	asserts.Equal("RoleB", roles[0].Roles[0].Name)
	asserts.Equal(1, len(roles[0].Roles[0].Roles))
	asserts.Equal("RoleC", roles[0].Roles[0].Roles[0].Name)
	asserts.Equal(0, len(roles[0].Roles[0].Roles[0].Roles))

	asserts.Equal(2, roles[1].ID)
	asserts.Equal("RoleB", roles[1].Name)
	asserts.Equal(1, len(roles[1].Roles))
	asserts.Equal(1, len(roles[1].Roles))
	asserts.Equal("RoleC", roles[1].Roles[0].Name)
	asserts.Equal(0, len(roles[1].Roles[0].Roles))

	// error: id 4 infinity loop.
	err = role.All(&roles)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrInfinityLoop, "orm_test.Role"), err.Error())
}

// TestEager_First tests:
// - If Init, Re-init, Scope returns no error and the amount of relations is correct.
// - If hasOne returns an error if hasOne is mandatory and the result is empty.
// - If soft deleted rows will be displayed if required and if they will be removed by default.
// - If custom relation configurations will be applied. Merged or as a new condition.
// - If First can be called multiple times with the correct result.
// - If First relations: hasOne, belongsTo, hasMany and manyToMany will return the correct result. (struct, slice, ptr, ptr-slice, slice-ptr, ptr-slice-ptr)
// - If Back-references are working correctly (slice backref - working) struct (backref - TODO will return an error atm because of the validation.V10 package. orm could handle it)
func TestEager_First(t *testing.T) {
	asserts := assert.New(t)

	// delete existing cache because of the saved field (deleted_at).
	if c.Exist("orm_", "orm_test.Animal") {
		err := c.Delete("orm_", "orm_test.Animal")
		asserts.NoError(err)
	}

	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// Test preCache.
	err := orm.PreInit(&Animal{})
	asserts.NoError(err)

	animal := Animal{}
	// Init user model.
	err = animal.Init(&animal)
	asserts.NoError(err)

	// 2nd init test.
	err = animal.Init(&animal)
	asserts.NoError(err)

	// get scope
	s, err := animal.Scope()
	asserts.NoError(err)

	// testing the amount of defined relations
	asserts.Equal(24, len(s.SQLRelations(orm.Permission{Read: true})))

	// run test cases
	for _, test := range helperTestCases() {
		t.Run("ID:"+strconv.Itoa(test.fetchID), func(t *testing.T) {
			err = animal.First(condition.New().SetWhere("id = ?", test.fetchID))
			helperTestResults(asserts, err, test, animal, false)
			// calling First a second time to guarantee everything is copied/reset correctly.
			err = animal.First(condition.New().SetWhere("id = ?", test.fetchID))
			helperTestResults(asserts, err, test, animal, false)
		})
	}

	// error if HasOne row is required.
	s.SetConfig(orm.NewConfig().SetAllowHasOneZero(false), "Address")
	err = animal.First(condition.New().SetWhere("id = ?", 3))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(orm.ErrNoRows, s.FqdnModel("Address"), sql.ErrNoRows).Error(), err.Error())

	// check if soft deleted rows will be displayed.
	s.SetConfig(orm.NewConfig().SetAllowHasOneZero(true), "Address")
	err = animal.First(condition.New().SetWhere("id = ?", 4))
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows, err)
	s.SetConfig(orm.NewConfig().SetShowDeletedRows(true))
	err = animal.First(condition.New().SetWhere("id = ?", 4))
	asserts.NoError(err)
	asserts.Equal("Nala", animal.Name)

	// test custom relation condition
	s.SetConfig(orm.NewConfig().SetAllowHasOneZero(true).SetCondition(condition.New().SetWhere("deleted_at"), true), "Address")
	err = animal.First(condition.New().SetWhere("id = ?", 4))
	asserts.Error(err)
	asserts.Equal("orm: orm_test.Animal:Address Error 1054: Unknown column 'deleted_at' in 'where clause'", err.Error())
	s.SetConfig(orm.NewConfig().SetAllowHasOneZero(true).SetCondition(nil), "Address")

	// test custom merge relation condition, manipulate that AddressID 2 will be loaded instead the ID 1.
	s.SetConfig(orm.NewConfig().SetAllowHasOneZero(true).SetCondition(condition.New().SetWhere("1=1 OR id = 2").SetOrder("-id"), true), "Address")
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal("Brandschenkestrasse", animal.Address.Street.String)
}

// TestEager_All tests:
// - If result is a ptr to a slice orm.Interface.
// - If All can be called multiple times with the correct result.
// - If the result length has the correct size.
// - The fetched results are tested in the same way as in First.
func TestEager_All(t *testing.T) {
	asserts := assert.New(t)

	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	animal := Animal{}
	// Init user model.
	err := animal.Init(&animal)
	asserts.NoError(err)

	// get scope
	s, err := animal.Scope()
	asserts.NoError(err)

	// testing the amount of defined relations.
	asserts.Equal(24, len(s.SQLRelations(orm.Permission{Read: true})))

	// error: result is no ptr to a struct orm.Interface.
	var resAnimals []Animal
	err = animal.All(resAnimals, condition.New().SetWhere("id IN (?)", []int{1, 2}))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrResultPtr, "orm_test.Animal"), err.Error())
	err = animal.All(Animal{}, condition.New().SetWhere("id IN (?)", []int{1, 2}))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrResultPtr, "orm_test.Animal"), err.Error())
	err = animal.All(&Animal{}, condition.New().SetWhere("id IN (?)", []int{1, 2}))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrResultPtr, "orm_test.Animal"), err.Error())
	err = animal.All(&[]string{}, condition.New().SetWhere("id IN (?)", []int{1, 2}))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrResultPtr, "orm_test.Animal"), err.Error())

	// fetch data twice.
	for i := 0; i < 2; i++ {
		var resAnimals []Animal
		err = animal.All(&resAnimals, condition.New().SetWhere("id IN (?)", []int{1, 2}))
		asserts.NoError(err)
		asserts.Equal(2, len(resAnimals))
		for i, animal := range resAnimals {
			test := helperTestCases()[i]
			t.Run("ID:"+strconv.Itoa(test.fetchID)+" run:"+strconv.Itoa(test.fetchID), func(t *testing.T) {
				helperTestResults(asserts, err, test, animal, true)
			})
		}
	}

	// ok: count all rows
	rows, err := animal.Count()
	asserts.NoError(err)
	asserts.Equal(4, rows)

	// ok: count with condition
	rows, err = animal.Count(condition.New().SetWhere("id > 2"))
	asserts.NoError(err)
	asserts.Equal(2, rows)

	// err: condition syntax error to trigger scan error
	rows, err = animal.Count(condition.New().SetWhere("id === 2"))
	asserts.Error(err)
	asserts.Equal(0, rows)

	// err: condition placeholder error to trigger First() error
	rows, err = animal.Count(condition.New().SetWhere("id === 2", 1))
	asserts.Error(err)
	asserts.Equal(0, rows)
}

// testCase is a helper struct for the test cases.
type testCase struct {
	fetchID int
	animal  Animal
	error   bool
}

// helperTestCases will return all testCases.
func helperTestCases() []testCase {
	// run test cases
	b1, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 10:10:10")
	b2, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 10:10:10")
	b3, _ := time.Parse("2006-01-02 15:04:05", "2021-01-03 10:10:10")

	var tests = []testCase{
		{fetchID: 1, animal: Animal{
			Name:      "Blacky",
			SpeciesID: query.NewNullInt(1, true),
			Species:   Species{Name: "Dog"},
			Address:   Address{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)},
			Toys: []Toy{
				{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			},
			Walkers: []Human{
				{Name: "Pat"},
				{Name: "Eva"},
			},
		}},
		{fetchID: 2, animal: Animal{
			Name:      "Snowflake",
			SpeciesID: query.NewNullInt(1, true),
			Species:   Species{Name: "Cat"},
			Address:   Address{Street: query.NewNullString("Brandschenkestrasse", true), Zip: "8002", City: "Zürich", Country: query.NewNullString("Switzerland", true)},
			Toys: []Toy{
				{Name: "Mouse", Brand: query.NewNullString("", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b3, true)},
			}, Walkers: []Human{
				{Name: "Eva"},
			},
		}},
	}
	return tests
}

// helperTestResults checks the given test case with the given orm.Interface.
// the isAll argument is used to change the test cases for HasOne back references. These are not working for All() at the moment.
// all relations and relation field types struct, ptr, slice, ptr-slice, slice-ptr, ptr-slice-ptr are checked.
func helperTestResults(asserts *assert.Assertions, err error, test testCase, animal Animal, isAll bool) {
	if test.error {
		asserts.Error(err)
		asserts.Equal("", err.Error())
	} else {
		asserts.NoError(err)
		// test root fields
		asserts.Equal(test.fetchID, animal.ID)
		asserts.Equal(test.animal.Name, animal.Name)
		// test belongsTo relation struct
		asserts.Equal(test.animal.Species.Name, animal.Species.Name)
		asserts.Equal(test.animal.Species.Name, animal.SpeciesPtr.Name)
		asserts.Equal(test.animal.Species.Name, animal.SpeciesPoly.Name)
		asserts.Equal(test.animal.Species.Name, animal.SpeciesPolyPtr.Name)
		if test.animal.Species.ID != 0 { // ensure id if set in test.
			asserts.Equal(test.animal.Species.ID, animal.Species.ID)
			asserts.Equal(test.animal.Species.ID, animal.SpeciesPtr.ID)
			asserts.Equal(test.animal.Species.ID, animal.SpeciesPoly.ID)
			asserts.Equal(test.animal.Species.ID, animal.SpeciesPolyPtr.ID)
		}
		// test hasOne relation struct
		asserts.Equal(test.animal.Address.Street, animal.Address.Street)
		asserts.Equal(test.animal.Address.Zip, animal.Address.Zip)
		asserts.Equal(test.animal.Address.City, animal.Address.City)
		asserts.Equal(test.animal.Address.Country, animal.Address.Country)
		if test.animal.Address.ID != 0 { // ensure id if set in test.
			asserts.Equal(test.animal.Address.ID, animal.Address.ID)
		}
		if isAll {
			// TODO this will end in a loop because of the validation.v10 package. The orm would handle it correctly.
			//asserts.Equal(0, animal.Address.Animal.ID)
		} else {
			//asserts.Equal(test.fetchID, animal.Address.Animal.ID)
		}
		// test hasOne relation ptr
		asserts.Equal(test.animal.Address.Street, animal.AddressPtr.Street)
		asserts.Equal(test.animal.Address.Zip, animal.AddressPtr.Zip)
		asserts.Equal(test.animal.Address.City, animal.AddressPtr.City)
		asserts.Equal(test.animal.Address.Country, animal.AddressPtr.Country)
		if test.animal.Address.ID != 0 { // ensure id if set in test.
			asserts.Equal(test.animal.Address.ID, animal.AddressPtr.ID)
		}
		if isAll {
			//asserts.Equal(0, animal.AddressPtr.Animal.ID)
		} else {
			//asserts.Equal(test.fetchID, animal.AddressPtr.Animal.ID)
		}
		// test hasOne relation poly
		asserts.Equal(test.animal.Address.Street, animal.AddressPoly.Street)
		asserts.Equal(test.animal.Address.Zip, animal.AddressPoly.Zip)
		asserts.Equal(test.animal.Address.City, animal.AddressPoly.City)
		asserts.Equal(test.animal.Address.Country, animal.AddressPoly.Country)
		if test.animal.Address.ID != 0 { // ensure id if set in test.
			asserts.Equal(test.animal.Address.ID, animal.AddressPoly.ID)
		}
		if isAll {
			//	asserts.Equal(0, animal.AddressPoly.Animal.ID)
		} else {
			//	asserts.Equal(test.fetchID, animal.AddressPoly.Animal.ID)
		}
		// test hasOne relation poly ptr
		asserts.Equal(test.animal.Address.Street, animal.AddressPolyPtr.Street)
		asserts.Equal(test.animal.Address.Zip, animal.AddressPolyPtr.Zip)
		asserts.Equal(test.animal.Address.City, animal.AddressPolyPtr.City)
		asserts.Equal(test.animal.Address.Country, animal.AddressPolyPtr.Country)
		if test.animal.Address.ID != 0 { // ensure id if set in test.
			asserts.Equal(test.animal.Address.ID, animal.AddressPolyPtr.ID)
		}
		if isAll {
			//	asserts.Equal(0, animal.AddressPolyPtr.Animal.ID)
		} else {
			//	asserts.Equal(test.fetchID, animal.AddressPolyPtr.Animal.ID)
		}

		// test hasMany relation struct
		asserts.Equal(len(test.animal.Toys), len(animal.Toys))
		asserts.Equal(len(test.animal.Toys), len(animal.ToysSlicePtr))
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.ToysPtrSlice).Elem().Len())
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Len())
		asserts.Equal(len(test.animal.Toys), len(animal.ToyPoly))
		asserts.Equal(len(test.animal.Toys), len(animal.ToyPolySlicePtr))
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.ToysPtrSlice).Elem().Len())
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Len())
		for i := range test.animal.Toys {
			if test.animal.Toys[i].ID != 0 { // ensure id if set in test.
				asserts.Equal(test.animal.Toys[i].ID, animal.Toys[i].ID)
				asserts.Equal(test.animal.Toys[i].ID, animal.ToysSlicePtr[i].ID)
				asserts.Equal(test.animal.Toys[i].ID, reflect.ValueOf(animal.ToysPtrSlice).Elem().Index(i).FieldByName("ID").Interface())
				asserts.Equal(test.animal.Toys[i].ID, reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Index(i).Elem().FieldByName("ID").Interface())
				asserts.Equal(test.animal.Toys[i].ID, animal.ToyPoly[i].ID)
				asserts.Equal(test.animal.Toys[i].ID, animal.ToyPolySlicePtr[i].ID)
				asserts.Equal(test.animal.Toys[i].ID, reflect.ValueOf(animal.ToyPolyPtrSlice).Elem().Index(i).FieldByName("ID").Interface())
				asserts.Equal(test.animal.Toys[i].ID, reflect.ValueOf(animal.ToyPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("ID").Interface())
			}
			if isAll {
				//asserts.Equal("", animal.Toys[i].AnimalRef.Name) // backref
			} else {
				//asserts.Equal(test.animal.Name, animal.Toys[i].AnimalRef.Name) // backref
			}
			asserts.Equal(test.animal.Toys[i].Name, animal.Toys[i].Name)
			asserts.Equal(test.animal.Toys[i].Name, animal.ToysSlicePtr[i].Name)
			asserts.Equal(test.animal.Toys[i].Name, reflect.ValueOf(animal.ToysPtrSlice).Elem().Index(i).FieldByName("Name").Interface())
			asserts.Equal(test.animal.Toys[i].Name, reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Name").Interface())
			asserts.Equal(test.animal.Toys[i].Name, animal.ToyPoly[i].Name)
			asserts.Equal(test.animal.Toys[i].Name, animal.ToyPolySlicePtr[i].Name)
			asserts.Equal(test.animal.Toys[i].Name, reflect.ValueOf(animal.ToyPolyPtrSlice).Elem().Index(i).FieldByName("Name").Interface())
			asserts.Equal(test.animal.Toys[i].Name, reflect.ValueOf(animal.ToyPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Name").Interface())

			asserts.Equal(test.animal.Toys[i].Brand, animal.Toys[i].Brand)
			asserts.Equal(test.animal.Toys[i].Brand, animal.ToysSlicePtr[i].Brand)
			asserts.Equal(test.animal.Toys[i].Brand, reflect.ValueOf(animal.ToysPtrSlice).Elem().Index(i).FieldByName("Brand").Interface())
			asserts.Equal(test.animal.Toys[i].Brand, reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Brand").Interface())
			asserts.Equal(test.animal.Toys[i].Brand, animal.ToyPoly[i].Brand)
			asserts.Equal(test.animal.Toys[i].Brand, animal.ToyPolySlicePtr[i].Brand)
			asserts.Equal(test.animal.Toys[i].Brand, reflect.ValueOf(animal.ToyPolyPtrSlice).Elem().Index(i).FieldByName("Brand").Interface())
			asserts.Equal(test.animal.Toys[i].Brand, reflect.ValueOf(animal.ToyPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Brand").Interface())

			asserts.Equal(test.animal.Toys[i].DestroyAble, animal.Toys[i].DestroyAble)
			asserts.Equal(test.animal.Toys[i].DestroyAble, animal.ToysSlicePtr[i].DestroyAble)
			asserts.Equal(test.animal.Toys[i].DestroyAble, reflect.ValueOf(animal.ToysPtrSlice).Elem().Index(i).FieldByName("DestroyAble").Interface())
			asserts.Equal(test.animal.Toys[i].DestroyAble, reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Index(i).Elem().FieldByName("DestroyAble").Interface())
			asserts.Equal(test.animal.Toys[i].DestroyAble, animal.ToyPoly[i].DestroyAble)
			asserts.Equal(test.animal.Toys[i].DestroyAble, animal.ToyPolySlicePtr[i].DestroyAble)
			asserts.Equal(test.animal.Toys[i].DestroyAble, reflect.ValueOf(animal.ToyPolyPtrSlice).Elem().Index(i).FieldByName("DestroyAble").Interface())
			asserts.Equal(test.animal.Toys[i].DestroyAble, reflect.ValueOf(animal.ToyPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("DestroyAble").Interface())

			asserts.Equal(test.animal.Toys[i].BoughtAt, animal.Toys[i].BoughtAt)
			asserts.Equal(test.animal.Toys[i].BoughtAt, animal.ToysSlicePtr[i].BoughtAt)
			asserts.Equal(test.animal.Toys[i].BoughtAt, reflect.ValueOf(animal.ToysPtrSlice).Elem().Index(i).FieldByName("BoughtAt").Interface())
			asserts.Equal(test.animal.Toys[i].BoughtAt, reflect.ValueOf(animal.ToysPtrSlicePtr).Elem().Index(i).Elem().FieldByName("BoughtAt").Interface())
			asserts.Equal(test.animal.Toys[i].BoughtAt, animal.ToyPoly[i].BoughtAt)
			asserts.Equal(test.animal.Toys[i].BoughtAt, animal.ToyPolySlicePtr[i].BoughtAt)
			asserts.Equal(test.animal.Toys[i].BoughtAt, reflect.ValueOf(animal.ToyPolyPtrSlice).Elem().Index(i).FieldByName("BoughtAt").Interface())
			asserts.Equal(test.animal.Toys[i].BoughtAt, reflect.ValueOf(animal.ToyPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("BoughtAt").Interface())
		}
		// test m2m relation struct
		asserts.Equal(len(test.animal.Walkers), len(animal.Walkers))
		asserts.Equal(len(test.animal.Walkers), len(animal.WalkersSlicePtr))
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.WalkersPtrSlice).Elem().Len())
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.WalkersPtrSlicePtr).Elem().Len())
		asserts.Equal(len(test.animal.Walkers), len(animal.WalkersPoly))
		asserts.Equal(len(test.animal.Walkers), len(animal.WalkersPolySlicePtr))
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.WalkersPolyPtrSlice).Elem().Len())
		asserts.Equal(len(test.animal.Toys), reflect.ValueOf(animal.WalkersPolyPtrSlicePtr).Elem().Len())
		for i := range test.animal.Walkers {
			if test.animal.Walkers[i].ID != 0 { // ensure id if set in test.
				asserts.Equal(test.animal.Walkers[i].ID, animal.Walkers[i].ID)
				asserts.Equal(test.animal.Walkers[i].ID, animal.WalkersSlicePtr[i].ID)
				asserts.Equal(test.animal.Walkers[i].ID, reflect.ValueOf(animal.WalkersPtrSlice).Elem().Index(i).FieldByName("ID").Interface())
				asserts.Equal(test.animal.Walkers[i].ID, reflect.ValueOf(animal.WalkersPtrSlicePtr).Elem().Index(i).Elem().FieldByName("ID").Interface())
				asserts.Equal(test.animal.Walkers[i].ID, animal.WalkersPoly[i].ID)
				asserts.Equal(test.animal.Walkers[i].ID, animal.WalkersPolySlicePtr[i].ID)
				asserts.Equal(test.animal.Walkers[i].ID, reflect.ValueOf(animal.WalkersPolyPtrSlice).Elem().Index(i).FieldByName("ID").Interface())
				asserts.Equal(test.animal.Walkers[i].ID, reflect.ValueOf(animal.WalkersPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("ID").Interface())
			}
			asserts.Equal(test.animal.Walkers[i].Name, animal.Walkers[i].Name)
			asserts.Equal(test.animal.Walkers[i].Name, animal.WalkersSlicePtr[i].Name)
			asserts.Equal(test.animal.Walkers[i].Name, reflect.ValueOf(animal.WalkersPtrSlice).Elem().Index(i).FieldByName("Name").Interface())
			asserts.Equal(test.animal.Walkers[i].Name, reflect.ValueOf(animal.WalkersPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Name").Interface())
			asserts.Equal(test.animal.Walkers[i].Name, animal.WalkersPoly[i].Name)
			asserts.Equal(test.animal.Walkers[i].Name, animal.WalkersPolySlicePtr[i].Name)
			asserts.Equal(test.animal.Walkers[i].Name, reflect.ValueOf(animal.WalkersPolyPtrSlice).Elem().Index(i).FieldByName("Name").Interface())
			asserts.Equal(test.animal.Walkers[i].Name, reflect.ValueOf(animal.WalkersPolyPtrSlicePtr).Elem().Index(i).Elem().FieldByName("Name").Interface())
		}
	}
}
