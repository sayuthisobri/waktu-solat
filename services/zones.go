package services

import (
	"github.com/gocolly/colly"
	"github.com/sayuthisobri/waktu-solat/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"strings"
	"time"
)

type ZoneStates []State

func (zs *ZoneStates) ToAlfredResponse() common.AlfredResponse {
	states := []State(*zs)
	var items []common.AlfredResponseItem
	for _, s := range states {
		for _, zone := range s.Zones {
			items = append(items, common.AlfredResponseItem{
				Title:    zone.ID,
				Subtitle: &zone.Locations,
				Arg:      zone.ID,
			})
		}
	}
	return common.AlfredResponse{Items: items}
}

type State struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string
	Zones     []Zone
}

type Zone struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Locations string
	StateID   string
	State     *State
}

func GetZoneById(ctx *common.Ctx, id string) *Zone {
	db, _ := OpenDb(ctx)

	zone := &Zone{ID: id}
	if db.First(&zone).Error == nil {
		return zone
	}

	return nil
}

func GetZoneStates(ctx *common.Ctx) []State {
	var states []State
	db, _ := OpenDb(ctx)
	if db != nil {
		_ = db.Model(&State{}).Preload("Zones").Find(&states)
	}
	if len(states) == 0 {
		states = fetchZones(ctx)
	}
	return states
}

func fetchZones(ctx *common.Ctx) []State {
	var states []State
	c := colly.NewCollector()

	c.OnHTML("select#inputZone:first-child", func(p *colly.HTMLElement) {
		db, _ := OpenDb(ctx)
		p.ForEach("optgroup", func(_ int, eState *colly.HTMLElement) {
			state := &State{
				Name: eState.Attr("label"),
			}
			eState.ForEach("option", func(i int, eZone *colly.HTMLElement) {
				id := eZone.Attr("value")
				if i == 0 {
					state.ID = id[:3]
				}
				zone := &Zone{
					ID:        id,
					Locations: processLocationName(eZone.Text),
					State:     state,
				}
				state.Zones = append(state.Zones, *zone)
				//zones = append(zones, *zone)
			})
			states = append(states, *state)
		})
		updateRecords(&states, db)
	})
	_ = c.Visit(`https://www.e-solat.gov.my/index.php?siteId=24&pageId=24`)
	return states
}

func getZone(ctx *common.Ctx, zoneId string) *Zone {
	db, _ := OpenDb(ctx)
	zone := &Zone{ID: zoneId}
	tx := db.First(zone)
	if tx.Error != nil {
		zoneCount := int64(0)
		db.Model(&Zone{}).Count(&zoneCount)
		if zoneCount < 1 {
			states := GetZoneStates(ctx)
			for _, s := range states {
				for _, z := range s.Zones {
					if z.ID == zoneId {
						return &z
					}
				}
			}
		}
		return nil
	}
	return zone
}

func updateRecords(states *[]State, db *gorm.DB) {
	if db != nil {
		err := db.
			Session(&gorm.Session{FullSaveAssociations: true}).
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(states).Error
		if err != nil {
			log.Panicln(err)
		}
	}
}

func processLocationName(locations string) string {
	return strings.ReplaceAll(locations[8:], " dan ", ", ")
}
