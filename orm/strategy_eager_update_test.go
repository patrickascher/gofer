// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"reflect"
	"testing"
	"time"

	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
)

// TestEager_Update_SelfRef tests:
// - If the orm model gets updated correctly on a self referencing model. (child roles added, updated, deleted)
func TestEager_Update_SelfRef(t *testing.T) {
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

	// fetch id 1
	err = role.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(1, role.ID)
	asserts.Equal("RoleA", role.Name)
	asserts.Equal("RoleB", role.Roles[0].Name)
	asserts.Equal("RoleC", role.Roles[0].Roles[0].Name)

	// update root role and child role.
	role.Name = "RoleA-Updated"
	role.Roles = append(role.Roles, Role{Name: "RoleD"}) // add element
	role.Roles[0].Name = "RoleB-Updated"                 // update element
	role.Roles[0].Roles = nil                            // delete element
	err = role.Update()
	asserts.NoError(err)

	// fetch id 1 again and check if the result was saved correctly.
	err = role.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(1, role.ID)
	asserts.Equal("RoleA-Updated", role.Name)
	asserts.Equal(2, len(role.Roles)) // RoleB,RoleD
	asserts.Equal("RoleB-Updated", role.Roles[0].Name)
	asserts.Equal(2, role.Roles[0].ID)         //ensure id did not change
	asserts.Equal(6, role.Roles[1].ID)         // new added role
	asserts.Equal("RoleD", role.Roles[1].Name) // new added role
	asserts.Equal(0, len(role.Roles[0].Roles)) // deleted role
}

// TestEager_Update tests:
// - If all values are getting updated correctly. (ensure id stays the same on relations)
// - If UpdatedAt gets set - if exists.
func TestEager_Update(t *testing.T) {

	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// drop deleted_at column
	_, err := builder.Query().DB().Exec("ALTER TABLE `animals` ADD `updated_at` DATETIME NULL;")
	asserts.NoError(err)

	// delete existing cache because of the saved field (created_at).
	if c.Exist("orm_", "orm_test.Animal") {
		err = c.Delete("orm_", "orm_test.Animal")
		asserts.NoError(err)
	}

	// Test preCache.
	err = orm.PreInit(&Animal{})
	asserts.NoError(err)

	// Init user model.
	animal := Animal{}
	err = animal.Init(&animal)
	asserts.NoError(err)

	// load id 1 - run test case to ensure the AS-IS state.
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	helperTestResults(asserts, err, testCasesUpdate()[0], animal, false)
	asserts.Nil(animal.UpdatedAt) // test updatedAt

	// update
	animal.Name = testCasesUpdate()[1].animal.Name
	animal.Species.Name = testCasesUpdate()[1].animal.Species.Name
	animal.SpeciesPoly.Name = testCasesUpdate()[1].animal.SpeciesPoly.Name
	animal.Address.Street = testCasesUpdate()[1].animal.Address.Street
	animal.AddressPoly.Street = testCasesUpdate()[1].animal.AddressPoly.Street
	animal.Toys[0].Name = testCasesUpdate()[1].animal.Toys[0].Name
	animal.Toys[0].Brand = testCasesUpdate()[1].animal.Toys[0].Brand
	animal.Toys[1].Name = testCasesUpdate()[1].animal.Toys[1].Name
	animal.Toys[1].Brand = testCasesUpdate()[1].animal.Toys[1].Brand
	animal.ToyPoly[0].Name = testCasesUpdate()[1].animal.ToyPoly[0].Name
	animal.ToyPoly[0].Brand = testCasesUpdate()[1].animal.ToyPoly[0].Brand
	animal.ToyPoly[1].Name = testCasesUpdate()[1].animal.ToyPoly[1].Name
	animal.ToyPoly[1].Brand = testCasesUpdate()[1].animal.ToyPoly[1].Brand
	animal.Walkers[0].Name = testCasesUpdate()[1].animal.Walkers[0].Name
	animal.Walkers[1].Name = testCasesUpdate()[1].animal.Walkers[1].Name
	animal.WalkersPoly[0].Name = testCasesUpdate()[1].animal.WalkersPoly[0].Name
	animal.WalkersPoly[1].Name = testCasesUpdate()[1].animal.WalkersPoly[1].Name
	err = animal.Update()
	asserts.NoError(err)

	// load id 1 again - run test case to ensure the TO-BE state.
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	helperTestResults(asserts, err, testCasesUpdate()[1], animal, false)
	asserts.True(animal.UpdatedAt.Valid) // test if updatedAt has a valid timestamp.

	// tear down updated_at test.
	// delete because of the other tests which does not include the updated_at field anymore in the db.
	if c.Exist("orm_", "orm_test.Animal") {
		err = c.Delete("orm_", "orm_test.Animal")
		asserts.NoError(err)
	}
}

