package dm

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"
)

var _ driver.Grammar = &Grammar{}

type Grammar struct {
	attributeCommands []string
	modifiers         []func(driver.Blueprint, driver.ColumnDefinition) string
	prefix            string
	serials           []string
	wrap              *schema.Wrap
}

func NewGrammar(prefix string) *Grammar {
	grammar := &Grammar{
		attributeCommands: []string{schema.CommandComment},
		prefix:            prefix,
		serials:           []string{"bigInteger", "integer", "mediumInteger", "smallInteger", "tinyInteger"},
		wrap:              schema.NewWrap(prefix),
	}
	grammar.modifiers = []func(driver.Blueprint, driver.ColumnDefinition) string{
		grammar.ModifyDefault,
		grammar.ModifyIncrement,
		grammar.ModifyNullable,
		grammar.ModifyGeneratedAsForChange,
		grammar.ModifyGeneratedAs,
	}

	return grammar
}

func (r *Grammar) CompileAdd(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s add column %s", r.wrap.Table(blueprint.GetTableName()), r.getColumn(blueprint, command.Column))
}

func (r *Grammar) CompileChange(blueprint driver.Blueprint, command *driver.Command) []string {
	changes := []string{fmt.Sprintf("alter column %s type %s", r.wrap.Column(command.Column.GetName()), schema.ColumnType(r, command.Column))}
	for _, modifier := range r.modifiers {
		if change := modifier(blueprint, command.Column); change != "" {
			changes = append(changes, fmt.Sprintf("alter column %s%s", r.wrap.Column(command.Column.GetName()), change))
		}
	}

	return []string{
		fmt.Sprintf("alter table %s %s", r.wrap.Table(blueprint.GetTableName()), strings.Join(changes, ", ")),
	}
}

func (r *Grammar) CompileColumns(schema, table string) (string, error) {
	schema, table, err := parseSchemaAndTable(table, schema)
	if err != nil {
		return "", err
	}
	table = r.prefix + table

	return fmt.Sprintf(
		`SELECT c.COLUMN_NAME AS name,
        c.DATA_TYPE AS type_name,
        CASE
          WHEN c.DATA_TYPE IN ('CHAR','NCHAR','VARCHAR','VARCHAR2','NVARCHAR2')
            THEN c.DATA_TYPE || '(' || c.CHAR_LENGTH || ')'
          WHEN c.DATA_TYPE IN ('NUMBER','DECIMAL','NUMERIC')
            THEN c.DATA_TYPE || '(' || c.DATA_PRECISION || ',' || c.DATA_SCALE || ')'
          ELSE c.DATA_TYPE
        END AS type,
        '' AS collation,
        CASE c.NULLABLE WHEN 'Y' THEN 1 ELSE 0 END AS nullable,
        c.DATA_DEFAULT AS default,
        cc.COMMENTS AS comment
FROM ALL_TAB_COLUMNS c
LEFT JOIN ALL_COL_COMMENTS cc
  ON cc.OWNER = c.OWNER AND cc.TABLE_NAME = c.TABLE_NAME AND cc.COLUMN_NAME = c.COLUMN_NAME
WHERE c.OWNER = %s AND c.TABLE_NAME = %s
ORDER BY c.COLUMN_ID`, r.wrap.Quote(strings.ToUpper(schema)), r.wrap.Quote(strings.ToUpper(table))), nil
}

func (r *Grammar) CompileComment(blueprint driver.Blueprint, command *driver.Command) string {
	comment := "NULL"
	if command.Column.IsSetComment() {
		comment = r.wrap.Quote(strings.ReplaceAll(command.Column.GetComment(), "'", "''"))
	}

	return fmt.Sprintf("comment on column %s.%s is %s",
		r.wrap.Table(blueprint.GetTableName()),
		r.wrap.Column(command.Column.GetName()),
		comment)
}

func (r *Grammar) CompileCreate(blueprint driver.Blueprint) string {
	return fmt.Sprintf("create table %s (%s)", r.wrap.Table(blueprint.GetTableName()), strings.Join(r.getColumns(blueprint), ", "))
}

func (r *Grammar) CompileDefault(_ driver.Blueprint, _ *driver.Command) string { return "" }

