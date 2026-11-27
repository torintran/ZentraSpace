package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ImportPreview struct {
	Period, Data, SiteID, SiteName string
	Add, Skip                      int
}

const roomCSVHeader = "phong,nguoi_dai_dien,tien_thue,gia_dien,gia_nuoc,dien_dau,nuoc_dau"

func (b *Book) importRooms(period, siteID, data string, commit bool) (ImportPreview, error) {
	p := ImportPreview{Period: period, Data: data, SiteID: siteID}
	if len(data) > roomImportTextLimit || !utf8.ValidString(data) {
		return p, fmt.Errorf("Dùng nội dung CSV UTF-8 hợp lệ, tối đa 128 KB.")
	}
	var capacity, existing int
	if siteID != "" {
		for _, site := range b.Sites {
			if site.ID == siteID {
				capacity, p.SiteName = site.Capacity, site.Name
				break
			}
		}
		if capacity == 0 {
			return p, errMissing
		}
		for _, room := range b.Rooms {
			if room.SiteID == siteID {
				existing++
			}
		}
	}
	source := csv.NewReader(strings.NewReader(strings.TrimPrefix(data, "\ufeff")))
	source.FieldsPerRecord = 7
	source.TrimLeadingSpace = true
	header, err := source.Read()
	if err != nil || strings.Join(header, ",") != roomCSVHeader {
		return p, fmt.Errorf("Dòng đầu cần theo đúng mẫu CSV.")
	}
	var additions []Room
	for count := 1; ; count++ {
		row, err := source.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if parse, ok := err.(*csv.ParseError); ok {
				return p, fmt.Errorf("Dòng %d: CSV thiếu hoặc thừa cột.", parse.StartLine)
			}
			return p, fmt.Errorf("Không đọc được một dòng CSV.")
		}
		line, _ := source.FieldPos(0)
		if count > 150 {
			return p, fmt.Errorf("Một lần nhập tối đa 150 phòng.")
		}
		for j := range row {
			row[j] = strings.TrimSpace(row[j])
		}
		room := Room{Name: row[0], Tenant: row[1], Start: period, SiteID: siteID, Prepared: row[1] == ""}
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
				if saved.SiteID != room.SiteID || !strings.EqualFold(saved.Name, room.Name) {
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
	if siteID != "" && existing+len(additions) > capacity {
		return p, fmt.Errorf("Dãy trọ còn chỗ cho %d phòng. Danh sách cần thêm %d phòng.", max(0, capacity-existing), len(additions))
	}
	if commit {
		b.Rooms = append(b.Rooms, additions...)
	}
	p.Add = len(additions)
	return p, nil
}