// TestEager_Update_HasOneDel tests:
// - If belongsTo gets added, updated and deleted correctly.
// - If belongsTo updates on the Reference if configured so.
// - If hasOne gets added, updated and deleted correctly.
func TestEager_Update_HasOne_BelongsTo(t *testing.T) {

	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// needed because the two species tables are not equal by last autoincrement id because of other test cases.
	_, err := builder2.Query().Delete("species_polies").Where("id = 4").Exec()
	asserts.NoError(err)
	_, err = builder2.Query().DB().Exec("ALTER TABLE species_polies AUTO_INCREMENT=4;")
	asserts.NoError(err)

	// Init user model.
	animal := Animal{}
	err = animal.Init(&animal)
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	helperTestResults(asserts, err, testCasesUpdate()[0], animal, false)
	asserts.Nil(animal.UpdatedAt) // test updatedAt

	// delete hasOne, belongsTo
	animal.Name = testCasesUpdate()[1].animal.Name
	animal.Species = Species{}
	animal.SpeciesPoly = SpeciesPoly{}
	animal.Address = Address{}
	animal.AddressPtr = nil
	animal.AddressPoly = AddressPoly{}
	animal.AddressPolyPtr = nil
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(int64(0), animal.SpeciesID.Int64)
	asserts.Equal(0, animal.Species.ID)
	asserts.Equal((*Species)(nil), animal.SpeciesPtr)
	asserts.Equal(0, animal.SpeciesPoly.ID)
	asserts.Equal((*SpeciesPoly)(nil), animal.SpeciesPolyPtr)
	asserts.Equal(0, animal.Address.ID)
	asserts.Equal((*Address)(nil), animal.AddressPtr)
	asserts.Equal(0, animal.AddressPoly.ID)
	asserts.Equal((*AddressPoly)(nil), animal.AddressPolyPtr)

	// create new entries - without ID
	animal.SpeciesPtr = &Species{Name: "NewSpecies"}
	animal.SpeciesPolyPtr = &SpeciesPoly{Name: "NewSpeciesPoly"}
	animal.AddressPtr = &Address{Street: query.NewNullString("NewStreet", true)}
	animal.AddressPolyPtr = &AddressPoly{Street: query.NewNullString("NewStreetPoly", true)}
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(4, animal.Species.ID)
	asserts.Equal("NewSpecies", animal.Species.Name)
	asserts.Equal(4, animal.SpeciesPtr.ID)
	asserts.Equal("NewSpecies", animal.SpeciesPtr.Name)
	asserts.Equal(4, animal.SpeciesPoly.ID)
	asserts.Equal("NewSpeciesPoly", animal.SpeciesPoly.Name)
	asserts.Equal(4, animal.SpeciesPolyPtr.ID)
	asserts.Equal("NewSpeciesPoly", animal.SpeciesPolyPtr.Name)
	asserts.Equal(3, animal.Address.ID)
	asserts.Equal("NewStreet", animal.Address.Street.String)
	asserts.Equal(3, animal.AddressPtr.ID)
	asserts.Equal("NewStreet", animal.AddressPtr.Street.String)
	asserts.Equal(4, animal.AddressPoly.ID)
	asserts.Equal("NewStreetPoly", animal.AddressPoly.Street.String)
	asserts.Equal(4, animal.AddressPolyPtr.ID)
	asserts.Equal("NewStreetPoly", animal.AddressPolyPtr.Street.String)

	// create new entries - with ID
	animal.SpeciesPtr = &Species{Base: Base{ID: 10}, Name: "NewSpecies"}
	animal.SpeciesPoly = SpeciesPoly{ID: 10, Name: "NewSpeciesPoly"}
	animal.AddressPtr = &Address{Base: Base{ID: 10}, Street: query.NewNullString("NewStreet", true)}
	animal.AddressPolyPtr = &AddressPoly{ID: 10, Street: query.NewNullString("NewStreetPoly", true)}
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(10, animal.Species.ID)
	asserts.Equal("NewSpecies", animal.Species.Name)
	asserts.Equal(10, animal.SpeciesPtr.ID)
	asserts.Equal("NewSpecies", animal.SpeciesPtr.Name)
	asserts.Equal(10, animal.SpeciesPoly.ID)
	asserts.Equal("NewSpeciesPoly", animal.SpeciesPoly.Name)
	asserts.Equal(10, animal.SpeciesPolyPtr.ID)
	asserts.Equal("NewSpeciesPoly", animal.SpeciesPolyPtr.Name)
	asserts.Equal(10, animal.Address.ID)
	asserts.Equal("NewStreet", animal.Address.Street.String)
	asserts.Equal(10, animal.AddressPtr.ID)
	asserts.Equal("NewStreet", animal.AddressPtr.Street.String)
	asserts.Equal(10, animal.AddressPoly.ID)
	asserts.Equal("NewStreetPoly", animal.AddressPoly.Street.String)
	asserts.Equal(10, animal.AddressPolyPtr.ID)
	asserts.Equal("NewStreetPoly", animal.AddressPolyPtr.Street.String)

	// update entries
	animal.SpeciesPtr.Name = "updated-NewSpecies"
	animal.SpeciesPoly.Name = "updated-NewSpeciesPoly"
	animal.AddressPtr.Street.String = "updated-NewStreet"
	animal.AddressPolyPtr.Street.String = "updated-NewStreetPoly"
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(10, animal.Species.ID)
	asserts.Equal("updated-NewSpecies", animal.Species.Name)
	asserts.Equal(10, animal.SpeciesPtr.ID)
	asserts.Equal("updated-NewSpecies", animal.SpeciesPtr.Name)
	asserts.Equal(10, animal.SpeciesPoly.ID)
	asserts.Equal("updated-NewSpeciesPoly", animal.SpeciesPoly.Name)
	asserts.Equal(10, animal.SpeciesPolyPtr.ID)
	asserts.Equal("updated-NewSpeciesPoly", animal.SpeciesPolyPtr.Name)
	asserts.Equal(10, animal.Address.ID)
	asserts.Equal("updated-NewStreet", animal.Address.Street.String)
	asserts.Equal(10, animal.AddressPtr.ID)
	asserts.Equal("updated-NewStreet", animal.AddressPtr.Street.String)
	asserts.Equal(10, animal.AddressPoly.ID)
	asserts.Equal("updated-NewStreetPoly", animal.AddressPoly.Street.String)
	asserts.Equal(10, animal.AddressPolyPtr.ID)
	asserts.Equal("updated-NewStreetPoly", animal.AddressPolyPtr.Street.String)

	// update entries with config REF only. Values should stay the same.
	animal.SpeciesPtr.Name = "updatedRef-NewSpecies"
	animal.SpeciesPoly.Name = "updatedRef-NewSpeciesPoly"
	scope, err := animal.Scope()
	asserts.NoError(err)
	scope.SetConfig(orm.NewConfig().SetUpdateReferenceOnly(true))
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(10, animal.Species.ID)
	asserts.Equal("updated-NewSpecies", animal.Species.Name)
	asserts.Equal(10, animal.SpeciesPtr.ID)
	asserts.Equal("updated-NewSpecies", animal.SpeciesPtr.Name)
	asserts.Equal(10, animal.SpeciesPoly.ID)
	asserts.Equal("updated-NewSpeciesPoly", animal.SpeciesPoly.Name)
	asserts.Equal(10, animal.SpeciesPolyPtr.ID)
	asserts.Equal("updated-NewSpeciesPoly", animal.SpeciesPolyPtr.Name)
}

