// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"time"

	"github.com/patrickascher/gofer/cache"
	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/stretchr/testify/assert"
)

var builder query.Builder
var builder2 query.Builder
var c cache.Manager

func helperCreateDatabaseAndTable(asserts *assert.Assertions) {
	cfg := testConfig()
	cfg.Database = ""
	b, err := query.New("mysql", cfg)

	_, err = b.Query().DB().Exec("DROP DATABASE IF EXISTS `tests`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("CREATE DATABASE `tests` DEFAULT CHARACTER SET = `utf8`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP DATABASE IF EXISTS `tests2`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("CREATE DATABASE `tests2` DEFAULT CHARACTER SET = `utf8`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`animals`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`animals` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '',  `species_id` int(11) DEFAULT NULL,`deleted_at` datetime DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`species`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE  `tests`.`species` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests2`.`species_polies`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE  `tests2`.`species_polies` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '',`species_poly_type` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`toys`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`toys` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', `brand` varchar(250) DEFAULT NULL, `destroy_able` tinyint(1) DEFAULT NULL, `bought_at` datetime DEFAULT NULL,  `animal_id` int(11) unsigned NOT NULL ,PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests2`.`toy_polies`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests2`.`toy_polies` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `toy_type` varchar(250) NOT NULL DEFAULT '', `name` varchar(250) NOT NULL DEFAULT '', `brand` varchar(250) DEFAULT NULL, `destroy_able` tinyint(1) DEFAULT NULL, `bought_at` datetime DEFAULT NULL,  `animal_id` int(11) unsigned NOT NULL ,PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`humans`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`humans` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests2`.`human_polies`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests2`.`human_polies` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`addresses`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`addresses` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `animal_id` int(11) NOT NULL, `street` varchar(250) NOT NULL, `zip` varchar(250) NOT NULL DEFAULT '', `city` varchar(205) NOT NULL DEFAULT '', `country` varchar(250) DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests2`.`address_polies`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests2`.`address_polies` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `animal_poly_id` int(11) NOT NULL, `animal_poly_type` varchar(250) NOT NULL,  `street` varchar(250) NOT NULL, `zip` varchar(250) NOT NULL DEFAULT '', `city` varchar(205) NOT NULL DEFAULT '', `country` varchar(250) DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`animal_walkers`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`animal_walkers` (`animal_id` int(11) unsigned NOT NULL, `human_id` int(11) unsigned NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`animal_walker_polies`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`animal_walker_polies` (`animal_id` int(11) unsigned NOT NULL, `animal_type` varchar(250) NOT NULL, `human_id` int(11) unsigned NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`roles`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`roles` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`role_roles`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`role_roles` (`role_id` int(11) unsigned NOT NULL, `child_id` int(11) unsigned NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	// set default builder
	builder, err = query.New("mysql", testConfig())
	asserts.NoError(err)

	// set default builder
	cfg = testConfig()
	cfg.Database = "tests2"
	builder2, err = query.New("mysql", cfg)
	asserts.NoError(err)

	// set default cache
	c, err = cache.New("memory", nil)
	asserts.NoError(err)
}

func insertUserData(asserts *assert.Assertions) {

	// add species
	values := []map[string]interface{}{
		{"id": 1, "name": "Dog"},
		{"id": 2, "name": "Cat"},
		{"id": 3, "name": "Lion"},
	}
	_, err := builder.Query().Insert("tests.species").Values(values).Exec()
	asserts.NoError(err)

	// add species
	values = []map[string]interface{}{
		{"id": 1, "name": "Dog", "species_poly_type": "Animal"},
		{"id": 2, "name": "Cat", "species_poly_type": "Animal"},
		{"id": 3, "name": "Lion", "species_poly_type": "Animal"},
		{"id": 4, "name": "Lion", "species_poly_type": "Lion"},
	}
	_, err = builder2.Query().Insert("tests2.species_polies").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "Blacky", "species_id": 1, "deleted_at": nil},
		{"id": 2, "name": "Snowflake", "species_id": 2, "deleted_at": nil},
		{"id": 3, "name": "Simba", "species_id": 3, "deleted_at": nil},
		{"id": 4, "name": "Nala", "species_id": 3, "deleted_at": query.NewNullTime(time.Now(), true)},
	}
	_, err = builder.Query().Insert("tests.animals").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "Bone", "brand": "Trixie", "destroy_able": query.NewNullBool(true, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 01, 10, 10, 10, 0, time.UTC), true), "animal_id": 1},
		{"id": 2, "name": "Kong", "brand": "Kong", "destroy_able": query.NewNullBool(false, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 02, 10, 10, 10, 0, time.UTC), true), "animal_id": 1},
		{"id": 3, "name": "Mouse", "brand": "", "destroy_able": query.NewNullBool(true, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 03, 10, 10, 10, 0, time.UTC), true), "animal_id": 2},
		{"id": 4, "name": "Nala", "brand": "", "destroy_able": query.NewNullBool(false, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 04, 10, 10, 10, 0, time.UTC), true), "animal_id": 3},
	}
	_, err = builder.Query().Insert("tests.toys").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "Bone", "brand": "Trixie", "destroy_able": query.NewNullBool(true, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 01, 10, 10, 10, 0, time.UTC), true), "animal_id": 1, "toy_type": "Animal"},
		{"id": 2, "name": "Kong", "brand": "Kong", "destroy_able": query.NewNullBool(false, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 02, 10, 10, 10, 0, time.UTC), true), "animal_id": 1, "toy_type": "Animal"},
		{"id": 3, "name": "Kong", "brand": "Kong", "destroy_able": query.NewNullBool(false, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 02, 10, 10, 10, 0, time.UTC), true), "animal_id": 1, "toy_type": "Something"},
		{"id": 4, "name": "Mouse", "brand": "", "destroy_able": query.NewNullBool(true, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 03, 10, 10, 10, 0, time.UTC), true), "animal_id": 2, "toy_type": "Animal"},
		{"id": 5, "name": "Nala", "brand": "", "destroy_able": query.NewNullBool(false, true), "bought_at": query.NewNullTime(time.Date(2021, 01, 04, 10, 10, 10, 0, time.UTC), true), "animal_id": 3, "toy_type": "Animal"},
	}
	_, err = builder.Query().Insert("tests2.toy_polies").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "Pat"},
		{"id": 2, "name": "Eva"},
		{"id": 3, "name": "Christian"},
	}
	_, err = builder.Query().Insert("tests.humans").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "Pat"},
		{"id": 2, "name": "Eva"},
		{"id": 3, "name": "Christian"},
	}
	_, err = builder.Query().Insert("tests2.human_polies").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "animal_id": 1, "street": "Erika-Mann-Str. 33", "zip": "80636", "city": "Munich", "country": "Germany"},
		{"id": 2, "animal_id": 2, "street": "Brandschenkestrasse", "zip": "8002", "city": "Zürich", "country": "Switzerland"},
		//{"id": 3, "animal_id": 3, "street": "76 Buckingham Palace Road", "zip": "SW1W 9TQ", "city": "London", "country": "United Kingdom"},
	}
	_, err = builder.Query().Insert("tests.addresses").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "animal_poly_id": 1, "animal_poly_type": "Animal", "street": "Erika-Mann-Str. 33", "zip": "80636", "city": "Munich", "country": "Germany"},
		{"id": 2, "animal_poly_id": 2, "animal_poly_type": "Animal", "street": "Brandschenkestrasse", "zip": "8002", "city": "Zürich", "country": "Switzerland"},
		{"id": 3, "animal_poly_id": 2, "animal_poly_type": "Something", "street": "Brandschenkestrasse", "zip": "8002", "city": "Zürich", "country": "Switzerland"},
	}
	_, err = builder2.Query().Insert("tests2.address_polies").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"animal_id": 1, "human_id": 1},
		{"animal_id": 1, "human_id": 2},
		{"animal_id": 2, "human_id": 2},
	}
	_, err = builder.Query().Insert("tests.animal_walkers").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"animal_id": 1, "human_id": 1, "animal_type": "Fast"},
		{"animal_id": 1, "human_id": 2, "animal_type": "Fast"},
		{"animal_id": 2, "human_id": 2, "animal_type": "Fast"},
		{"animal_id": 2, "human_id": 1, "animal_type": "Slow"},
	}
	_, err = builder.Query().Insert("tests.animal_walker_polies").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"id": 1, "name": "RoleA"},
		{"id": 2, "name": "RoleB"},
		{"id": 3, "name": "RoleC"},
		{"id": 4, "name": "Loop-1"},
		{"id": 5, "name": "Loop-2"},
	}
	_, err = builder.Query().Insert("tests.roles").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"role_id": 1, "child_id": 2},
		{"role_id": 2, "child_id": 3},
		{"role_id": 4, "child_id": 5},
		{"role_id": 5, "child_id": 4},
	}
	_, err = builder.Query().Insert("tests.role_roles").Values(values).Exec()
	asserts.NoError(err)
}

