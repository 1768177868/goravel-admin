//go:build dm

package dm

import (
	"database/sql"
	"database/sql/driver"
	"dm"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strconv"
	"strings"
)

type multiRows struct {
	Val interface{}
}

func (a *multiRows) Value() (driver.Value, error) {
	return a.Val, nil
}

func Create(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	if db.Statement.Schema != nil && !db.Statement.Unscoped {
		for _, c := range db.Statement.Schema.CreateClauses {
			db.Statement.AddClause(c)
		}
	}

	var onConflict clause.OnConflict
	var hasConflict bool
	var rowCount int64
	if db.Statement.SQL.String() == "" {
		var (
			values = callbacks.ConvertToCreateValues(db.Statement)
			c      = db.Statement.Clauses["ON CONFLICT"]
		)
		onConflict, hasConflict = c.Expression.(clause.OnConflict)

		//mysql兼容处理
		if db.Statement.Config.Dialector.(*Dialector).GormMode == 1 {

			if db.Statement.Schema != nil {

				if field := db.Statement.Schema.PrioritizedPrimaryField; field != nil && field.AutoIncrement {
					setIdentityInsert := initSetIdentity(db, field)

					if setIdentityInsert && !db.DryRun && db.Error == nil {
						//尝试执行ON语句
						setIdentityInsert = setIdentityOn(db)

						//如果真实存在自增列，需要执行OFF语句
						if setIdentityInsert {
							defer setIdentityOff(db)
						}
					}
				}
			}

			MysqlUpsert(db, onConflict, values)
			rowCount = int64(len(values.Values))
			hasConflict = false
		} else {
			if hasConflict {
				if len(db.Statement.Schema.PrimaryFields) > 0 {
					columnsMap := map[string]bool{}
					for _, column := range values.Columns {
						columnsMap[column.Name] = true
					}

					for _, field := range db.Statement.Schema.PrimaryFields {
						if _, ok := columnsMap[field.DBName]; !ok {
							hasConflict = false
						}
					}
				} else {
					hasConflict = false
				}
			}

			if hasConflict {
				MergeCreate(db, onConflict, values)
			} else {
				if db.Statement.Schema != nil {
					if field := db.Statement.Schema.PrioritizedPrimaryField; field != nil && field.AutoIncrement {
						setIdentityInsert := initSetIdentity(db, field)

						if setIdentityInsert && !db.DryRun && db.Error == nil {
							//尝试执行ON语句
							setIdentityInsert = setIdentityOn(db)

							//如果真实存在自增列，需要执行OFF语句
							if setIdentityInsert {
								defer setIdentityOff(db)
							}
						}
					}
				}

				//拼写insert语句
				db.Statement.AddClauseIfNotExists(clause.Insert{})
				normalInsert(db, values)
			}
		}
	}

	if !db.DryRun && db.Error == nil {
		var (
			rows           *sql.Rows
			result         sql.Result
			err            error
			updateInsertID bool  // 是否需要更新主键自增列
			insertID       int64 // 主键自增列最新值
		)
		if hasConflict {
			rows, err = db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if db.AddError(err) != nil {
				return
			}
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&insertID)
				if insertID > 0 {
					updateInsertID = true
				}
			}
		} else {
			result, err = db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if db.AddError(err) != nil {
				return
			}
			if db.Statement.Config.Dialector.(*Dialector).GormMode == 1 {
				db.RowsAffected = rowCount
			} else {
				db.RowsAffected, _ = result.RowsAffected()
			}
			if db.RowsAffected != 0 && db.Statement.Schema != nil &&
				db.Statement.Schema.PrioritizedPrimaryField != nil &&
				db.Statement.Schema.PrioritizedPrimaryField.HasDefaultValue {
				insertID, err = result.LastInsertId()
				insertOk := err == nil && insertID > 0
				if !insertOk {
					db.AddError(err)
					return
				}
				updateInsertID = true
			}
		}

		if updateInsertID {
			switch db.Statement.ReflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				//if config.LastInsertIDReversed {
				for i := db.Statement.ReflectValue.Len() - 1; i >= 0; i-- {
					rv := db.Statement.ReflectValue.Index(i)
					if reflect.Indirect(rv).Kind() != reflect.Struct {
						break
					}

					_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv)
					if isZero {
						db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
						insertID -= db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
					}
				}
				//} else {
				//	for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				//		rv := db.Statement.ReflectValue.Index(i)
				//		if reflect.Indirect(rv).Kind() != reflect.Struct {
				//			break
				//		}
				//
				//		if _, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv); isZero {
				//			db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
				//			insertID += db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
				//		}
				//	}
				//}
			case reflect.Struct:
				_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if isZero {
					db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, db.Statement.ReflectValue, insertID))
				}
			}
		}
	}
}

