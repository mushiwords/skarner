# skarner
Package to scan data(ej: sql.Rows) into structs

func describeColumnDatas(db *sql.DB, schema string, table *TableOptions) error {
	var fields []string
	for _, col := range table.Columns {
		col.Name = "`" + col.Name + "`"
		fields = append(fields, col.Name)
	}
	var fieldsStr = strings.Join(fields, ",")
	cmd := `select ` + fieldsStr + ` from ` + schema + `.` + table.Name + ` limit 5`

	fmt.Println(cmd)
	rows, err := db.Query(cmd)
	if err != nil {
		return  err
	}

	cols, _ := rows.Columns()
	columns1 := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns1 {
		columnPointers[i] = &columns1[i]
	}
	for rows.Next() {
		//record := make(map[string]interface{})
		err = rows.Scan(columnPointers...)
		if err != nil {
			return  err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			//record[colName] = *val // colName verify
			value := Strval(*val)

			fmt.Println(value)
			if table.Columns[i].Name == colName {
				table.Columns[i].Data = append(table.Columns[i].Data,value)
			}
		}
	}

	return  nil
}

//
    columns := make([]interface{}, len(columnNames))
	columnPointers := make([]interface{}, len(columnNames))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	if scanError := rows.Scan(columnPointers...); scanError != nil {
		return nil, scanError
	}

	record := make(map[string]interface{})
	for i, colName := range columnNames {
		val := columnPointers[i].(*interface{})
		record[colName] = *val
	}
