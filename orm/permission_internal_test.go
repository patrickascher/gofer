// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestNewWBList testing if the given policy and fields are added correctly
func TestNewWBList(t *testing.T) {
	asserts := assert.New(t)

	// whitelist factory
	w := newPermissionList(WHITELIST, []string{"ID", "Name"})
	asserts.Equal(WHITELIST, w.policy)
	asserts.Equal([]string{"ID", "Name"}, w.fields)

	// blacklist factory
	b := newPermissionList(BLACKLIST, []string{"ID", "Name"})
	asserts.Equal(BLACKLIST, b.policy)
	asserts.Equal([]string{"ID", "Name"}, b.fields)
}

// Test_setFieldPermission tests:
// - Whitelist/Blacklist: If additional fields get deleted if the whole relation is added.
func Test_setFieldPermission(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	createBuilderCache(asserts)

	c := Animal{}
	err := c.Init(&c)
	if asserts.NoError(err) {

		// testing if wb list is still nil
		err = c.scope.setFieldPermission()
		if asserts.NoError(err) {
			asserts.Nil(c.permissionList)

			// user added whitelist
			c.SetPermissions(WHITELIST, "Name", "Species.Name", "Address", "Address.Street", "Address.Zip", "does not exist")
			err = c.scope.setFieldPermission()
			if asserts.NoError(err) {
				// Address.Street and Address.Zip got removed because the whole relation was added.
				asserts.Equal(&permissionList{policy: 1, fields: []string{"Name", "Species.Name", "Address", "does not exist", "ID", "DeletedAt", "Species.ID", "SpeciesID"}}, c.permissionList)
				fields := c.scope.SQLFields(Permission{Read: true})
				asserts.Equal(4, len(fields)) // Name,ID,SpeciesID,DeletedAt
				relations := c.scope.SQLRelations(Permission{Read: true})
				asserts.Equal(2, len(relations)) // Species, Address
			}

			// user added blacklist
			c.SetPermissions(BLACKLIST, "Name", "Species.Name", "Address", "Address.Street", "Address.Zip")
			err = c.scope.setFieldPermission()
			if asserts.NoError(err) {
				// Other fields were removed because they are mandatory
				asserts.Equal(&permissionList{policy: 0, fields: []string{"Name", "Species.Name", "Address"}}, c.permissionList)
				fields := c.scope.SQLFields(Permission{Read: true})
				asserts.Equal(3, len(fields)) // ID,SpeciesID,DeletedAt
				relations := c.scope.SQLRelations(Permission{Read: true})
				asserts.Equal(23, len(relations))
			}
		}
	}
}

// TestModel_SetWBList_Unique: testing if double entered keys will be unique in the wb list.
func TestModel_SetWBList_Unique(t *testing.T) {
	asserts := assert.New(t)
	createBuilderCache(asserts)
	helperCreateDatabaseAndTable(asserts)

	c := &Animal{}
	err := c.Init(c)
	asserts.NoError(err)

	c.SetPermissions(WHITELIST, "ID", "ID", "SpeciesPoly.Name", "Species.Name")
	asserts.NoError(err)

	asserts.Equal(WHITELIST, c.permissionList.policy)
	asserts.Equal([]string{"ID", "SpeciesPoly.Name", "Species.Name"}, c.permissionList.fields)

	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"ID", "SpeciesPoly.Name", "Species.Name", "DeletedAt", "Species.ID", "SpeciesID", "SpeciesPoly.ID", "SpeciesPoly.SpeciesPolyType"}, c.permissionList.fields)
}

// Test_addMandatoryFields testing if all mandatory fields were added.
func Test_addMandatoryFields(t *testing.T) {
	asserts := assert.New(t)
	createBuilderCache(asserts)
	helperCreateDatabaseAndTable(asserts)

	c := &Animal{}
	err := c.Init(c)
	asserts.NoError(err)
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Nil(c.permissionList)

	// all fk,afk,poly and primary keys must be loaded that the relation relevant data is given.
	c.SetPermissions(WHITELIST, "SpeciesPoly.Name")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"SpeciesPoly.Name", "ID", "DeletedAt", "SpeciesPoly.ID", "SpeciesID", "SpeciesPoly.SpeciesPolyType"}, c.permissionList.fields)

	// full relation name
	c.SetPermissions(WHITELIST, "Species")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"Species", "ID", "DeletedAt", "SpeciesID"}, c.permissionList.fields)

	// child relation, self-ref
	c.SetPermissions(WHITELIST, "Toys.AnimalRef.Name")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"Toys.AnimalRef.Name", "ID", "DeletedAt", "Toys.ID", "Toys.AnimalID", "Toys.AnimalRef.ID"}, c.permissionList.fields)

	// polymorphic relation
	c.SetPermissions(WHITELIST, "SpeciesPoly.Name")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"SpeciesPoly.Name", "ID", "DeletedAt", "SpeciesPoly.ID", "SpeciesID", "SpeciesPoly.SpeciesPolyType"}, c.permissionList.fields)

	// On blacklist the owner can name is disabled but the mandatory fields are deleted of the blacklist.
	c.SetPermissions(BLACKLIST, "SpeciesPoly.Name", "SpeciesID", "ID", "SpeciesPoly.ID", "SpeciesPoly.SpeciesPolyType")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"SpeciesPoly.Name"}, c.permissionList.fields)

	// full relation name - no additional keys needed to remove
	c.SetPermissions(BLACKLIST, "SpeciesPoly")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"SpeciesPoly"}, c.permissionList.fields)

	// child relation - theses are all mandatory keys and can not be removed.
	c.SetPermissions(BLACKLIST, "ID", "DeletedAt", "Toys.ID", "Toys.AnimalID", "Toys.AnimalRef.ID")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Nil(c.permissionList)

	// child relation - theses are all mandatory keys and can not be removed.
	c.SetPermissions(BLACKLIST, "Name", "Toys.AnimalRef.Name", "ID", "DeletedAt", "Toys.ID", "Toys.AnimalID", "Toys.AnimalRef.ID")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"Name", "Toys.AnimalRef.Name"}, c.permissionList.fields)

	// polymorphic relation
	c.SetPermissions(BLACKLIST, "SpeciesPoly.Name", "SpeciesID", "ID", "SpeciesPoly.ID", "SpeciesPoly.SpeciesPolyType")
	err = addMandatoryFields(c.scope)
	asserts.NoError(err)
	asserts.Equal([]string{"SpeciesPoly.Name"}, c.permissionList.fields)
}
