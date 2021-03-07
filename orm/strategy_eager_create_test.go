// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
)

// TestEager_Create_SelfRef tests:
// - If the self referencing struct is added correctly.
func TestEager_Create_SelfRef(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)

	// Init user model.
	role := Role{}
	err := role.Init(&role)
	asserts.NoError(err)

	// Set roles.
	role.Name = "RoleA"
	role.Roles = append(role.Roles, Role{Name: "RoleB", Roles: []Role{{Name: "RoleC"}}})
	err = role.Create()
	asserts.NoError(err)

	// check if the lastID was added.
	asserts.Equal(1, role.ID)
	asserts.Equal("RoleA", role.Name)
	asserts.Equal(1, len(role.Roles))
	asserts.Equal("RoleB", role.Roles[0].Name)
	asserts.Equal(1, len(role.Roles[0].Roles))
	asserts.Equal("RoleC", role.Roles[0].Roles[0].Name)

	// Testing a update on Role-B and setting a new Role-D. Role-C should be deleted.
	// TODO on m2m only the junction entries are deleted, not the entries behind. To so if there is no reference?
	// Init user model.
	role = Role{}
	err = role.Init(&role)
	asserts.NoError(err)

	// Set roles.
	role.Name = "RoleA"
	role.Roles = append(role.Roles, Role{Base: Base{ID: 2}, Name: "RoleB-updated", Roles: []Role{{Name: "RoleD"}}})
	err = role.Create()
	asserts.NoError(err)

	// check if the lastID was added.
	asserts.Equal(4, role.ID)
	asserts.Equal("RoleA", role.Name)
	asserts.Equal(1, len(role.Roles))
	asserts.Equal("RoleB-updated", role.Roles[0].Name)
	asserts.Equal(1, len(role.Roles[0].Roles))
	asserts.Equal("RoleD", role.Roles[0].Roles[0].Name)
}

