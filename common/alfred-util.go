package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type AlfredResult interface {
	ToAlfredResponse() AlfredResponse
}

type AlfredResponse struct {
	Variables map[string]string    `json:"variables,omitempty"`
	Rerun     float64              `json:"rerun,omitempty"`
	Items     []AlfredResponseItem `json:"items"`
}

func (r *AlfredResponse) AddItem(item AlfredResponseItem) {
	r.Items = append(r.Items, item)
}

func (r *AlfredResponse) Print() {
	bytes, _ := json.Marshal(r)
	fmt.Println(string(bytes))
}

type AlfredResponseItem struct {
	Title     string               `json:"title"`
	Subtitle  *string              `json:"subtitle,omitempty"`
	Match     *string              `json:"match,omitempty"`
	Auto      *string              `json:"autocomplete,omitempty"`
	Arg       interface{}          `json:"arg,omitempty"`
	UID       *string              `json:"uid,omitempty"`
	Valid     bool                 `json:"valid"`
	Type      string               `json:"type,omitempty"`
	Text      *itemText            `json:"text,omitempty"`
	Icon      *Icon                `json:"icon,omitempty"`
	Quicklook string               `json:"quicklookurl,omitempty"`
	Variables map[string]string    `json:"variables,omitempty"`
	Mods      map[string]*Modifier `json:"mods,omitempty"`
	Actions   map[string][]string  `json:"action,omitempty"`
}

type itemText struct {
	// Copied to the clipboard on CMD+C
	Copy *string `json:"copy,omitempty"`
	// Shown in Alfred's Large Type window on CMD+L
	Large *string `json:"largetype,omitempty"`
}
type Modifier struct {
	// The modifier key, e.g. "cmd", "alt".
	// With Alfred 4+, modifiers can be combined, e.g. "cmd+alt", "ctrl+shift+cmd"
	Key      string
	arg      []string
	subtitle *string
	valid    bool
	icon     *Icon
	vars     map[string]string
}
type Icon struct {
	Value string   `json:"path"`           // Path or UTI
	Type  IconType `json:"type,omitempty"` // "fileicon", "filetype" or ""
}

type IconType string

// Valid icon types.
// const (
//
//	// IconTypeImage Indicates that Icon.Value is the path to an image file that should
//	// be used as the Item's icon.
//	IconTypeImage IconType = ""
//	// IconTypeFileIcon Indicates that Icon.Value points to an object whose icon should be show
//	// in Alfred, e.g. combine with "/Applications/Safari.app" to show Safari's icon.
//	IconTypeFileIcon IconType = "fileicon"
//	// IconTypeFileType Indicates that Icon.Value is a UTI, e.g. "public.folder",
//	// which will give you the icon for a folder.
//	IconTypeFileType IconType = "filetype"
//
// )
var (
	// IconWorkflow Workflow's own icon

	sysIcons = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/"
	//IconAccount System icons
	//IconAccount   = &Icon{Value: sysIcons + "Accounts.icns"}
	//IconBurn      = &Icon{Value: sysIcons + "BurningIcon.icns"}
	//IconClock     = &Icon{Value: sysIcons + "Clock.icns"}
	//IconColor     = &Icon{Value: sysIcons + "ProfileBackgroundColor.icns"}
	//IconColour    = &Icon{Value: sysIcons + "ProfileBackgroundColor.icns"}
	//IconEject     = &Icon{Value: sysIcons + "EjectMediaIcon.icns"}
	IconError = &Icon{Value: sysIcons + "AlertStopIcon.icns"}
	//IconFavorite  = &Icon{Value: sysIcons + "ToolbarFavoritesIcon.icns"}
	//IconFavourite = &Icon{Value: sysIcons + "ToolbarFavoritesIcon.icns"}
	//IconGroup     = &Icon{Value: sysIcons + "GroupIcon.icns"}
	//IconHelp      = &Icon{Value: sysIcons + "HelpIcon.icns"}
	//IconHome      = &Icon{Value: sysIcons + "HomeFolderIcon.icns"}
	//IconInfo      = &Icon{Value: sysIcons + "ToolbarInfo.icns"}
	//IconNetwork   = &Icon{Value: sysIcons + "GenericNetworkIcon.icns"}
	IconNote = &Icon{Value: sysIcons + "AlertNoteIcon.icns"}
	//IconSettings  = &Icon{Value: sysIcons + "ToolbarAdvanced.icns"}
	//IconSwirl     = &Icon{Value: sysIcons + "ErasingIcon.icns"}
	//IconSwitch    = &Icon{Value: sysIcons + "General.icns"}
	//IconSync      = &Icon{Value: sysIcons + "Sync.icns"}
	//IconTrash     = &Icon{Value: sysIcons + "TrashIcon.icns"}
	//IconUser      = &Icon{Value: sysIcons + "UserIcon.icns"}

	IconWarning = &Icon{Value: sysIcons + "AlertCautionIcon.icns"}
	//IconWeb       = &Icon{Value: sysIcons + "BookmarkIcon.icns"}
)

func AlfredError(title string, sub *string) *AlfredResponseItem {
	return &AlfredResponseItem{
		Valid:    false,
		Title:    title,
		Subtitle: sub,
		Icon:     IconError,
	}
}

func AlfredWarning(title string, sub *string) *AlfredResponseItem {
	return &AlfredResponseItem{
		Valid:    false,
		Title:    title,
		Subtitle: sub,
		Icon:     IconWarning,
	}
}

func AlfredInfo(title string, sub *string) *AlfredResponseItem {
	return &AlfredResponseItem{
		Valid:    false,
		Title:    title,
		Subtitle: sub,
		Icon:     IconNote,
	}
}

func BackgroundUpdate() error {
	cmd := exec.Command(os.Args[0], "update")
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Printf("Background pid %#v", cmd.Process.Pid)
	return nil
}