func (r *Grammar) CompileDrop(blueprint driver.Blueprint) string {
	return fmt.Sprintf("drop table %s", r.wrap.Table(blueprint.GetTableName()))
}

func (r *Grammar) CompileDropAllDomains(domains []string) string {
	if len(domains) == 0 {
		return ""
	}
	return fmt.Sprintf("drop domain %s", strings.Join(r.EscapeNames(domains), ", "))
}

func (r *Grammar) CompileDropAllTables(schema string, tables []driver.Table) []string {
	excludedTables := r.EscapeNames([]string{"spatial_ref_sys"})
	escapedSchema := r.EscapeNames([]string{schema})[0]
	var dropTables []string
	for _, table := range tables {
		qualifiedName := fmt.Sprintf("%s.%s", table.Schema, table.Name)
		isExcludedTable := slices.Contains(excludedTables, qualifiedName) || slices.Contains(excludedTables, table.Name)
		isInCurrentSchema := escapedSchema == r.EscapeNames([]string{table.Schema})[0]
		if !isExcludedTable && isInCurrentSchema {
			dropTables = append(dropTables, qualifiedName)
		}
	}
	if len(dropTables) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("drop table %s", strings.Join(r.EscapeNames(dropTables), ", "))}
}

func (r *Grammar) CompileDropAllTypes(schema string, types []driver.Type) []string {
	var dropTypes, dropDomains []string
	for _, t := range types {
		if !t.Implicit && schema == t.Schema {
			if t.Type == "domain" {
				dropDomains = append(dropDomains, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			} else {
				dropTypes = append(dropTypes, fmt.Sprintf("%s.%s", t.Schema, t.Name))
			}
		}
	}
	var sql []string
	if len(dropTypes) > 0 {
		sql = append(sql, fmt.Sprintf("drop type %s", strings.Join(r.EscapeNames(dropTypes), ", ")))
	}
	if len(dropDomains) > 0 {
		sql = append(sql, fmt.Sprintf("drop domain %s", strings.Join(r.EscapeNames(dropDomains), ", ")))
	}
	return sql
}

func (r *Grammar) CompileDropAllViews(schema string, views []driver.View) []string {
	var dropViews []string
	for _, view := range views {
		if schema == view.Schema {
			dropViews = append(dropViews, fmt.Sprintf("%s.%s", view.Schema, view.Name))
		}
	}
	if len(dropViews) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("drop view %s", strings.Join(r.EscapeNames(dropViews), ", "))}
}

func (r *Grammar) CompileDropColumn(blueprint driver.Blueprint, command *driver.Command) []string {
	columns := r.wrap.PrefixArray("drop column", r.wrap.Columns(command.Columns))
	return []string{fmt.Sprintf("alter table %s %s", r.wrap.Table(blueprint.GetTableName()), strings.Join(columns, ", "))}
}

func (r *Grammar) CompileDropForeign(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s drop constraint %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index))
}
func (r *Grammar) CompileDropFullText(blueprint driver.Blueprint, command *driver.Command) string {
	return r.CompileDropIndex(blueprint, command)
}
func (r *Grammar) CompileDropIfExists(blueprint driver.Blueprint) string {
	return fmt.Sprintf("drop table if exists %s", r.wrap.Table(blueprint.GetTableName()))
}
func (r *Grammar) CompileDropIndex(_ driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("drop index %s", r.wrap.Column(command.Index))
}
func (r *Grammar) CompileDropPrimary(blueprint driver.Blueprint, _ *driver.Command) string {
	tableName := blueprint.GetTableName()
	index := r.wrap.Column(fmt.Sprintf("%s%s_pkey", r.wrap.GetPrefix(), tableName))
	return fmt.Sprintf("alter table %s drop constraint %s", r.wrap.Table(tableName), index)
}
func (r *Grammar) CompileDropUnique(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s drop constraint %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index))
}
func (r *Grammar) CompileForeign(blueprint driver.Blueprint, command *driver.Command) string {
	sql := fmt.Sprintf("alter table %s add constraint %s foreign key (%s) references %s (%s)",
		r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index), r.wrap.Columnize(command.Columns), r.wrap.Table(command.On), r.wrap.Columnize(command.References))
	if command.OnDelete != "" {
		sql += " on delete " + command.OnDelete
	}
	if command.OnUpdate != "" {
		sql += " on update " + command.OnUpdate
	}
	return sql
}

