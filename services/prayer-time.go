package services

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/sayuthisobri/waktu-solat/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	URL               = "https://www.e-solat.gov.my/index.php?r=esolatApi/takwimsolat&period=year&zone=%s"
	PrimaryDateLayout = "02/01/2006"
	//DisplayTimeLayout = "03:04PM"
)

type PrayTime struct {
	Key          string
	Time         time.Time
	DisplayValue string
	Duration     time.Duration
	IsCurrent    bool
}

type PrayerDate struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Hijri     string         `json:"hijri" ptMode:"-"`
	Date      string         `json:"date" fromFormat:"02-Jan-2006" toFormat:"02/01/2006" ptMode:"-"`
	Imsak     string         `json:"imsak" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Subuh     string         `json:"fajr" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Syuruk    string         `json:"syuruk" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Zohor     string         `json:"dhuhr" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Asar      string         `json:"asr" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Maghrib   string         `json:"maghrib" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	Isyak     string         `json:"isha" fromFormat:"15:04:05" toFormat:"03:04PM" ptMode:"1"`
	ZoneID    string         `ptMode:"-"`
	Zone      *Zone

	Times []PrayTime `gorm:"-:all"`
}

func (p *PrayerDate) UnmarshalJSON(bytes []byte) error {
	var tmp map[string]string
	_ = json.Unmarshal(bytes, &tmp)

	convertTime := func(t reflect.StructTag, v string) (time.Time, error) {
		f, _ := common.FindTagValue(t, "fromFormat")
		if f != "" {
			return time.ParseInLocation(f, v, time.Local)
		}
		return time.Time{}, fmt.Errorf("date format not found")
	}

	v := reflect.ValueOf(p).Elem()
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		tag := typeField.Tag
		key, _ := common.FindTagValue(tag, "json")
		if s, ok := tmp[key]; ok {
			field := v.Field(i)
			switch field.Type().String() {
			case "string":
				field.Set(reflect.ValueOf(common.ConvertFormatBasedOnTag(tag, s)))
			case "time.Time":
				timeValue, err := convertTime(tag, s)
				if err == nil {
					field.Set(reflect.ValueOf(timeValue))
				}
			}
		}
	}
	return nil
}

func (p *PrayerDate) ToAlfredResponse() common.AlfredResponse {
	var items []common.AlfredResponseItem
	var vars = make(map[string]string)
	vars["Locations"] = p.Zone.Locations
	for _, pt := range p.Times {
		val := pt.DisplayValue
		if pt.IsCurrent {
			val = fmt.Sprintf("%s | Current", pt.DisplayValue)
		} else if pt.Duration > 0 {
			val = fmt.Sprintf("%s | In %s", pt.DisplayValue, common.Timespan(pt.Time.Sub(time.Now()).Round(time.Second)).Format())
		}
		item := common.AlfredResponseItem{
			Title:    pt.Key,
			Subtitle: &val,
			Valid:    true,
			//Variables: map[string]string{
			//	"current": strconv.FormatBool(pt.IsCurrent),
			//},
		}
		items = append(items, item)

	}
	return common.AlfredResponse{
		//Rerun:     1,
		Variables: vars,
		Items:     items,
	}
}

func (p *PrayerDate) init() {
	rp := reflect.ValueOf(p).Elem()
	dateField := rp.FieldByName("Date")
	dateStr := dateField.String()
	currentIndex := -1
	for i := rp.NumField() - 1; i > -1; i-- {
		typeField := rp.Type().Field(i)
		tag := typeField.Tag
		tagValue, _ := common.FindTagValue(tag, "ptMode")
		field := rp.Field(i)
		timeStr := field.String()
		key := typeField.Name
		if tagValue == "1" {
			pTime, _ := time.ParseInLocation(fmt.Sprintf("%s 03:04PM", PrimaryDateLayout), fmt.Sprintf("%s %s", dateStr, timeStr), time.Local)
			duration := pTime.Sub(time.Now()).Round(time.Second)
			if duration < 0 && currentIndex == -1 {
				currentIndex = i
			}
			p.Times = append(p.Times, PrayTime{
				Key:          key,
				Time:         pTime,
				DisplayValue: timeStr,
				Duration:     duration,
				IsCurrent:    currentIndex == i,
			})
		}
	}
	common.Reverse(p.Times)

}

type PrayerTimesDto struct {
	PrayerTimes []PrayerDate `json:"prayerTime"`
}

func GetPrayerTimes(ctx *common.Ctx, zoneId string, mode string) []PrayerDate {
	if len(zoneId) == 0 {
		zoneId = GetUserConfig(ctx, "ZONE_ID", "WLY01")
	}
	zoneId = strings.ToUpper(zoneId)
	db, _ := OpenDb(ctx)
	prayerTime := &PrayerDate{}
	zone := getZone(ctx, zoneId)
	switch mode {
	case "daily":
		tx := db.Joins("Zone").First(prayerTime, "zone_id = ? AND date = ?",
			zoneId, time.Now().Format(PrimaryDateLayout))
		if tx.Error == nil {
			prayerTime.init()
			return []PrayerDate{*prayerTime}
		}
	}
	resDto := fetchData(zoneId, zone, db)
	var res []PrayerDate
	for _, p := range resDto.PrayerTimes {
		if p.Date == time.Now().Format(PrimaryDateLayout) {
			p.init()
			res = append(res, p)
		}
	}
	return res
}

func fetchData(zoneId string, zone *Zone, db *gorm.DB) *PrayerTimesDto {
	c := colly.NewCollector()
	var resDto = &PrayerTimesDto{}
	c.OnResponse(func(r *colly.Response) {
		_ = json.Unmarshal(r.Body, resDto)
		for i := range resDto.PrayerTimes {
			t := &resDto.PrayerTimes[i]
			handleDateToId := func(from string) string {
				t, err := time.ParseInLocation("02/01/2006", from, time.Local)
				if err != nil {
					return strings.ReplaceAll(from, "/", "")
				}
				return t.Format("20060102")
			}
			t.ID = fmt.Sprintf("%s-%s", handleDateToId(t.Date), zoneId)
			t.ZoneID = zoneId
			t.Zone = zone
		}
		updatePrayerTime(&resDto.PrayerTimes, db)
	})
	_ = c.Visit(fmt.Sprintf(URL, strings.ToUpper(zoneId)))
	return resDto
}

//func parseTime(date string, timeStr string) time.Time {
//	t, _ := time.ParseInLocation("_2-Jan-2006 15:04:05", fmt.Sprintf("%s %s", date, timeStr), time.Local)
//	return t
//}

func updatePrayerTime(prayerTimes *[]PrayerDate, db *gorm.DB) {
	if db != nil {
		err := db.
			Session(&gorm.Session{FullSaveAssociations: true}).
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(prayerTimes).Error
		if err != nil {
			log.Panicln(err)
		}
	}
}