func initSetIdentity(db *gorm.DB, field *schema.Field) bool {
	setIdentityInsert := false
	switch db.Statement.ReflectValue.Kind() {
	case reflect.Struct:
		_, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
		setIdentityInsert = !isZero
	case reflect.Slice, reflect.Array:
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			obj := db.Statement.ReflectValue.Index(i)
			if reflect.Indirect(obj).Kind() == reflect.Struct {
				_, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue.Index(i))
				setIdentityInsert = !isZero
			}
			break
		}
	}
	return setIdentityInsert
}

func setIdentityOn(db *gorm.DB) bool {
	db.Statement.SQL.Reset()
	db.Statement.WriteString("SET IDENTITY_INSERT ")
	db.Statement.WriteQuoted(db.Statement.Table)
	db.Statement.WriteString(" ON;")
	_, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
	if db.AddError(err) != nil {
		if err.(*dm.DmError).ErrCode == -2717 {
			//bug676758 表中不存在自增列，修改相关标识
			db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement = false
			db.Error = nil
			return false
		}
	}
	return true
}

func setIdentityOff(db *gorm.DB) {
	db.Statement.SQL.Reset()
	db.Statement.WriteString("SET IDENTITY_INSERT ")
	db.Statement.WriteQuoted(db.Statement.Table)
	db.Statement.WriteString(" OFF;")
	db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
}

