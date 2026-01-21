// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package schema

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	entdrv "entgo.io/ent/dialect/ydb"
	"entgo.io/ent/schema/field"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	atlas "ariga.io/atlas/sql/ydb"
)

// YDB adapter for Atlas migration engine.
type YDB struct {
	dialect.Driver

	version string
}

// init loads the YDB version from the database for later use in the migration process.
func (d *YDB) init(ctx context.Context) error {
	if d.version != "" {
		return nil // already initialized.
	}

	rows := &entsql.Rows{}
	if err := d.Driver.Query(ctx, "SELECT version()", []any{}, rows); err != nil {
		return fmt.Errorf("ydb: failed to query version: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return fmt.Errorf("ydb: version was not found")
	}

	var version string
	if err := rows.Scan(&version); err != nil {
		return fmt.Errorf("ydb: failed to scan version: %w", err)
	}

	d.version = version
	return nil
}

// tableExist checks if a table exists in the database by querying the .sys/tables system table.
func (d *YDB) tableExist(ctx context.Context, conn dialect.ExecQuerier, name string) (bool, error) {
	query, args := entsql.Dialect(dialect.YDB).
		Select(entsql.Count("*")).
		From(entsql.Table(".sys/tables")).
		Where(entsql.EQ("table_name", name)).
		Query()

	return exist(ctx, conn, query, args...)
}

// atOpen returns a custom Atlas migrate.Driver for YDB.
func (d *YDB) atOpen(conn dialect.ExecQuerier) (migrate.Driver, error) {
	var ydbDriver *entdrv.YDBDriver

	switch drv := conn.(type) {
	case *entdrv.YDBDriver:
		ydbDriver = drv
	case *YDB:
		if ydb, ok := drv.Driver.(*entdrv.YDBDriver); ok {
			ydbDriver = ydb
		}
	}
	if ydbDriver == nil {
		if ydb, ok := d.Driver.(*entdrv.YDBDriver); ok {
			ydbDriver = ydb
		} else {
			return nil, fmt.Errorf("expected dialect/ydb.YDBDriver, but got %T", conn)
		}
	}

	return atlas.Open(
		ydbDriver.NativeDriver(),
		ydbDriver.DB(),
	)
}

func (d *YDB) atTable(table1 *Table, table2 *schema.Table) {
	if table1.Annotation != nil {
		setAtChecks(table1, table2)
	}
}

// supportsDefault returns whether YDB supports DEFAULT values for the given column type.
func (d *YDB) supportsDefault(column *Column) bool {
	switch column.Default.(type) {
	case Expr, map[string]Expr:
		// Expression defaults are not well supported in YDB
		return false
	default:
		// Simple literal defaults should work for basic types
		return column.supportDefault()
	}
}

// atTypeC converts an Ent column type to a YDB Atlas schema type.
func (d *YDB) atTypeC(column1 *Column, column2 *schema.Column) error {
	// Check for custom schema type override.
	if column1.SchemaType != nil && column1.SchemaType[dialect.YDB] != "" {
		typ, err := atlas.ParseType(
			column1.SchemaType[dialect.YDB],
		)
		if err != nil {
			return err
		}
		column2.Type.Type = typ
		return nil
	}

	var (
		typ schema.Type
		err error
	)

	switch column1.Type {
	case field.TypeBool:
		typ = &schema.BoolType{T: atlas.TypeBool}
	case field.TypeInt8:
		typ = &schema.IntegerType{T: atlas.TypeInt8}
	case field.TypeInt16:
		typ = &schema.IntegerType{T: atlas.TypeInt16}
	case field.TypeInt32:
		typ = &schema.IntegerType{T: atlas.TypeInt32}
	case field.TypeInt, field.TypeInt64:
		typ = &schema.IntegerType{T: atlas.TypeInt64}
	case field.TypeUint8:
		typ = &schema.IntegerType{T: atlas.TypeUint8, Unsigned: true}
	case field.TypeUint16:
		typ = &schema.IntegerType{T: atlas.TypeUint16, Unsigned: true}
	case field.TypeUint32:
		typ = &schema.IntegerType{T: atlas.TypeUint32, Unsigned: true}
	case field.TypeUint, field.TypeUint64:
		typ = &schema.IntegerType{T: atlas.TypeUint64, Unsigned: true}
	case field.TypeFloat32:
		typ = &schema.FloatType{T: atlas.TypeFloat}
	case field.TypeFloat64:
		typ = &schema.FloatType{T: atlas.TypeDouble}
	case field.TypeBytes:
		typ = &schema.BinaryType{T: atlas.TypeString}
	case field.TypeString:
		typ = &schema.StringType{T: atlas.TypeUtf8}
	case field.TypeJSON:
		typ = &schema.JSONType{T: atlas.TypeJSON}
	case field.TypeTime:
		typ = &schema.TimeType{T: atlas.TypeTimestamp}
	case field.TypeUUID:
		typ = &schema.UUIDType{T: atlas.TypeUUID}
	case field.TypeEnum:
		// YDB doesn't support enum types in DDL statements
		// But ent can handle enum validation, so we just map it to Utf8
		typ = &schema.StringType{T: atlas.TypeUtf8}
	case field.TypeOther:
		typ = &schema.UnsupportedType{T: column1.typ}
	default:
		typ, err = atlas.ParseType(column1.typ)
	}

	if err != nil {
		return err
	}

	column2.Type.Type = typ
	return nil
}

// atUniqueC adds a unique constraint for a column.
// In YDB, unique constraints are implemented as GLOBAL UNIQUE SYNC indexes.
func (d *YDB) atUniqueC(
	table1 *Table,
	column1 *Column,
	table2 *schema.Table,
	column2 *schema.Column,
) {
	if !canBeIndexKey(column1) {
		return
	}

	// Check if there's already an explicit unique index defined for this column.
	for _, idx := range table1.Indexes {
		if idx.Unique && len(idx.Columns) == 1 && idx.Columns[0].Name == column1.Name {
			// Index already defined explicitly, will be added in atIndexes.
			return
		}
	}
	// Create a unique index for this column.
	idxName := fmt.Sprintf("%s_%s_index", table1.Name, column1.Name)
	index := schema.NewUniqueIndex(idxName).AddColumns(column2)

	table2.AddIndexes(index)
}

// atIncrementC configures auto-increment for a column.
// YDB uses Serial types for auto-increment.
func (d *YDB) atIncrementC(table *schema.Table, column *schema.Column) {
	if intType, ok := column.Type.Type.(*schema.IntegerType); ok {
		serial, err := atlas.SerialFromInt(intType)
		if err != nil {
			panic(err)
		}
		column.Type.Type = serial
	}
}

// atIncrementT sets the table-level auto-increment starting value.
func (d *YDB) atIncrementT(table *schema.Table, v int64) {
	// not implemented
}

// atIndex configures an index for ydb.
func (d *YDB) atIndex(
	index1 *Index,
	table2 *schema.Table,
	index2 *schema.Index,
) error {
	for _, column1 := range index1.Columns {
		if isPrimaryKeyColumn(table2, column1.Name) {
			continue
		}

		column2, ok := table2.Column(column1.Name)
		if !ok {
			return fmt.Errorf("unexpected index %q column: %q", index1.Name, column1.Name)
		}

		if !canBeIndexKeyBySchema(column2) {
			continue
		}

		index2.AddParts(&schema.IndexPart{C: column2})
	}

	// Set YDB-specific index attributes.
	// By default, use GLOBAL SYNC for consistency.
	idxAttrs := &atlas.IndexAttributes{}

	if index1.Annotation != nil {
		annotation := index1.Annotation

		if len(annotation.IncludeColumns) > 0 {
			columns := make([]*schema.Column, len(annotation.IncludeColumns))

			for i, include := range annotation.IncludeColumns {
				column, ok := table2.Column(include)
				if !ok {
					return fmt.Errorf("include column %q was not found for index %q", include, index1.Name)
				}
				columns[i] = column
			}

			idxAttrs.CoverColumns = columns
		}

		if indexType, ok := indexType(index1, dialect.YDB); ok {
			upperIndexType := strings.ToUpper(indexType)

			if strings.Contains(upperIndexType, "ASYNC") {
				idxAttrs.Async = true
			}
		}
	}

	index2.AddAttrs(idxAttrs)
	return nil
}

// atTypeRangeSQL returns the SQL statement to insert type ranges for global unique IDs.
func (*YDB) atTypeRangeSQL(ts ...string) string {
	values := make([]string, len(ts))
	for i, t := range ts {
		values[i] = fmt.Sprintf("('%s')", t)
	}
	return fmt.Sprintf(
		"UPSERT INTO `%s` (`type`) VALUES %s",
		TypeTable,
		strings.Join(values, ", "),
	)
}

// canBeIndexKey checks if a column type can be used as an index key in YDB.
// YDB doesn't allow Float/Double types as index keys.
func canBeIndexKey(column *Column) bool {
	switch column.Type {
	case field.TypeFloat32, field.TypeFloat64:
		return false
	default:
		return true
	}
}

// canBeIndexKeyBySchema checks if a column type can be used as an index key in YDB
// by checking the Atlas schema column type.
func canBeIndexKeyBySchema(column *schema.Column) bool {
	if column.Type == nil || column.Type.Type == nil {
		return true
	}
	switch column.Type.Type.(type) {
	case *schema.FloatType:
		return false
	default:
		return true
	}
}

// isPrimaryKeyColumn checks if a column is part of the table's primary key.
func isPrimaryKeyColumn(table *schema.Table, columnName string) bool {
	if table.PrimaryKey == nil {
		return false
	}
	for _, primaryKeyPart := range table.PrimaryKey.Parts {
		if primaryKeyPart.C != nil && primaryKeyPart.C.Name == columnName {
			return true
		}
	}
	return false
}
