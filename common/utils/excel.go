package utils

/**
 * @Description
 * @Author ggr
 * @Date 2021/7/8 21:06
 * @Form https://cloud.tencent.com/developer/article/1846410
 **/

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
)

// maxCharCount 最多26个字符A-Z
const maxCharCount = 26

// ExportExcel 导出Excel文件
// sheetName 工作表名称, 注意这里不要取sheet1这种名字,否则导致文件打开时发生部分错误。
// headers 列名切片， 表头
// rows 数据切片，是一个二维数组
// f 为了创建多个单元格，得把标签往外放
func ExportExcel(sheetName string, headers []string, rows [][]interface{}, f *excelize.File) (*excelize.File, error) {
	sheetIndex, err := f.GetSheetIndex(sheetName)
	if sheetIndex == -1 {
		sheetIndex, err = f.NewSheet(sheetName)
	}
	if err != nil {
		return nil, err
	}
	maxColumnRowNameLen := 1 + len(strconv.Itoa(len(rows)))
	columnCount := len(headers)
	if columnCount > maxCharCount {
		maxColumnRowNameLen++
	} else if columnCount > maxCharCount*maxCharCount {
		maxColumnRowNameLen += 2
	}
	columnNames := make([][]byte, 0, columnCount)
	for i, header := range headers {
		columnName := getColumnName(i, maxColumnRowNameLen)
		columnNames = append(columnNames, columnName)
		// 初始化excel表头，这里的index从1开始要注意
		curColumnName := getColumnRowName(columnName, 1)
		err := f.SetCellValue(sheetName, curColumnName, header)
		if err != nil {
			return nil, err
		}
	}
	for rowIndex, row := range rows {
		for columnIndex, columnName := range columnNames {
			// 从第二行开始
			err := f.SetCellValue(sheetName, getColumnRowName(columnName, rowIndex+2), row[columnIndex])

			if err != nil {
				return nil, err
			}
		}
	}
	f.SetActiveSheet(sheetIndex)
	return f, nil
}

// getColumnName 生成列名
// Excel的列名规则是从A-Z往后排;超过Z以后用两个字母表示，比如AA,AB,AC;两个字母不够以后用三个字母表示，比如AAA,AAB,AAC
// 这里做数字到列名的映射：0 -> A, 1 -> B, 2 -> C
// maxColumnRowNameLen 表示名称框的最大长度，假设数据是10行，1000列，则最后一个名称框是J1000(如果有表头，则是J1001),是4位
// 这里根据 maxColumnRowNameLen 生成切片，后面生成名称框的时候可以复用这个切片，而无需扩容
func getColumnName(column, maxColumnRowNameLen int) []byte {
	const A = 'A'
	if column < maxCharCount {
		// 第一次就分配好切片的容量
		slice := make([]byte, 0, maxColumnRowNameLen)
		return append(slice, byte(A+column))
	} else {
		// 递归生成类似AA,AB,AAA,AAB这种形式的列名
		return append(getColumnName(column/maxCharCount-1, maxColumnRowNameLen), byte(A+column%maxCharCount))
	}
}

// getColumnRowName 生成名称框
// Excel的名称框是用A1,A2,B1,B2来表示的，这里需要传入前一步生成的列名切片，然后直接加上行索引来生成名称框，就无需每次分配内存
func getColumnRowName(columnName []byte, rowIndex int) (columnRowName string) {
	l := len(columnName)
	columnName = strconv.AppendInt(columnName, int64(rowIndex), 10)
	columnRowName = string(columnName)
	// 将列名恢复回去
	columnName = columnName[:l]
	return
}

func StreamWriterFunc(contents [][]string) {
	//打开工作簿
	file, err := excelize.OpenFile("Book1.xlsx")
	if err != nil {
		return
	}
	sheet_name := "Sheet1"
	//获取流式写入器
	streamWriter, _ := file.NewStreamWriter(sheet_name)
	if err != nil {
		fmt.Println(err)
	}

	rows, _ := file.GetRows(sheet_name) //获取行内容
	cols, _ := file.GetCols(sheet_name) //获取列内容
	fmt.Println("行数rows:  ", len(rows), "列数cols:  ", len(cols))

	//将源文件内容先写入excel
	for rowid, row_pre := range rows {
		row_p := make([]interface{}, len(cols))
		for colID_p := 0; colID_p < len(cols); colID_p++ {
			//fmt.Println(row_pre)
			//fmt.Println(colID_p)
			if row_pre == nil {
				row_p[colID_p] = nil
			} else {
				row_p[colID_p] = row_pre[colID_p]
			}
		}
		cell_pre, _ := excelize.CoordinatesToCellName(1, rowid+1)
		if err := streamWriter.SetRow(cell_pre, row_p); err != nil {
			fmt.Println(err)
		}
	}

	//将新加contents写进流式写入器
	for rowID := 0; rowID < len(contents); rowID++ {
		row := make([]interface{}, len(contents[0]))
		for colID := 0; colID < len(contents[0]); colID++ {
			row[colID] = contents[rowID][colID]
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowID+len(rows)+1) //决定写入的位置
		if err := streamWriter.SetRow(cell, row); err != nil {
			fmt.Println(err)
		}
	}

	//结束流式写入过程
	if err := streamWriter.Flush(); err != nil {
		fmt.Println(err)
	}
	//保存工作簿
	if err := file.SaveAs("Book1.xlsx"); err != nil {
		fmt.Println(err)
	}
}
