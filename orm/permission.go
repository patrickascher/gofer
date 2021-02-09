// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"strings"

	"github.com/patrickascher/gofer/slicer"
)

// permission policy
const (
	BLACKLIST = 0
	WHITELIST = 1
)

// permissionList struct
type permissionList struct {
	policy   int
	fields   []string
	explicit bool
}

// newPermissionList creates a new white/blacklist with the given policy and fields.
func newPermissionList(policy int, fields []string) *permissionList {
	wb := &permissionList{}
	wb.policy = policy
	wb.fields = fields
	return wb
}

// setFieldPermission sets the permission read/write for all columns for the given black/whitelist.
// It is called on first, all, create, update and delete.
// The fields wb fields are not getting decreased, because they are added to the child object on a self referencing model.
func (s scope) setFieldPermission() error {

	// adding/removing mandatory fields of the wb list.
	// foreign keys, primary keys and the time fields are always allowed.
	err := addMandatoryFields(s)
	if err != nil {
		return err
	}

	// if no wb list is defined, return.
	if s.model.permissionList == nil {
		return nil
	}

	whitelisted := false
	if s.model.permissionList.policy == WHITELIST {
		whitelisted = true
	}

	// loop over all sql fields
fields:
	for i := range s.model.fields {

		// set all fields to the opposite value.
		s.model.fields[i].Permission.Read = !whitelisted
		s.model.fields[i].Permission.Write = !whitelisted

		// loop over the given white/blacklist fields and set the given policy
		for _, wbField := range s.model.permissionList.fields {
			if s.model.fields[i].Name == wbField {
				s.model.fields[i].Permission.Read = whitelisted
				s.model.fields[i].Permission.Write = whitelisted
				continue fields
			}
		}
	}

	// loop over relations
relations:
	for i := range s.model.relations {

		// set all fields to the opposite value.
		s.model.relations[i].Permission.Read = !whitelisted
		s.model.relations[i].Permission.Write = !whitelisted

		// loop over the white/blacklist fields and set the given policy
		// if there is a dot notation (example User.Name) the User relation is set to required on a whitelist.
		for _, wbField := range s.model.permissionList.fields {
			if s.model.relations[i].Field == wbField || (s.model.permissionList.policy == WHITELIST && strings.HasPrefix(wbField, s.model.relations[i].Field+".")) {
				// if the relation exist, by the name add or remove it
				// if whitelist and its a relation dot notation, add the relation because its mandatory.
				s.model.relations[i].Permission.Read = whitelisted
				s.model.relations[i].Permission.Write = whitelisted

				if s.model.relations[i].Field == wbField {
					// if a relation is added completely and there is also a dot notation on that relation, remove it because the whole relation is added anyway.
					if deleteFields := slicer.StringPrefixExists(s.model.permissionList.fields, s.model.relations[i].Field+"."); len(deleteFields) > 0 {
						for _, deleteField := range deleteFields {
							if i, exists := slicer.StringExists(s.model.permissionList.fields, deleteField); exists {
								s.model.permissionList.fields = append(s.model.permissionList.fields[:i], s.model.permissionList.fields[i+1:]...)
							}
						}
					}
				}

				continue relations
			}
		}
	}

	return nil
}