/*
*
兼容mysql on duplicate key update语法
*/
func MysqlUpsert(db *gorm.DB, onConflict clause.OnConflict, values clause.Values) {
	rowCount := int64(len(values.Values))

	upsertFlag := true
	//todo 普通insert语句
	if !(onConflict.UpdateAll || len(onConflict.DoUpdates) > 0) {
		normalInsert(db, values)
		upsertFlag = false
	}

	//upsert语句块拼接
	if upsertFlag {
		db.Statement.SQL.Reset()
		var tableName = db.Statement.Table
		db.Statement.WriteString("/*DMGORM-UPSERT*/DECLARE ")
		for i := 0; i < int(rowCount); i++ {
			for _, columnName := range values.Columns {
				db.Statement.WriteString(appendVariableName(columnName.Name, strconv.Itoa(i)))
				db.Statement.WriteQuoted(tableName)
				db.Statement.WriteByte('.')
				db.Statement.WriteQuoted(columnName.Name)
				db.Statement.WriteString("%TYPE := ?")
				db.Statement.WriteByte(';')
			}
		}
		//定义变量,判断是否全部走insert逻辑,否则不回填id
		db.Statement.WriteString(" v_upsert_flag TINYINT := 1; ")
		db.Statement.WriteString(" TYPE emp_tab IS TABLE OF VARCHAR2; ")

		for i := 0; i < int(rowCount); i++ {
			rowIndex := strconv.Itoa(i)
			db.Statement.WriteString("v_emp_data" + rowIndex + " emp_tab; ")
			db.Statement.WriteString("v_sql_update" + rowIndex)
			db.Statement.WriteString(" VARCHAR2(32767) := 'DECLARE ")

			for _, columnName := range values.Columns {
				//内部sql变量，重命名，防止与上面变量重名
				db.Statement.WriteString(strings.ToUpper(columnName.Name) + "X ")
				db.Statement.WriteQuoted(tableName)
				db.Statement.WriteByte('.')
				db.Statement.WriteQuoted(columnName.Name)
				db.Statement.WriteString("%TYPE := :")
				db.Statement.WriteString(appendVariableName(columnName.Name, rowIndex))
				db.Statement.WriteString("; ")
			}
			db.Statement.WriteString("BEGIN UPDATE ")
			db.Statement.WriteQuoted(tableName)
			db.Statement.WriteString(" SET ")

			//更新所有的列
			if onConflict.UpdateAll {
				for idx, columnName := range values.Columns {
					if idx > 0 {
						db.Statement.WriteString(", ")
					}
					db.Statement.WriteQuoted(columnName.Name)
					db.Statement.WriteString(" = :")
					db.Statement.WriteString(appendVariableName(columnName.Name, rowIndex))
				}
			} else if len(onConflict.DoUpdates) > 0 {
				//更新部分列
				for idx, conflictName := range onConflict.DoUpdates {
					if idx > 0 {
						db.Statement.WriteString(", ")
					}
					db.Statement.WriteQuoted(conflictName.Column.Name)
					db.Statement.WriteString(" = :")
					db.Statement.WriteString(appendVariableName(conflictName.Column.Name, rowIndex))
				}
			} else {
				//todo 普通insert情况, 不应该进入这里, 报错
			}
			db.Statement.WriteString(` WHERE ';`)
			db.Statement.WriteString(` v_condition` + rowIndex + ` VARCHAR2(32767) := '';`)
			db.Statement.WriteString(` v_index_name` + rowIndex + ` VARCHAR2(32767);`)
		}

		db.Statement.WriteString(` BEGIN `)

		//多行upsert语句核心
		for i := 0; i < int(rowCount); i++ {
			db.Statement.WriteString(` BEGIN `)
			rowIndex := strconv.Itoa(i)
			db.Statement.WriteString(` EXECUTE IMMEDIATE 'INSERT INTO `)

			db.Statement.WriteQuoted(tableName)
			db.Statement.WriteString(" (")
			for idx, columnName := range values.Columns {
				if idx > 0 {
					db.Statement.WriteString(", ")
				}
				db.Statement.WriteQuoted(columnName.Name)
			}
			db.Statement.WriteString(") VALUES(")
			for idx, columnName := range values.Columns {
				if idx > 0 {
					db.Statement.WriteString(", ")
				}
				db.Statement.WriteByte(':')
				db.Statement.WriteString(appendVariableName(columnName.Name, rowIndex))
			}
			db.Statement.WriteString(");' USING ")
			for idx, columnName := range values.Columns {
				if idx > 0 {
					db.Statement.WriteString(", ")
				}
				db.Statement.WriteString(appendVariableName(columnName.Name, rowIndex))
			}
			db.Statement.WriteString(`; `)

			db.Statement.WriteString(`EXCEPTION WHEN DUP_VAL_ON_INDEX THEN `)
			db.Statement.WriteString("v_upsert_flag := 0;")
			db.Statement.WriteString(`v_index_name` + rowIndex + ` := (SELECT SUBSTR(SQLERRM, INSTR(SQLERRM, '[', -1, 1) + 1, INSTR(SQLERRM, ']', -1, 1) - INSTR(SQLERRM, '[', -1, 1) - 1));
		 SELECT COLS.NAME
		 BULK COLLECT INTO v_emp_data` + rowIndex)
			db.Statement.WriteString(` from 
		 SYSCOLUMNS COLS, 
         SYSCONS CONS, 
         SYSINDEXES INDS `)

			db.Statement.WriteString(`where CONS.ID = （select ID from SYSOBJECTS WHERE NAME = v_index_name` + rowIndex + ") ")
			db.Statement.WriteString(`and COLS.ID = (select ID from sysobjects WHERE SCHID = CURRENT_SCHID() AND name = `)
			db.Statement.WriteString("'" + tableName + "'")
			db.Statement.WriteString(`)
         AND INDS.ID = CONS.INDEXID AND SF_COL_IS_IDX_KEY(INDS.KEYNUM, INDS.KEYINFO, COLS.COLID)= 1
         union
         SELECT c.column_name AS NAME  FROM user_indexes i JOIN USER_IND_COLUMNS c 
         on i.index_name = c.index_name where i.index_name = v_index_name` + rowIndex + " ")
			db.Statement.WriteString(`and i.UNIQUENESS = 'UNIQUE' and i.STATUS = 'VALID' and i.TABLE_NAME = `)
			db.Statement.WriteString("'" + tableName + "'")
			db.Statement.WriteString(` AND I.TABLE_OWNER = (SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA') FROM DUAL);`)
			db.Statement.WriteString(`FOR i IN 1..v_emp_data` + rowIndex + `.COUNT LOOP
        	v_condition` + rowIndex + `:= v_condition` + rowIndex + ` || ' "' || v_emp_data` + rowIndex + `(i) ||'" = :'|| UPPER(v_emp_data` + rowIndex + `(i)) || '` + rowIndex + ` AND ';
         END LOOP;
         v_condition` + rowIndex + ` := v_condition` + rowIndex + ` || '1 = 1; END;' ;
         V_sql_update` + rowIndex + ` := ` + `v_sql_update` + rowIndex + ` || ' ' || v_condition` + rowIndex + `;
         EXECUTE IMMEDIATE v_sql_update` + rowIndex + ` using `)
			for idx, columnName := range values.Columns {
				if idx > 0 {
					db.Statement.WriteString(", ")
				}
				db.Statement.WriteString(appendVariableName(columnName.Name, rowIndex))
			}
			db.Statement.WriteString(`;
	WHEN OTHERS THEN
		RAISE;
	END;`)
		}

		//根据v_upsert_flag来决定是否回填id
		db.Statement.WriteString(`BEGIN
   IF v_upsert_flag = 1 THEN
   	SELECT IDENT_CURRENT('` + tableName + `') from dual;
   	END IF;
END;
`)

		db.Statement.WriteString("END;")

		BindVar(db.Statement, values.Values)
		//if len(values.Values) > 1 {
		//	//多行绑定
		//	var batchArgs = make([]interface{}, len(values.Columns))
		//	var lastArgs = make([][]interface{}, 0)
		//	for _, value1 := range values.Values {
		//		lastArgs = append(lastArgs, value1)
		//	}
		//	batchArgs[len(batchArgs)-1] = &multiRows{Val: lastArgs}
		//	BindVar(db.Statement, batchArgs...)
		//	//db.Statement.AddVar(db.Statement,batchArgs...)
		//} else {
		//	//单行绑定
		//	BindVar(db.Statement, values.Values)
		//}
	}
}