type Base struct {
	orm.Model
	ID int
}

func (b Base) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (b Base) DefaultBuilder() query.Builder {
	return builder
}

type Species struct {
	Base
	Name string
}

type SpeciesPoly struct {
	orm.Model
	ID              int
	Name            string
	SpeciesPolyType string
}

func (b SpeciesPoly) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (b SpeciesPoly) DefaultBuilder() query.Builder {
	return builder2
}

type Toy struct {
	Base
	Name        string
	Brand       query.NullString
	DestroyAble query.NullBool
	BoughtAt    query.NullTime

	AnimalID int

	//Animal    Animal  `orm:"relation:belongsTo;"`// should end in a infinity loop.
	AnimalRef *Animal `orm:"relation:belongsTo;"` // should create a reference.
}

type ToyPoly struct {
	orm.Model
	ID      int
	ToyType string

	Name        string
	Brand       query.NullString
	DestroyAble query.NullBool
	BoughtAt    query.NullTime

	AnimalID int

	//Animal    Animal  `orm:"relation:belongsTo;"`// should end in a infinity loop.
	//AnimalRef *Animal `orm:"relation:belongsTo;"` // should create a reference.
}

func (t ToyPoly) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (t ToyPoly) DefaultBuilder() query.Builder {
	return builder2
}