// TestEager_Create tests:
// - 0-3: tests all struct, ptr, slice, ptr slice values on hasOne, belongsTo, hasMany and m2m relations (+poly).
// - 4	: belongsTo, m2m values are changed with an existing ID. The reference value should get updated.
// - 5	: belongsTo, m2m values are set with a primary key but the primary key does not exist yet. The reference value should get created.
// - 6	: belongsTo, m2m values are set but only the reference (junction) entry should get created.
// - 7  : test hasOne, belongsTo, hasMany and m2m with a nil as value. No entry/reference should get created.
// - 8 	: test hasOne, belongsTo, hasMany and m2m with an empty value. No entry/reference should get created.
func TestEager_Create(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)

	// drop deleted_at column
	_, err := builder.Query().DB().Exec("ALTER TABLE `animals` ADD `created_at` DATETIME NULL;")
	asserts.NoError(err)

	// delete existing cache because of the saved field (created_at).
	if c.Exist("orm_", "orm_test.Animal") {
		err = c.Delete("orm_", "orm_test.Animal")
		asserts.NoError(err)
	}

	// Run test cases
	for i := 0; i <= 8; i++ {

		// Init user model.
		animal := Animal{}
		err = animal.Init(&animal)
		asserts.NoError(err)

		animal.Name = "Blacky"
		b1, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 10:10:10")
		b2, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 10:10:10")

		// test cases - differents are: the struct, struct-poly, ptr, poly-ptr values.
		switch i {
		case 0:
			animal.Species = Species{Name: "Dog"}
			animal.SpeciesPoly = SpeciesPoly{Name: "Dog"}
			//TODO enable after validation.v10 bug			animal.Address = Address{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.Address = Address{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			// 			animal.AddressPoly = AddressPoly{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.AddressPoly = AddressPoly{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.Toys = append(animal.Toys,
				Toy{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true), AnimalRef: &animal},
				Toy{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true), AnimalRef: &animal},
			)
			animal.ToyPoly = append(animal.ToyPoly,
				ToyPoly{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				ToyPoly{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			)
			animal.Walkers = append(animal.Walkers, Human{Name: "Pat"}, Human{Name: "Eva"})
			animal.WalkersPoly = append(animal.WalkersPoly, HumanPoly{Name: "Pat"}, HumanPoly{Name: "Eva"})
		case 1:
			animal.SpeciesPtr = &Species{Name: "Dog"}
			animal.SpeciesPolyPtr = &SpeciesPoly{Name: "Dog"}
			//TODO enable after validation.v10 bug 			animal.AddressPtr = &Address{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.AddressPtr = &Address{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			//			animal.AddressPolyPtr = &AddressPoly{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.AddressPolyPtr = &AddressPoly{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.ToysSlicePtr = append(animal.ToysSlicePtr,
				&Toy{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true), AnimalRef: &animal},
				&Toy{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true), AnimalRef: &animal},
			)
			animal.ToyPolySlicePtr = append(animal.ToyPolySlicePtr,
				&ToyPoly{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				&ToyPoly{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			)
			animal.WalkersSlicePtr = append(animal.WalkersSlicePtr, &Human{Name: "Pat"}, &Human{Name: "Eva"})
			animal.WalkersPolySlicePtr = append(animal.WalkersPolySlicePtr, &HumanPoly{Name: "Pat"}, &HumanPoly{Name: "Eva"})
		case 2:
			animal.Species = Species{Name: "Dog"}
			animal.SpeciesPoly = SpeciesPoly{Name: "Dog"}
			//TODO enable after validation.v10 bug 				animal.Address = Address{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.Address = Address{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			//			animal.AddressPoly = AddressPoly{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.AddressPoly = AddressPoly{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.ToysPtrSlice = &[]Toy{
				{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true), AnimalRef: &animal},
				{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true), AnimalRef: &animal},
			}
			animal.ToyPolyPtrSlice = &[]ToyPoly{
				{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			}
			animal.WalkersPtrSlice = &[]Human{{Name: "Pat"}, {Name: "Eva"}}
			animal.WalkersPolyPtrSlice = &[]HumanPoly{{Name: "Pat"}, {Name: "Eva"}}
		default:
			animal.Species = Species{Name: "Dog"}
			animal.SpeciesPoly = SpeciesPoly{Name: "Dog"}
			//TODO enable after validation.v10 bug			animal.Address = Address{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.Address = Address{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			//			animal.AddressPoly = AddressPoly{Animal: &animal, Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.AddressPoly = AddressPoly{Street: query.NewNullString("Erika-Mann-Str. 33", true), Zip: "80636", City: "Munich", Country: query.NewNullString("Germany", true)}
			animal.ToysPtrSlicePtr = &[]*Toy{
				{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true), AnimalRef: &animal},
				{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true), AnimalRef: &animal},
			}
			animal.ToyPolyPtrSlicePtr = &[]*ToyPoly{
				{Name: "Bone", Brand: query.NewNullString("Trixie", true), DestroyAble: query.NewNullBool(true, true), BoughtAt: query.NewNullTime(b1, true)},
				{Name: "Kong", Brand: query.NewNullString("Kong", true), DestroyAble: query.NewNullBool(false, true), BoughtAt: query.NewNullTime(b2, true)},
			}
			animal.WalkersPtrSlicePtr = &[]*Human{{Name: "Pat"}, {Name: "Eva"}}
			animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Name: "Pat"}, {Name: "Eva"}}
		}

		// run tests
		switch i {
		case 0, 1, 2, 3:
			err = animal.Create()
			asserts.NoError(err)

			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			test := helperTestCases()[0]
			test.fetchID = i + 1 // manipulate fetchID, because its fixed in the test as 1.
			helperTestResults(asserts, err, test, animal, false)

			// Delete entry
			err = animal.Delete()
			asserts.NoError(err)

			// error because it's soft deleted.
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.Error(err)
			asserts.Equal(sql.ErrNoRows, err)

			// ok because soft delete is included.
			scope, err := animal.Scope()
			asserts.NoError(err)
			scope.SetConfig(orm.NewConfig().SetShowDeletedRows(true))
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			helperTestResults(asserts, err, test, animal, false)
		case 4:
			// change belongsTo to an existing ID - value should be updated
			animal.Species.ID = 1
			animal.Species.Name = "Dog-updated"
			animal.SpeciesPoly.ID = 1
			animal.SpeciesPoly.Name = "Dog-updated"
			// change m2m to an existing ID - value should be updated instead of created.
			animal.WalkersPtrSlicePtr = &[]*Human{{Base: Base{ID: 1}, Name: "Pat-updated"}, {Base: Base{ID: 2}, Name: "Eva-updated"}}
			animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Base: Base{ID: 1}, Name: "Pat-updated"}, {Base: Base{ID: 2}, Name: "Eva-updated"}}

			// create entry
			err = animal.Create()
			asserts.NoError(err)

			// check result
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			asserts.Equal(1, animal.Species.ID)
			asserts.Equal("Dog-updated", animal.Species.Name)
			asserts.Equal(1, animal.Walkers[0].ID)
			asserts.Equal("Pat-updated", animal.Walkers[0].Name)
			asserts.Equal(2, animal.Walkers[1].ID)
			asserts.Equal("Eva-updated", animal.Walkers[1].Name)
			asserts.Equal(1, animal.WalkersPoly[0].ID)
			asserts.Equal("Pat-updated", animal.WalkersPoly[0].Name)
			asserts.Equal(2, animal.WalkersPoly[1].ID)
			asserts.Equal("Eva-updated", animal.WalkersPoly[1].Name)

		case 5:
			// testing none existing relation with primary key set.
			animal.Species.ID = 10
			animal.Species.Name = "Dog-created"
			animal.SpeciesPoly.ID = 10
			animal.SpeciesPoly.Name = "Dog-created"
			animal.WalkersPtrSlicePtr = &[]*Human{{Base: Base{ID: 10}, Name: "Pat"}, {Base: Base{ID: 20}, Name: "Eva"}}
			animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Base: Base{ID: 10}, Name: "Pat"}, {Base: Base{ID: 20}, Name: "Eva"}}

			// create entry
			err = animal.Create()
			asserts.NoError(err)

			// check result
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			asserts.Equal(10, animal.Species.ID)
			asserts.Equal("Dog-created", animal.Species.Name)
			asserts.Equal(10, animal.SpeciesPoly.ID)
			asserts.Equal("Dog-created", animal.SpeciesPoly.Name)
			asserts.Equal(10, animal.Walkers[0].ID)
			asserts.Equal("Pat", animal.Walkers[0].Name)
			asserts.Equal(20, animal.Walkers[1].ID)
			asserts.Equal("Eva", animal.Walkers[1].Name)
			asserts.Equal(10, animal.WalkersPoly[0].ID)
			asserts.Equal("Pat", animal.WalkersPoly[0].Name)
			asserts.Equal(20, animal.WalkersPoly[1].ID)
			asserts.Equal("Eva", animal.WalkersPoly[1].Name)
		case 6:
			// only update refs.
			s, err := animal.Scope()
			asserts.NoError(err)
			s.SetConfig(orm.NewConfig().SetUpdateReferenceOnly(true))
			animal.Species.ID = 2
			animal.Species.Name = "Dog-updated"
			animal.SpeciesPoly.ID = 2
			animal.SpeciesPoly.Name = "Dog-updated"
			animal.WalkersPtrSlicePtr = &[]*Human{{Base: Base{ID: 3}, Name: "Pat-updated"}, {Base: Base{ID: 4}, Name: "Eva-updated"}}
			animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{Base: Base{ID: 3}, Name: "Pat-updated"}, {Base: Base{ID: 4}, Name: "Eva-updated"}}

			// create entry
			err = animal.Create()
			asserts.NoError(err)

			// check result
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			asserts.Equal(2, animal.Species.ID)
			asserts.Equal("Dog", animal.Species.Name)
			asserts.Equal(2, animal.SpeciesPoly.ID)
			asserts.Equal("Dog", animal.SpeciesPoly.Name)
			asserts.Equal(3, animal.Walkers[0].ID)
			asserts.Equal("Pat", animal.Walkers[0].Name)
			asserts.Equal(4, animal.Walkers[1].ID)
			asserts.Equal("Eva", animal.Walkers[1].Name)
			asserts.Equal(3, animal.WalkersPoly[0].ID)
			asserts.Equal("Pat", animal.WalkersPoly[0].Name)
			asserts.Equal(4, animal.WalkersPoly[1].ID)
			asserts.Equal("Eva", animal.WalkersPoly[1].Name)
		case 7:
			// test hasMany,m2m nil
			animal.Species = Species{}
			animal.SpeciesPtr = nil
			animal.SpeciesPoly = SpeciesPoly{}
			animal.SpeciesPolyPtr = nil
			animal.Address = Address{}
			animal.AddressPtr = nil
			animal.AddressPoly = AddressPoly{}
			animal.AddressPolyPtr = nil
			animal.Toys = nil
			animal.ToysPtrSlice = nil
			animal.ToysSlicePtr = nil
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

			// create entry
			err = animal.Create()
			asserts.NoError(err)

			// check result
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			asserts.Equal(Species{}, animal.Species)
			asserts.Nil(animal.SpeciesPtr)
			asserts.Equal(SpeciesPoly{}, animal.SpeciesPoly)
			asserts.Nil(animal.SpeciesPolyPtr)
			asserts.Equal(Address{}, animal.Address)
			asserts.Nil(animal.AddressPtr)
			asserts.Equal(AddressPoly{}, animal.AddressPoly)
			asserts.Nil(animal.AddressPolyPtr)
			asserts.Equal(0, len(animal.Toys))
			asserts.Equal(0, len(animal.ToyPoly))
			asserts.Equal(0, len(animal.Walkers))
			asserts.Equal(0, len(animal.WalkersPoly))
		case 8:
			// test hasMany,m2m with zero value.
			animal.Species = Species{}
			animal.SpeciesPtr = &Species{}
			animal.SpeciesPoly = SpeciesPoly{}
			animal.SpeciesPolyPtr = &SpeciesPoly{}
			animal.Address = Address{}
			animal.AddressPtr = &Address{}
			animal.AddressPoly = AddressPoly{}
			animal.AddressPolyPtr = &AddressPoly{}
			animal.ToysPtrSlicePtr = &[]*Toy{{}}
			animal.ToyPolyPtrSlicePtr = &[]*ToyPoly{{}}
			animal.WalkersPtrSlicePtr = &[]*Human{{}}
			animal.WalkersPolyPtrSlicePtr = &[]*HumanPoly{{}}

			// create entry
			err = animal.Create()
			asserts.NoError(err)

			// check result
			err = animal.First(condition.New().SetWhere("id = ?", animal.ID))
			asserts.NoError(err)
			asserts.Equal(Species{}, animal.Species)
			asserts.Nil(animal.SpeciesPtr)
			asserts.Equal(SpeciesPoly{}, animal.SpeciesPoly)
			asserts.Nil(animal.SpeciesPolyPtr)
			asserts.Equal(Address{}, animal.Address)
			asserts.Nil(animal.AddressPtr)
			asserts.Equal(AddressPoly{}, animal.AddressPoly)
			asserts.Nil(animal.AddressPolyPtr)
			asserts.Equal(0, len(animal.Toys))
			asserts.Equal(0, len(animal.ToyPoly))
			asserts.Equal(0, len(animal.Walkers))
			asserts.Equal(0, len(animal.WalkersPoly))
		}

		// ensure created_at has a valid timestamp.
		asserts.True(animal.CreatedAt.Valid)
	}

	// tear down created_at test.
	// delete because of the other tests which does not include the created_at field anymore in the db.
	if c.Exist("orm_", "orm_test.Animal") {
		err = c.Delete("orm_", "orm_test.Animal")
		asserts.NoError(err)
	}
}