func (r *Grammar) CompileForeignKeys(schema, table string) string {
	return fmt.Sprintf(
		`SELECT
  uc.CONSTRAINT_NAME AS name,
  LISTAGG(cols.COLUMN_NAME, ',') WITHIN GROUP (ORDER BY cols.POSITION) AS columns,
  ruc.OWNER AS foreign_schema,
  ruc.TABLE_NAME AS foreign_table,
  LISTAGG(rcols.COLUMN_NAME, ',') WITHIN GROUP (ORDER BY rcols.POSITION) AS foreign_columns,
  '' AS on_update,
  LOWER(uc.DELETE_RULE) AS on_delete
FROM ALL_CONSTRAINTS uc
JOIN ALL_CONS_COLUMNS cols
  ON cols.OWNER = uc.OWNER AND cols.CONSTRAINT_NAME = uc.CONSTRAINT_NAME
JOIN ALL_CONSTRAINTS ruc
  ON ruc.OWNER = uc.R_OWNER AND ruc.CONSTRAINT_NAME = uc.R_CONSTRAINT_NAME
JOIN ALL_CONS_COLUMNS rcols
  ON rcols.OWNER = ruc.OWNER AND rcols.CONSTRAINT_NAME = ruc.CONSTRAINT_NAME AND rcols.POSITION = cols.POSITION
WHERE uc.CONSTRAINT_TYPE = 'R'
  AND uc.OWNER = %s
  AND uc.TABLE_NAME = %s
GROUP BY uc.CONSTRAINT_NAME, ruc.OWNER, ruc.TABLE_NAME, uc.DELETE_RULE`,
		r.wrap.Quote(strings.ToUpper(schema)), r.wrap.Quote(strings.ToUpper(table)))
}

func (r *Grammar) CompileFullText(blueprint driver.Blueprint, command *driver.Command) string {
	language := "english"
	if command.Language != "" {
		language = command.Language
	}
	columns := collect.Map(command.Columns, func(column string, _ int) string {
		return fmt.Sprintf("to_tsvector(%s, %s)", r.wrap.Quote(language), r.wrap.Column(column))
	})
	return fmt.Sprintf("create index %s on %s using gin(%s)", r.wrap.Column(command.Index), r.wrap.Table(blueprint.GetTableName()), strings.Join(columns, " || "))
}

func (r *Grammar) CompileIndex(blueprint driver.Blueprint, command *driver.Command) string {
	var algorithm string
	if command.Algorithm != "" {
		algorithm = " using " + command.Algorithm
	}
	return fmt.Sprintf("create index %s on %s%s (%s)", r.wrap.Column(command.Index), r.wrap.Table(blueprint.GetTableName()), algorithm, r.wrap.Columnize(command.Columns))
}

func (r *Grammar) CompileIndexes(schema, table string) (string, error) {
	schema, table, err := parseSchemaAndTable(table, schema)
	if err != nil {
		return "", err
	}
	table = r.prefix + table
	return fmt.Sprintf(
		`SELECT
  ui.INDEX_NAME AS name,
  LISTAGG(uic.COLUMN_NAME, ',') WITHIN GROUP (ORDER BY uic.COLUMN_POSITION) AS columns,
  LOWER(ui.INDEX_TYPE) AS type,
  CASE ui.UNIQUENESS WHEN 'UNIQUE' THEN 1 ELSE 0 END AS "unique",
  CASE
    WHEN ui.INDEX_NAME IN (
      SELECT ac.CONSTRAINT_NAME
      FROM ALL_CONSTRAINTS ac
      WHERE ac.OWNER = ui.OWNER
        AND ac.TABLE_NAME = ui.TABLE_NAME
        AND ac.CONSTRAINT_TYPE = 'P'
    ) THEN 1 ELSE 0
  END AS "primary"
FROM ALL_INDEXES ui
JOIN ALL_IND_COLUMNS uic
  ON uic.INDEX_OWNER = ui.OWNER AND uic.INDEX_NAME = ui.INDEX_NAME
WHERE ui.OWNER = %s
  AND ui.TABLE_NAME = %s
GROUP BY ui.OWNER, ui.TABLE_NAME, ui.INDEX_NAME, ui.INDEX_TYPE, ui.UNIQUENESS`,
		r.wrap.Quote(strings.ToUpper(schema)), r.wrap.Quote(strings.ToUpper(table)),
	), nil
}