func normalInsert(db *gorm.DB, values clause.Values) {
	db.Statement.SQL.Reset()
	db.Statement.WriteString("INSERT INTO ")
	db.Statement.WriteQuoted(db.Statement.Table)
	db.Statement.AddClause(values)
	if values, ok := db.Statement.Clauses["VALUES"].Expression.(clause.Values); ok {
		if len(values.Columns) > 0 {
			db.Statement.WriteByte('(')
			for idx, column := range values.Columns {
				if idx > 0 {
					db.Statement.WriteByte(',')
				}
				db.Statement.WriteQuoted(column)
			}
			db.Statement.WriteByte(')')

			//outputInserted(db)

			db.Statement.WriteString(" VALUES ")

			for idx, value := range values.Values {
				if idx > 0 {
					db.Statement.WriteByte(',')
				}

				db.Statement.WriteByte('(')
				db.Statement.AddVar(db.Statement, value...)
				db.Statement.WriteByte(')')
			}

			db.Statement.WriteString(";")
		} else {
			db.Statement.WriteString("DEFAULT VALUES")
		}
	}
}

// 变量名需要在后缀+1, 防止服务器关键字对sql拼接的影响
func appendVariableName(columnName string, rowIndex string) string {
	return strings.ToUpper(columnName) + rowIndex
}