// TestEager_Update_HasMany_M2M tests:
// - If hasMany gets added, updated and deleted correctly.
// - If m2m gets added, updated and deleted correctly.
// - If m2m updates reference only when configured.
func TestEager_Update_HasMany_M2M(t *testing.T) {

	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)

	// Init user model.
	animal := Animal{}
	err := animal.Init(&animal)
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	helperTestResults(asserts, err, testCasesUpdate()[0], animal, false)

	// delete hasOne, belongsTo
	animal.Toys = nil
	animal.ToysSlicePtr = nil
	animal.ToysPtrSlice = nil
	animal.ToysPtrSlicePtr = nil
	animal.ToyPoly = nil
	animal.ToyPolySlicePtr = nil
	animal.ToyPolyPtrSlice = nil
	animal.ToyPolyPtrSlicePtr = nil
	animal.Walkers = nil
	animal.WalkersSlicePtr = nil
	animal.WalkersPtrSlice = nil
	animal.WalkersPtrSlicePtr = nil
	animal.WalkersPoly = nil
	animal.WalkersPolySlicePtr = nil
	animal.WalkersPolyPtrSlice = nil
	animal.WalkersPolyPtrSlicePtr = nil
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(0, len(animal.Toys))
	asserts.Equal(0, len(animal.ToysSlicePtr))
	asserts.True(reflect.ValueOf(animal.ToysPtrSlice).IsZero())
	asserts.True(reflect.ValueOf(animal.ToysPtrSlicePtr).IsZero())
	asserts.Equal(0, len(animal.ToyPoly))
	asserts.Equal(0, len(animal.ToyPolySlicePtr))
	asserts.True(reflect.ValueOf(animal.ToyPolyPtrSlice).IsZero())
	asserts.True(reflect.ValueOf(animal.ToyPolyPtrSlicePtr).IsZero())
	asserts.Equal(0, len(animal.Walkers))
	asserts.Equal(0, len(animal.WalkersSlicePtr))
	asserts.True(reflect.ValueOf(animal.WalkersPtrSlice).IsZero())
	asserts.True(reflect.ValueOf(animal.WalkersPtrSlicePtr).IsZero())
	asserts.Equal(0, len(animal.WalkersPoly))
	asserts.Equal(0, len(animal.WalkersPolySlicePtr))
	asserts.True(reflect.ValueOf(animal.WalkersPolyPtrSlice).IsZero())
	asserts.True(reflect.ValueOf(animal.WalkersPolyPtrSlicePtr).IsZero())

	// create new entries - without ID
	animal.ToysPtrSlicePtr = &[]*Toy{
		{Name: "NewBone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true)},
		{Name: "NewKong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true)},
	}
	animal.ToyPolyPtrSlicePtr = &[]*ToyPoly{
		{Name: "NewBone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true)},
		{Name: "NewKong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true)},
	}
	animal.WalkersPtrSlicePtr = &[]*Human{{Name: "NewPat"}, {Name: "NewEva"}}
	animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Name: "NewPat"}, {Name: "NewEva"}}
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(2, len(animal.Toys))
	asserts.Equal("NewBone", animal.Toys[0].Name)
	asserts.Equal("NewKong", animal.Toys[1].Name)
	asserts.Equal(2, len(animal.ToyPoly))
	asserts.Equal("NewBone", animal.ToyPoly[0].Name)
	asserts.Equal("NewKong", animal.ToyPoly[1].Name)
	asserts.Equal(2, len(animal.Walkers))
	asserts.Equal("NewPat", animal.Walkers[0].Name)
	asserts.Equal("NewEva", animal.Walkers[1].Name)
	asserts.Equal(2, len(animal.WalkersPoly))
	asserts.Equal("NewPat", animal.WalkersPoly[0].Name)
	asserts.Equal("NewEva", animal.WalkersPoly[1].Name)

	// create new entries - with ID
	animal.ToysPtrSlicePtr = &[]*Toy{
		{Base: Base{ID: 10}, Name: "NewBone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true)},
		{Base: Base{ID: 11}, Name: "NewKong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true)},
	}
	animal.ToyPolyPtrSlicePtr = &[]*ToyPoly{
		{ID: 10, Name: "NewBone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true)},
		{ID: 11, Name: "NewKong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true)},
	}
	animal.WalkersPtrSlicePtr = &[]*Human{{Base: Base{ID: 10}, Name: "NewPat"}, {Base: Base{ID: 11}, Name: "NewEva"}}
	animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Base: Base{ID: 10}, Name: "NewPat"}, {Base: Base{ID: 11}, Name: "NewEva"}}
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(2, len(animal.Toys))
	asserts.Equal(10, animal.Toys[0].ID)
	asserts.Equal("NewBone", animal.Toys[0].Name)
	asserts.Equal(11, animal.Toys[1].ID)
	asserts.Equal("NewKong", animal.Toys[1].Name)
	asserts.Equal(2, len(animal.ToyPoly))
	asserts.Equal(10, animal.ToyPoly[0].ID)
	asserts.Equal("NewBone", animal.ToyPoly[0].Name)
	asserts.Equal(11, animal.ToyPoly[1].ID)
	asserts.Equal("NewKong", animal.ToyPoly[1].Name)
	asserts.Equal(2, len(animal.Walkers))
	asserts.Equal(10, animal.Walkers[0].ID)
	asserts.Equal("NewPat", animal.Walkers[0].Name)
	asserts.Equal(11, animal.Walkers[1].ID)
	asserts.Equal("NewEva", animal.Walkers[1].Name)
	asserts.Equal(2, len(animal.WalkersPoly))
	asserts.Equal(10, animal.WalkersPoly[0].ID)
	asserts.Equal("NewPat", animal.WalkersPoly[0].Name)
	asserts.Equal(11, animal.WalkersPoly[1].ID)
	asserts.Equal("NewEva", animal.WalkersPoly[1].Name)

	// update entries
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.ToysPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedBone"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.ToysPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedKong"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.ToyPolyPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedBone"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.ToyPolyPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedKong"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedPat"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedEva"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPolyPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedPat"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPolyPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedEva"))
	err = animal.Update()
	asserts.NoError(err)
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(2, len(animal.Toys))
	asserts.Equal(10, animal.Toys[0].ID)
	asserts.Equal("updatedBone", animal.Toys[0].Name)
	asserts.Equal(11, animal.Toys[1].ID)
	asserts.Equal("updatedKong", animal.Toys[1].Name)
	asserts.Equal(2, len(animal.ToyPoly))
	asserts.Equal(10, animal.ToyPoly[0].ID)
	asserts.Equal("updatedBone", animal.ToyPoly[0].Name)
	asserts.Equal(11, animal.ToyPoly[1].ID)
	asserts.Equal("updatedKong", animal.ToyPoly[1].Name)
	asserts.Equal(2, len(animal.Walkers))
	asserts.Equal(10, animal.Walkers[0].ID)
	asserts.Equal("updatedPat", animal.Walkers[0].Name)
	asserts.Equal(11, animal.Walkers[1].ID)
	asserts.Equal("updatedEva", animal.Walkers[1].Name)
	asserts.Equal(2, len(animal.WalkersPoly))
	asserts.Equal(10, animal.WalkersPoly[0].ID)
	asserts.Equal("updatedPat", animal.WalkersPoly[0].Name)
	asserts.Equal(11, animal.WalkersPoly[1].ID)
	asserts.Equal("updatedEva", animal.WalkersPoly[1].Name)

	// update entries with ref only configured - value should stay the same.
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedRefPat"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedRefEva"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPolyPtrSlicePtr)).Index(0)).FieldByName("Name").Set(reflect.ValueOf("updatedRefPat"))
	reflect.Indirect(reflect.Indirect(reflect.ValueOf(animal.WalkersPolyPtrSlicePtr)).Index(1)).FieldByName("Name").Set(reflect.ValueOf("updatedRefEva"))
	scope, err := animal.Scope()
	asserts.NoError(err)
	scope.SetConfig(orm.NewConfig().SetUpdateReferenceOnly(true))
	err = animal.Update()
	asserts.NoError(err)
	// load id 1 again - run test case to ensure the TO-BE state.
	err = animal.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(2, len(animal.Walkers))
	asserts.Equal(10, animal.Walkers[0].ID)
	asserts.Equal("updatedPat", animal.Walkers[0].Name)
	asserts.Equal(11, animal.Walkers[1].ID)
	asserts.Equal("updatedEva", animal.Walkers[1].Name)
	asserts.Equal(2, len(animal.WalkersPoly))
	asserts.Equal(10, animal.WalkersPoly[0].ID)
	asserts.Equal("updatedPat", animal.WalkersPoly[0].Name)
	asserts.Equal(11, animal.WalkersPoly[1].ID)
	asserts.Equal("updatedEva", animal.WalkersPoly[1].Name)
}