func (r *Grammar) CompileJsonColumnsUpdate(values map[string]any) (map[string]any, error) {
	compiled := make(map[string]any, len(values))
	for key, value := range values {
		if strings.Contains(key, "->") {
			return nil, stderrors.New("dm grammar does not support json path update yet")
		}
		compiled[key] = value
	}
	return compiled, nil
}

func (r *Grammar) CompileJsonContains(column string, value any, isNot bool) (string, []any, error) {
	return "", nil, stderrors.New("dm grammar does not support json contains yet")
}

func (r *Grammar) CompileJsonContainsKey(column string, isNot bool) string {
	_ = column
	_ = isNot
	return "1 = 0"
}

func (r *Grammar) CompileJsonLength(column string) string {
	return "0"
}

func (r *Grammar) CompileJsonSelector(column string) string {
	segments := strings.Split(column, "->")
	return r.wrap.Column(segments[0])
}

func (r *Grammar) CompileJsonValues(args ...any) []any {
	for i, arg := range args {
		val := reflect.ValueOf(arg)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			args[i] = fmt.Sprint(val.Interface())
		case reflect.Slice, reflect.Array:
			if length := val.Len(); length > 0 {
				values := make([]any, length)
				for j := 0; j < length; j++ {
					values[j] = val.Index(j).Interface()
				}
				args[i] = r.CompileJsonValues(values...)
			}
		default:
		}
	}
	return args
}