// addMandatoryFields will add all foreign keys, primary keys, time fields, soft delete and poly fields
// If a full relation is added, it will be ignored on Blacklist, because all fields are added or removed anyway. On Whitelist the foreign key is added.
// If something like Relation.Child1.Child2.Name exists, it will recursively add all mandatory keys.
// If the wb list is empty, it will skip.
func addMandatoryFields(s scope) error {

	// skip if no  wb list is defined
	if s.model.permissionList == nil {
		return nil
	}

	// field list
	var rv []string

	// always allow primary keys
	// needed for select + relations, create, update, delete
	pKeys, err := s.PrimaryKeys()
	if err != nil {
		return err
	}
	for _, pkey := range pKeys {
		rv = append(rv, pkey.Name)
	}

	// relation permission
	// its set to false because otherwise the user could disable mandatory fields by tag.
	perm := Permission{}

	// time fields and soft deletion field is always added if they exist.
	for _, f := range s.SQLFields(perm) {
		if f.Name == CreatedAt || f.Name == UpdatedAt || f.Name == DeletedAt {
			rv = append(rv, f.Name)
		}
		if s.model.softDelete != nil && f.Information.Name == s.model.softDelete.Field {
			rv = append(rv, f.Name)
		}
	}

	for _, relation := range s.SQLRelations(perm) {

		// Whole relations on white or blacklist can be ignored because they are added completely.
		// Only on a whitelist the relation fk is added because the data is needed for referencing.
		// The fk is only added if it does not exist yet.
		if _, exists := slicer.StringExists(s.model.permissionList.fields, relation.Field); exists {
			if s.model.permissionList.policy == WHITELIST {
				if _, exists := slicer.StringExists(rv, relation.Mapping.ForeignKey.Name); !exists {
					rv = append(rv, relation.Mapping.ForeignKey.Name)
				}
			}
			continue
		}

		// Relation with dot notations adds all mandatory fields recursively.
		// The fields are only added if they dont exist yet.
		if relChild := slicer.StringPrefixExists(s.model.permissionList.fields, relation.Field+"."); relChild != nil {
			for _, rc := range relChild {
				relFields, err := mandatoryKeys(s, perm, relation, strings.Split(rc, ".")[1:])
				if err != nil {
					return err
				}
				for _, relField := range relFields {
					if _, exists := slicer.StringExists(rv, relField); !exists {
						rv = append(rv, relField)
					}
				}
			}
		} else {
			// If relation is not blacklisted, the foreign key field is required and not allowed to be blacklisted.
			if s.model.permissionList.policy == BLACKLIST {
				if _, exists := slicer.StringExists(rv, relation.Mapping.ForeignKey.Name); !exists {
					rv = append(rv, relation.Mapping.ForeignKey.Name)
				}
			}
		}
	}

	// If its a whitelist, the needed keys will be merged with the existing list.
	// On a blacklist, the required fields will be excluded of the wb list, because they are mandatory for the relation chain.
	if s.model.permissionList.policy == WHITELIST {
		s.model.permissionList.fields = slicer.StringUnique(append(s.model.permissionList.fields, rv...))
	} else {
		for _, r := range rv {
			if p, exists := slicer.StringExists(s.model.permissionList.fields, r); exists {
				s.model.permissionList.fields = append(s.model.permissionList.fields[:p], s.model.permissionList.fields[p+1:]...)
			}
		}
	}

	// if the whole wb list is dissolve because the blacklisted fields are required, set the wb list to nil.
	if len(s.model.permissionList.fields) == 0 {
		s.model.permissionList = nil
	}

	return nil
}

// mandatoryKeys recursively adds all required keys (primary, fk, afk, poly)
// If the policy is Whitelist, all fields are added additional to the custom wb list.
// On Blacklist, the fields are removed from the wb list, because they are mandatory to guarantee the relation chain.
func mandatoryKeys(scope scope, p Permission, relation Relation, fields []string) ([]string, error) {
	var rv []string

	// relation model from cache
	relScope, err := scope.NewScopeFromType(relation.Type)
	if err != nil {
		return nil, err
	}

	// add all primary keys of the relation model
	pKeys, err := relScope.PrimaryKeys()
	if err != nil {
		return nil, err
	}
	for _, pkey := range pKeys {
		if _, exists := slicer.StringExists(rv, relation.Field+"."+pkey.Name); !exists {
			rv = append(rv, relation.Field+"."+pkey.Name)
		}
	}

	// add the foreign key and references
	if _, exists := slicer.StringExists(rv, relation.Mapping.ForeignKey.Name); !exists {
		rv = append(rv, relation.Mapping.ForeignKey.Name)
	}
	if _, exists := slicer.StringExists(rv, relation.Field+"."+relation.Mapping.References.Name); !exists {
		rv = append(rv, relation.Field+"."+relation.Mapping.References.Name)
	}

	// add polymorphic type field
	// relation.isPolymorphic is not allowed to use because its also true on a junction poly table.
	if relation.Mapping.Polymorphic.Value != "" {
		if _, exists := slicer.StringExists(rv, relation.Field+"."+relation.Mapping.Polymorphic.TypeField.Name); !exists {
			rv = append(rv, relation.Field+"."+relation.Mapping.Polymorphic.TypeField.Name)
		}
	}

	// if the depth of the added fields are bigger than 1, check the relation and run it recursively.
	if len(fields) > 1 {
		// if there is still a relation, get the cached model
		childRel, err := relScope.SQLRelation(fields[0], p)
		if err != nil {
			return nil, err
		}
		childScope, err := scope.NewScopeFromType(childRel.Type)
		if err != nil {
			return nil, err
		}

		// recursively add all mandatory fields
		relFields, err := mandatoryKeys(childScope.Model().scope, p, childRel, fields[1:])
		if err != nil {
			return nil, err
		}
		for _, relField := range relFields {
			if _, exists := slicer.StringExists(rv, relation.Field+"."+relField); !exists {
				rv = append(rv, relation.Field+"."+relField)
			}
		}
	}

	return rv, nil
}