// testCasesUpdate is a helper to check the result.
// test cases: 1 = db before update, 2 = how it should look after the update
func testCasesUpdate() []testCase {
	b1, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 10:10:10")
	b2, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 10:10:10")

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
		{fetchID: 1, animal: Animal{
			Name:        "update-Blacky",
			SpeciesID:   query.NewNullInt(1, true),
			Species:     Species{Base: Base{ID: 1}, Name: "update-Dog"},
			SpeciesPoly: SpeciesPoly{ID: 1, Name: "update-Dog"},
			Address:     Address{Base: Base{ID: 1}, Street: query.NewNullString("update-Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)},
			AddressPoly: AddressPoly{ID: 1, Street: query.NewNullString("update-Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)},
			Toys: []Toy{
				{Base: Base{ID: 1}, Name: "update-Bone", Brand: query.NewNullString("update-Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				{Base: Base{ID: 2}, Name: "update-Kong", Brand: query.NewNullString("update-Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			},
			ToyPoly: []ToyPoly{
				{ID: 1, Name: "update-Bone", Brand: query.NewNullString("update-Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				{ID: 2, Name: "update-Kong", Brand: query.NewNullString("update-Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			},
			Walkers: []Human{
				{Base: Base{ID: 1}, Name: "update-Pat"},
				{Base: Base{ID: 2}, Name: "update-Eva"},
			},
			WalkersPoly: []HumanPoly{
				{Base: Base{ID: 1}, Name: "update-Pat"},
				{Base: Base{ID: 2}, Name: "update-Eva"},
			},
		}},
	}
	return tests
}