func (r *Grammar) CompileLockForUpdate(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
	if conditions.LockForUpdate != nil && *conditions.LockForUpdate {
		builder = builder.Suffix("FOR UPDATE")
	}
	return builder
}
func (r *Grammar) CompileLockForUpdateForGorm() clause.Expression {
	return clause.Locking{Strength: "UPDATE"}
}
func (r *Grammar) CompilePlaceholderFormat() driver.PlaceholderFormat {
	return sq.Question
}
func (r *Grammar) CompilePrimary(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s add primary key (%s)", r.wrap.Table(blueprint.GetTableName()), r.wrap.Columnize(command.Columns))
}
func (r *Grammar) CompilePrune(_ string) string { return "" }
func (r *Grammar) CompileInRandomOrder(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
	if conditions.InRandomOrder != nil && *conditions.InRandomOrder {
		conditions.OrderBy = []string{"RAND()"}
	}
	return builder
}
func (r *Grammar) CompileRandomOrderForGorm() string { return "RAND()" }
func (r *Grammar) CompileRename(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s rename to %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Table(command.To))
}
func (r *Grammar) CompileRenameColumn(blueprint driver.Blueprint, command *driver.Command, _ []driver.Column) (string, error) {
	return fmt.Sprintf("alter table %s rename column %s to %s", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.From), r.wrap.Column(command.To)), nil
}
func (r *Grammar) CompileRenameIndex(_ driver.Blueprint, command *driver.Command, _ []driver.Index) []string {
	return []string{fmt.Sprintf("alter index %s rename to %s", r.wrap.Column(command.From), r.wrap.Column(command.To))}
}
func (r *Grammar) CompileSharedLock(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
	_ = conditions
	return builder
}
func (r *Grammar) CompileSharedLockForGorm() clause.Expression {
	return clause.Locking{Strength: "SHARE"}
}
func (r *Grammar) CompileTables(_ string) string {
	return `SELECT
  t.TABLE_NAME AS name,
  t.OWNER AS schema,
  0 AS size,
  tc.COMMENTS AS comment
FROM ALL_TABLES t
LEFT JOIN ALL_TAB_COMMENTS tc
  ON tc.OWNER = t.OWNER AND tc.TABLE_NAME = t.TABLE_NAME
WHERE t.OWNER NOT IN ('SYS', 'SYSTEM')
ORDER BY t.TABLE_NAME`
}
func (r *Grammar) CompileTableComment(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("comment on table %s is '%s'", r.wrap.Table(blueprint.GetTableName()), strings.ReplaceAll(command.Value, "'", "''"))
}
func (r *Grammar) CompileTypes() string {
	return `SELECT
  '' AS name,
  '' AS schema,
  '' AS type,
  '' AS category,
  1 AS implicit
FROM DUAL
WHERE 1 = 0`
}
func (r *Grammar) CompileUnique(blueprint driver.Blueprint, command *driver.Command) string {
	return fmt.Sprintf("alter table %s add constraint %s unique (%s)", r.wrap.Table(blueprint.GetTableName()), r.wrap.Column(command.Index), r.wrap.Columnize(command.Columns))
}
func (r *Grammar) CompileVersion() string {
	return "SELECT TO_CHAR(SYSTIMESTAMP, 'YYYYMMDDHH24MISS') AS value FROM DUAL"
}
func (r *Grammar) CompileViews(_ string) string {
	return `SELECT
  v.VIEW_NAME AS name,
  v.OWNER AS schema,
  v.TEXT AS definition
FROM ALL_VIEWS v
WHERE v.OWNER NOT IN ('SYS', 'SYSTEM')
ORDER BY v.VIEW_NAME`
}
func (r *Grammar) GetAttributeCommands() []string { return r.attributeCommands }
func (r *Grammar) EscapeNames(names []string) []string {
	escapedNames := make([]string, 0, len(names))
	for _, name := range names {
		segments := strings.Split(name, ".")
		for i, segment := range segments {
			segments[i] = strings.Trim(segment, `'"`)
		}
		escapedNames = append(escapedNames, `"`+strings.Join(segments, `"."`)+`"`)
	}
	return escapedNames
}
func (r *Grammar) ModifyDefault(_ driver.Blueprint, column driver.ColumnDefinition) string {
	if column.IsChange() {
		if column.GetAutoIncrement() || column.IsSetGeneratedAs() {
			return ""
		}
		if column.GetDefault() != nil {
			return fmt.Sprintf(" set default %s", schema.ColumnDefaultValue(column.GetDefault()))
		}
		return " drop default"
	}
	if column.GetDefault() != nil {
		return fmt.Sprintf(" default %s", schema.ColumnDefaultValue(column.GetDefault()))
	}
	return ""
}
func (r *Grammar) ModifyGeneratedAs(_ driver.Blueprint, column driver.ColumnDefinition) string {
	if !column.IsSetGeneratedAs() {
		return ""
	}
	option := "by default"
	if column.IsAlways() {
		option = "always"
	}
	identity := ""
	if generatedAs := column.GetGeneratedAs(); len(generatedAs) > 0 {
		identity = " (" + generatedAs + ")"
	}
	sql := fmt.Sprintf(" generated %s as identity%s", option, identity)
	if column.IsChange() {
		sql = " add" + sql
	}
	return sql
}
func (r *Grammar) ModifyGeneratedAsForChange(_ driver.Blueprint, column driver.ColumnDefinition) string {
	if column.IsChange() && column.IsSetGeneratedAs() && !column.GetAutoIncrement() {
		return " drop identity if exists"
	}
	return ""
}
func (r *Grammar) ModifyNullable(_ driver.Blueprint, column driver.ColumnDefinition) string {
	if column.IsChange() {
		if column.GetNullable() {
			return " drop not null"
		}
		return " set not null"
	}
	if column.GetNullable() {
		return " null"
	}
	return " not null"
}
func (r *Grammar) ModifyIncrement(blueprint driver.Blueprint, column driver.ColumnDefinition) string {
	if !column.IsChange() && !blueprint.HasCommand("primary") && (slices.Contains(r.serials, column.GetType()) || column.IsSetGeneratedAs()) && column.GetAutoIncrement() {
		return " primary key"
	}
	return ""
}
func (r *Grammar) TypeBigInteger(column driver.ColumnDefinition) string {
	if column.GetAutoIncrement() && !column.IsChange() && !column.IsSetGeneratedAs() {
		return "bigserial"
	}
	return "bigint"
}
func (r *Grammar) TypeBoolean(driver.ColumnDefinition) string { return "boolean" }
func (r *Grammar) TypeChar(column driver.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("char(%d)", length)
	}
	return "char"
}
func (r *Grammar) TypeDate(driver.ColumnDefinition) string { return "date" }
func (r *Grammar) TypeDateTime(column driver.ColumnDefinition) string {
	return r.TypeTimestamp(column)
}
func (r *Grammar) TypeDateTimeTz(column driver.ColumnDefinition) string {
	return r.TypeTimestampTz(column)
}
func (r *Grammar) TypeDecimal(column driver.ColumnDefinition) string {
	return fmt.Sprintf("decimal(%d, %d)", column.GetTotal(), column.GetPlaces())
}
func (r *Grammar) TypeDouble(driver.ColumnDefinition) string { return "double precision" }
func (r *Grammar) TypeEnum(column driver.ColumnDefinition) string {
	return fmt.Sprintf(`varchar(255) check ("%s" in (%s))`, column.GetName(), strings.Join(r.wrap.Quotes(cast.ToStringSlice(column.GetAllowed())), ", "))
}
func (r *Grammar) TypeFloat(column driver.ColumnDefinition) string {
	precision := column.GetPrecision()
	if precision > 0 {
		return fmt.Sprintf("float(%d)", precision)
	}
	return "float"
}
func (r *Grammar) TypeInteger(column driver.ColumnDefinition) string {
	if column.GetAutoIncrement() && !column.IsChange() && !column.IsSetGeneratedAs() {
		return "serial"
	}
	return "integer"
}
func (r *Grammar) TypeJson(driver.ColumnDefinition) string  { return "json" }
func (r *Grammar) TypeJsonb(driver.ColumnDefinition) string { return "json" }
func (r *Grammar) TypeLongText(driver.ColumnDefinition) string {
	return "text"
}
func (r *Grammar) TypeMediumInteger(column driver.ColumnDefinition) string {
	return r.TypeInteger(column)
}
func (r *Grammar) TypeMediumText(driver.ColumnDefinition) string { return "text" }
func (r *Grammar) TypeSmallInteger(column driver.ColumnDefinition) string {
	if column.GetAutoIncrement() && !column.IsChange() && !column.IsSetGeneratedAs() {
		return "smallserial"
	}
	return "smallint"
}
func (r *Grammar) TypeString(column driver.ColumnDefinition) string {
	length := column.GetLength()
	if length > 0 {
		return fmt.Sprintf("varchar(%d)", length)
	}
	return "varchar"
}
func (r *Grammar) TypeText(driver.ColumnDefinition) string { return "text" }
func (r *Grammar) TypeTime(column driver.ColumnDefinition) string {
	return fmt.Sprintf("time(%d) without time zone", column.GetPrecision())
}
func (r *Grammar) TypeTimeTz(column driver.ColumnDefinition) string {
	return fmt.Sprintf("time(%d) with time zone", column.GetPrecision())
}
func (r *Grammar) TypeTimestamp(column driver.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(schema.Expression("CURRENT_TIMESTAMP"))
	}
	return fmt.Sprintf("timestamp(%d) without time zone", column.GetPrecision())
}
func (r *Grammar) TypeTimestampTz(column driver.ColumnDefinition) string {
	if column.GetUseCurrent() {
		column.Default(schema.Expression("CURRENT_TIMESTAMP"))
	}
	return fmt.Sprintf("timestamp(%d) with time zone", column.GetPrecision())
}
func (r *Grammar) TypeTinyInteger(column driver.ColumnDefinition) string {
	return r.TypeSmallInteger(column)
}
func (r *Grammar) TypeTinyText(driver.ColumnDefinition) string { return "varchar(255)" }
func (r *Grammar) TypeUuid(driver.ColumnDefinition) string     { return "uuid" }

func (r *Grammar) getColumns(blueprint driver.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, r.getColumn(blueprint, column))
	}
	return columns
}

func (r *Grammar) getColumn(blueprint driver.Blueprint, column driver.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", r.wrap.Column(column.GetName()), schema.ColumnType(r, column))
	for _, modifier := range r.modifiers {
		sql += modifier(blueprint, column)
	}
	return sql
}

func parseSchemaAndTable(reference, defaultSchema string) (string, string, error) {
	if reference == "" {
		return "", "", errors.SchemaEmptyReferenceString
	}
	parts := strings.Split(reference, ".")
	if len(parts) > 2 {
		return "", "", errors.SchemaErrorReferenceFormat
	}
	schema := defaultSchema
	if len(parts) == 2 {
		schema = parts[0]
		parts = parts[1:]
	}
	table := parts[0]
	return schema, table, nil
}
