package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const roomImportUploadLimit = 2 << 20
const roomImportTextLimit = 128 << 10

func roomImportFile(r *http.Request) (string, error) {
	file, header, err := r.FormFile("spreadsheet")
	if errors.Is(err, http.ErrMissingFile) {
		return r.Form.Get("csv"), nil
	}
	if err != nil {
		return "", errors.New("Chọn tệp CSV hoặc XLSX để kiểm tra.")
	}
	defer file.Close()
	if header.Size > roomImportUploadLimit {
		return "", errors.New("Tệp nhập tối đa 2 MB.")
	}
	switch strings.ToLower(filepath.Ext(header.Filename)) {
	case ".csv":
		data, err := io.ReadAll(io.LimitReader(file, roomImportTextLimit+1))
		if err != nil || len(data) > roomImportTextLimit {
			return "", errors.New("Nội dung CSV quá dài hoặc không đọc được.")
		}
		return string(data), nil
	case ".xlsx":
		return roomWorkbookCSV(file)
	default:
		return "", errors.New("Dùng tệp .csv hoặc .xlsx theo mẫu.")
	}
}

func roomWorkbookCSV(reader io.Reader) (string, error) {
	const maxUnzipped = 16 << 20
	wb, err := excelize.OpenReader(io.LimitReader(reader, roomImportUploadLimit+1), excelize.Options{
		RawCellValue: true, UnzipSizeLimit: maxUnzipped, UnzipXMLSizeLimit: maxUnzipped,
	})
	if err != nil {
		return "", errors.New("Không đọc được tệp XLSX. Dùng mẫu và lưu lại dưới định dạng .xlsx.")
	}
	defer wb.Close()
	sheets := wb.GetSheetList()
	if len(sheets) != 1 {
		return "", errors.New("Tệp XLSX cần một trang tính theo mẫu.")
	}
	rows, err := wb.Rows(sheets[0])
	if err != nil {
		return "", errors.New("Không đọc được trang tính.")
	}
	defer rows.Close()
	var data bytes.Buffer
	output := csv.NewWriter(&data)
	readRows := 0
	for rows.Next() {
		readRows++
		if readRows > 151 {
			return "", errors.New("Tệp hỗ trợ tối đa 150 dòng phòng sau dòng tiêu đề.")
		}
		columns, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return "", fmt.Errorf("Dòng %d: không đọc được dữ liệu XLSX.", readRows)
		}
		if len(columns) > 7 {
			return "", fmt.Errorf("Dòng %d: cần đúng bảy cột theo mẫu.", readRows)
		}
		for i := 0; i < 7; i++ {
			cell := string(rune('A'+i)) + strconv.Itoa(readRows)
			formula, err := wb.GetCellFormula(sheets[0], cell)
			if err != nil || formula != "" {
				return "", fmt.Errorf("Ô %s: hãy dán giá trị thay cho công thức.", cell)
			}
		}
		if len(columns) == 0 && readRows != 1 {
			if err := output.Write(nil); err != nil {
				return "", err
			}
			continue
		}
		for len(columns) < 7 {
			columns = append(columns, "")
		}
		if err := output.Write(columns); err != nil {
			return "", err
		}
		output.Flush()
		if output.Error() != nil || data.Len() > roomImportTextLimit {
			return "", errors.New("Nội dung trang tính quá dài.")
		}
	}
	if rows.Error() != nil {
		return "", errors.New("Không đọc được toàn bộ trang tính.")
	}
	output.Flush()
	return data.String(), output.Error()
}

func roomImportWorkbook() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()
	columns := strings.Split(roomCSVHeader, ",")
	if err := wb.SetSheetRow(wb.GetSheetName(0), "A1", &columns); err != nil {
		return nil, err
	}
	return wb.WriteToBuffer()
}
