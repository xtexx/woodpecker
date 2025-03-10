// Copyright 2021 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"fmt"
	"regexp"
	"strings"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func renameTable(sess *xorm.Session, oldTable, newTable string) error {
	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.MYSQL:
		_, err := sess.Exec(fmt.Sprintf("RENAME TABLE `%s` TO `%s`;", oldTable, newTable))
		return err
	case schemas.POSTGRES, schemas.SQLITE:
		_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`;", oldTable, newTable))
		return err
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}
}

// WARNING: YOU MUST COMMIT THE SESSION AT THE END.
func dropTableColumns(sess *xorm.Session, tableName string, columnNames ...string) (err error) {
	// Copyright 2017 The Gitea Authors. All rights reserved.
	// Use of this source code is governed by a MIT-style
	// license that can be found in the LICENSE file.

	if tableName == "" || len(columnNames) == 0 {
		return nil
	}
	// TODO: This will not work if there are foreign keys

	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.SQLITE:
		// First drop the indexes on the columns
		res, errIndex := sess.Query(fmt.Sprintf("PRAGMA index_list(`%s`)", tableName))
		if errIndex != nil {
			return errIndex
		}
		for _, row := range res {
			indexName := row["name"]
			indexRes, err := sess.Query(fmt.Sprintf("PRAGMA index_info(`%s`)", indexName))
			if err != nil {
				return err
			}
			if len(indexRes) != 1 {
				continue
			}
			indexColumn := string(indexRes[0]["name"])
			for _, name := range columnNames {
				if name == indexColumn {
					_, err := sess.Exec(fmt.Sprintf("DROP INDEX `%s`", indexName))
					if err != nil {
						return err
					}
				}
			}
		}

		// Now drop the columns
		for _, columnName := range columnNames {
			_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;", tableName, columnName))
			if err != nil {
				return fmt.Errorf("table `%s`, drop column %v: %w", tableName, columnName, err)
			}
		}

	case schemas.POSTGRES:
		cols := ""
		for _, col := range columnNames {
			if cols != "" {
				cols += ", "
			}
			cols += "DROP COLUMN `" + col + "` CASCADE"
		}
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` %s", tableName, cols)); err != nil {
			return fmt.Errorf("table `%s`, drop columns %v: %w", tableName, columnNames, err)
		}
	case schemas.MYSQL:
		// Drop indexes on columns first
		sql := fmt.Sprintf("SHOW INDEX FROM %s WHERE column_name IN ('%s')", tableName, strings.Join(columnNames, "','"))
		res, err := sess.Query(sql)
		if err != nil {
			return err
		}
		for _, index := range res {
			indexName := index["column_name"]
			if len(indexName) > 0 {
				_, err := sess.Exec(fmt.Sprintf("DROP INDEX `%s` ON `%s`", indexName, tableName))
				if err != nil {
					return err
				}
			}
		}

		// Now drop the columns
		cols := ""
		for _, col := range columnNames {
			if cols != "" {
				cols += ", "
			}
			cols += "DROP COLUMN `" + col + "`"
		}
		if _, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` %s", tableName, cols)); err != nil {
			return fmt.Errorf("table `%s`, drop columns %v: %w", tableName, columnNames, err)
		}
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}

	return nil
}

func alterColumnDefault(sess *xorm.Session, table, column, defValue string) error {
	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.MYSQL:
		sql := fmt.Sprintf("SHOW COLUMNS FROM `%s` WHERE lower(field) = '%s'", table, strings.ToLower(column))
		res, err := sess.Query(sql)
		if err != nil {
			return err
		}

		if len(res) == 0 || len(res[0]["Type"]) == 0 {
			return fmt.Errorf("column %s data type in table %s can not be detected", column, table)
		}

		dataType := string(res[0]["Type"])
		var nullable string
		if string(res[0]["Null"]) == "NO" {
			nullable = "NOT NULL"
		}

		_, err = sess.Exec(fmt.Sprintf("ALTER TABLE `%s` MODIFY `%s` %s %s DEFAULT %s;", table, column, dataType, nullable, defValue))
		return err
	case schemas.POSTGRES:
		_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` ALTER COLUMN `%s` SET DEFAULT %s;", table, column, defValue))
		return err
	case schemas.SQLITE:
		return nil
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}
}

func alterColumnNull(sess *xorm.Session, table, column string, null bool) error {
	val := "NULL"
	if !null {
		val = "NOT NULL"
	}
	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.MYSQL:
		sql := fmt.Sprintf("SHOW COLUMNS FROM `%s` WHERE lower(field) = '%s'", table, strings.ToLower(column))
		res, err := sess.Query(sql)
		if err != nil {
			return err
		}

		if len(res) == 0 || len(res[0]["Type"]) == 0 {
			return fmt.Errorf("column %s data type in table %s can not be detected", column, table)
		}

		dataType := string(res[0]["Type"])
		defValue := string(res[0]["Default"])

		if defValue != "NULL" && defValue != "" {
			defValue = fmt.Sprintf("DEFAULT '%s'", defValue)
		} else {
			defValue = ""
		}

		_, err = sess.Exec(fmt.Sprintf("ALTER TABLE `%s` MODIFY `%s` %s %s %s;", table, column, dataType, val, defValue))
		return err
	case schemas.POSTGRES:
		_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` ALTER COLUMN `%s` SET %s;", table, column, val))
		return err
	case schemas.SQLITE:
		return nil
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}
}

func renameColumn(sess *xorm.Session, table, column, newName string) error {
	dialect := sess.Engine().Dialect().URI().DBType
	switch dialect {
	case schemas.MYSQL,
		schemas.POSTGRES,
		schemas.SQLITE:
		_, err := sess.Exec(fmt.Sprintf("ALTER TABLE `%s` RENAME COLUMN `%s` TO `%s`;", table, column, newName))
		return err
	default:
		return fmt.Errorf("dialect '%s' not supported", dialect)
	}
}

var (
	whitespaces     = regexp.MustCompile(`\s+`)
	columnSeparator = regexp.MustCompile(`\s?,\s?`)
)

func removeColumnFromSQLITETableSchema(schema string, names ...string) string {
	if len(names) == 0 {
		return schema
	}
	for i := range names {
		if len(names[i]) == 0 {
			continue
		}
		schema = regexp.MustCompile(`\s(`+
			regexp.QuoteMeta("`"+names[i]+"`")+
			"|"+
			regexp.QuoteMeta(names[i])+
			")[^`,)]*?[,)]").ReplaceAllString(schema, "")
	}
	return schema
}

func normalizeSQLiteTableSchema(schema string) string {
	return columnSeparator.ReplaceAllString(
		whitespaces.ReplaceAllString(
			strings.ReplaceAll(schema, "\n", " "),
			" "),
		", ")
}