func MergeCreate(db *gorm.DB, onConflict clause.OnConflict, values clause.Values) {
	db.Statement.WriteString("MERGE INTO ")
	db.Statement.WriteQuoted(db.Statement.Table)
	db.Statement.WriteString(" USING (")
	for idx, value := range values.Values {
		if idx > 0 {
			db.Statement.WriteString(" UNION ALL ")
		}

		db.Statement.WriteString("SELECT ")
		db.Statement.AddVar(db.Statement, value...)
		db.Statement.WriteString(" FROM DUAL")
	}

	db.Statement.WriteString(") AS \"excluded\" (")
	for idx, column := range values.Columns {
		if idx > 0 {
			db.Statement.WriteByte(',')
		}
		db.Statement.WriteQuoted(column.Name)
	}
	db.Statement.WriteString(") ON ")

	var where clause.Where
	for _, field := range db.Statement.Schema.PrimaryFields {
		where.Exprs = append(where.Exprs, clause.Eq{
			Column: clause.Column{Table: db.Statement.Table, Name: field.DBName},
			Value:  clause.Column{Table: "excluded", Name: field.DBName},
		})
	}
	where.Build(db.Statement)

	if len(onConflict.DoUpdates) > 0 {
		// 将UPDATE子句中出现在关联条件中的列去除（即上面的ON子句），否则会报错：-4064:不能更新关联条件中的列
		var withoutOnColumns = make([]clause.Assignment, 0, len(onConflict.DoUpdates))
	a:
		for _, assignment := range onConflict.DoUpdates {
			for _, field := range db.Statement.Schema.PrimaryFields {
				if assignment.Column.Name == field.DBName {
					continue a
				}
			}
			withoutOnColumns = append(withoutOnColumns, assignment)
		}
		onConflict.DoUpdates = clause.Set(withoutOnColumns)
		if len(onConflict.DoUpdates) > 0 {
			db.Statement.WriteString(" WHEN MATCHED THEN UPDATE SET ")
			onConflict.DoUpdates.Build(db.Statement)
		}
	}

	db.Statement.WriteString(" WHEN NOT MATCHED THEN INSERT (")

	written := false
	for _, column := range values.Columns {
		if db.Statement.Schema.PrioritizedPrimaryField == nil || !db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement || db.Statement.Schema.PrioritizedPrimaryField.DBName != column.Name {
			if written {
				db.Statement.WriteByte(',')
			}
			written = true
			db.Statement.WriteQuoted(column.Name)
		}
	}

	db.Statement.WriteString(") VALUES (")

	written = false
	for _, column := range values.Columns {
		if db.Statement.Schema.PrioritizedPrimaryField == nil || !db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement || db.Statement.Schema.PrioritizedPrimaryField.DBName != column.Name {
			if written {
				db.Statement.WriteByte(',')
			}
			written = true
			db.Statement.WriteQuoted(clause.Column{
				Table: "excluded",
				Name:  column.Name,
			})
		}
	}

	db.Statement.WriteString(")")
	//outputInserted(db)
	db.Statement.WriteString(";")

	// merge into 语句插入的记录，无法通过LastInsertID获取
	if db.Statement.Schema.PrioritizedPrimaryField != nil && db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement {
		db.Statement.WriteString("SELECT ")
		db.Statement.WriteQuoted(db.Statement.Schema.PrioritizedPrimaryField.DBName)
		db.Statement.WriteString(" FROM ")
		db.Statement.WriteQuoted(db.Statement.Table)
		db.Statement.WriteString(" ORDER BY ")
		db.Statement.WriteQuoted(db.Statement.Schema.PrioritizedPrimaryField.DBName)
		db.Statement.WriteString(" DESC LIMIT 1;")
	}
}

func BindVar(stmt *gorm.Statement, vars ...interface{}) {
	for _, v := range vars {
		switch v := v.(type) {
		case sql.NamedArg:
			stmt.Vars = append(stmt.Vars, v.Value)
		case clause.Column, clause.Table:
			continue
		case gorm.Valuer:
			reflectValue := reflect.ValueOf(v)
			if reflectValue.Kind() == reflect.Ptr && reflectValue.IsNil() {
				BindVar(stmt, nil)
			} else {
				BindVar(stmt, v.GormValue(stmt.Context, stmt.DB))
			}
		case clause.Interface:
			c := clause.Clause{Name: v.Name()}
			v.MergeClause(&c)
			c.Build(stmt)
		case clause.Expression:
			v.Build(stmt)
		case driver.Valuer:
			stmt.Vars = append(stmt.Vars, v)
		case []byte:
			stmt.Vars = append(stmt.Vars, v)
		case []interface{}:
			if len(v) > 0 {
				BindVar(stmt, v...)
			} else {
				BindVar(stmt, nil)
			}
		default:
			switch rv := reflect.ValueOf(v); rv.Kind() {
			case reflect.Slice, reflect.Array:
				if rv.Len() == 0 {
					BindVar(stmt, nil)
				} else if rv.Type().Elem() == reflect.TypeOf(uint8(0)) {
					stmt.Vars = append(stmt.Vars, v)
				} else {
					for i := 0; i < rv.Len(); i++ {
						BindVar(stmt, rv.Index(i).Interface())
					}
				}
			default:
				stmt.Vars = append(stmt.Statement.Vars, v)
			}
		}
	}
}
