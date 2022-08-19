package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/sayuthisobri/waktu-solat/common"
	"github.com/sayuthisobri/waktu-solat/services"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	ctx := &common.Ctx{Config: &common.Config{}}
	ctx.LoadEnv()
	cfg := ctx.Config
	dir, _ := os.UserCacheDir()
	defaultDbPath := filepath.Join(dir, fmt.Sprintf("%s.db", filepath.Base(os.Args[0])))
	app := &cli.App{
		UseShortOptionHandling: true,
		DefaultCommand:         "get",
		Usage:                  "Retrieve prayer time",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Value:       cfg.IsDebug,
				Usage:       "enable debug logs",
				EnvVars:     []string{common.ENV_PREFIX + "DEBUG"},
				Destination: &cfg.IsDebug,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{},
				Value:       "cli",
				Usage:       "output mode [cli, alfred]",
				EnvVars:     []string{common.ENV_PREFIX + "MODE"},
				Destination: &cfg.Mode,
			},
			&cli.StringFlag{
				Name:        "db",
				Aliases:     []string{},
				Value:       defaultDbPath,
				Usage:       "path to `DB_FILE`",
				Destination: &cfg.DbPath,
			},
		},
		Before: func(context *cli.Context) error {
			if cfg.IsAlfred() && !cfg.IsDebug {
				log.SetOutput(io.Discard)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Retrieve prayer time",
				Action: handlePrayerTimes(ctx),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "zone",
						Aliases: []string{},
						//Value:   "WLY01",
						Usage: "Zone ID (received using `zone` command)",
					},
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{},
						Value:   "daily",
						Usage:   "Result mode (daily|weekly|monthly|yearly)",
					},
				},
			},
			{
				Name:   "zone",
				Usage:  "List all accepted zone",
				Action: handleZones(ctx),
			},
			{
				Name:      "set-zone",
				Usage:     "Set default zone id",
				Action:    setZone(ctx),
				ArgsUsage: "<zone-id>",
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func setZone(ctx *common.Ctx) cli.ActionFunc {
	check := func(zId string) error {
		if len(zId) != 0 {
			if services.GetZoneById(ctx, zId) != nil {
				services.SetUserConfig(ctx, "ZONE_ID", zId)
				return nil
			}
		} else {
			return fmt.Errorf("zone id argument is required")
		}
		return fmt.Errorf("zone with id [%s] not found", zId)
	}

	return func(cli *cli.Context) error {
		zId := cli.Args().First()
		return check(zId)
	}
}

func handleZones(ctx *common.Ctx) func(cli *cli.Context) error {
	return func(cli *cli.Context) error {
		states := services.GetZoneStates(ctx)
		if ctx.Config.IsAlfred() {
			zs := services.ZoneStates(states)
			res, _ := json.Marshal(zs.ToAlfredResponse())
			fmt.Print(string(res))
		} else {
			for _, state := range states {
				color.Blue("State: %s", state.Name)
				for _, zone := range state.Zones {
					color.White("%s - %s", color.CyanString(zone.ID), color.YellowString(zone.Locations))
				}
				fmt.Println()
			}
		}
		return nil
	}
}

func handlePrayerTimes(ctx *common.Ctx) func(cli *cli.Context) error {
	return func(cli *cli.Context) error {
		prayerTimes := services.GetPrayerTimes(ctx, cli.String("zone"), cli.String("mode"))
		if prayerTimes != nil {
			for _, pt := range prayerTimes {
				if ctx.Config.IsAlfred() {
					response := pt.ToAlfredResponse()
					jsonResponse, _ := json.Marshal(response)
					fmt.Print(string(jsonResponse))
					return nil
				} else {
					color.Blue("Date\t\t: %s %s", pt.Date, color.MagentaString(pt.Hijri))
					color.Blue("Locations\t: %s", pt.Zone.Locations)
					for _, t := range pt.Times {
						var desc string
						if t.IsCurrent {
							desc = color.RedString("*Current")
						} else if d := t.Time.Sub(time.Now()); d > 0 {
							desc = color.WhiteString("%s", common.Timespan(d).Format())
						}
						color.White("%s\t: %s %s", color.CyanString(t.Key), color.YellowString(t.DisplayValue), desc)
					}
				}
			}
		}
		return nil
	}
}