type Human struct {
	Base
	Name string
	//Animal []Animal
}

func (h Human) DefaultTableName() string {
	return "humans"
}

type HumanPoly struct {
	Base
	Name string
	//Animal []Animal
}

func (t HumanPoly) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (t HumanPoly) DefaultBuilder() query.Builder {
	return builder2
}

type Role struct {
	Base
	Name string

	Roles []Role
}

type Address struct {
	Base

	AnimalID int
	//Animal   *Animal `orm:"relation:belongsTo"` // TODO: will end in a loop because of validation.v10

	Street  query.NullString
	Zip     string
	City    string
	Country query.NullString
}

type AddressPoly struct {
	orm.Model
	ID int

	AnimalPolyID   int
	AnimalPolyType string
	//Animal         *Animal `orm:"relation:belongsTo;fk:AnimalPolyID"` // TODO: will end in a loop because of validation.v10

	Street  query.NullString
	Zip     string
	City    string
	Country query.NullString
}

func (a AddressPoly) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (a AddressPoly) DefaultBuilder() query.Builder {
	return builder2
}

type Animal struct {
	Base
	Name string

	// BelongsTo
	SpeciesID      query.NullInt
	Species        Species      `orm:"relation:belongsTo"`
	SpeciesPtr     *Species     `orm:"relation:belongsTo"`
	SpeciesPoly    SpeciesPoly  `orm:"relation:belongsTo;fk:SpeciesID;refs:ID;poly"`
	SpeciesPolyPtr *SpeciesPoly `orm:"relation:belongsTo;fk:SpeciesID;refs:ID;poly:SpeciesPoly;poly_value:Animal"`

	// HasOne
	Address        Address      // belongsTo Animal
	AddressPtr     *Address     // belongsTo Animal
	AddressPoly    AddressPoly  `orm:"poly:AnimalPoly;refs:AnimalPolyID"`
	AddressPolyPtr *AddressPoly `orm:"poly:AnimalPoly;poly_value:Animal;refs:AnimalPolyID"`

	// HasMany
	Toys               []Toy // belongsTo Animal...
	ToysSlicePtr       []*Toy
	ToysPtrSlice       *[]Toy
	ToysPtrSlicePtr    *[]*Toy
	ToyPoly            []ToyPoly   `orm:"poly:Toy;refs:AnimalID"`
	ToyPolySlicePtr    []*ToyPoly  `orm:"poly:Toy;refs:AnimalID"`
	ToyPolyPtrSlice    *[]ToyPoly  `orm:"poly:Toy;refs:AnimalID"`
	ToyPolyPtrSlicePtr *[]*ToyPoly `orm:"poly:Toy;refs:AnimalID"`

	// ManyToMany
	Walkers                []Human       `orm:"relation:m2m;join_table:animal_walkers"`
	WalkersSlicePtr        []*Human      `orm:"relation:m2m;join_table:animal_walkers"`
	WalkersPtrSlice        *[]Human      `orm:"relation:m2m;join_table:animal_walkers"`
	WalkersPtrSlicePtr     *[]*Human     `orm:"relation:m2m;join_table:animal_walkers"`
	WalkersPoly            []HumanPoly   `orm:"relation:m2m;join_refs:human_id;join_table:animal_walker_polies;poly:Animal;poly_value:Fast"`
	WalkersPolySlicePtr    []*HumanPoly  `orm:"relation:m2m;join_refs:human_id;join_table:animal_walker_polies;poly:Animal;poly_value:Fast"`
	WalkersPolyPtrSlice    *[]HumanPoly  `orm:"relation:m2m;join_refs:human_id;join_table:animal_walker_polies;poly:Animal;poly_value:Fast"`
	WalkersPolyPtrSlicePtr *[]*HumanPoly `orm:"relation:m2m;join_refs:human_id;join_table:animal_walker_polies;poly:Animal;poly_value:Fast"`
}
