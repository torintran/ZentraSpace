package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ImportPreview struct {
	Period, Data string
	Add, Skip    int
}

const roomCSVHeader = "phong,nguoi_dai_dien,tien_thue,gia_dien,gia_nuoc,dien_dau,nuoc_dau"

func (b *Book) importRooms(period, data string, commit bool) (ImportPreview, error) {
	p := ImportPreview{Period: period, Data: data}
	source := csv.NewReader(strings.NewReader(strings.TrimPrefix(data, "\ufeff")))
	source.FieldsPerRecord = 7
	source.TrimLeadingSpace = true
	header, err := source.Read()
	if err != nil || strings.Join(header, ",") != roomCSVHeader {
		return p, fmt.Errorf("Dòng đầu cần theo đúng mẫu CSV.")
	}
	var additions []Room
	for line := 2; ; line++ {
		row, err := source.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return p, fmt.Errorf("Dòng %d: CSV thiếu hoặc thừa cột.", line)
		}
		if line > 151 {
			return p, fmt.Errorf("Một lần nhập tối đa 150 phòng.")
		}
		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}
		room := Room{Name: row[0], Tenant: row[1], Start: period}
		if room.Name == "" || len(room.Name) > 60 || len(room.Tenant) > 150 {
			return p, fmt.Errorf("Dòng %d: kiểm tra mã phòng và tên người đại diện.", line)
		}
		limits := []int64{100000000, 100000, 1000000, 999999, 999999}
		fields := []*int64{&room.Rent, &room.ElectricRate, &room.WaterRate, &room.Electric, &room.Water}
		for j, field := range fields {
			val, err := strconv.ParseInt(row[j+2], 10, 64)
			if err != nil || val < 0 || val > limits[j] {
				return p, fmt.Errorf("Dòng %d: cột %d cần số nguyên hợp lệ.", line, j+3)
			}
			*field = val
		}
		same := false
		for _, rooms := range [][]Room{b.Rooms, additions} {
			for _, saved := range rooms {
				if !strings.EqualFold(saved.Name, room.Name) {
					continue
				}
				room.ID = saved.ID
				room.LeaseMonths = saved.LeaseMonths
				if room != saved {
					return p, fmt.Errorf("Dòng %d: mã phòng đã có với thông tin khác.", line)
				}
				same = true
				break
			}
		}
		if same {
			p.Skip++
			continue
		}
		room.ID = token()
		additions = append(additions, room)
	}
	if len(additions)+len(b.Rooms) > 150 {
		return p, fmt.Errorf("Bản thử hỗ trợ tối đa 150 phòng.")
	}
	if commit {
		b.Rooms = append(b.Rooms, additions...)
	}
	p.Add = len(additions)
	return p, nil
}
