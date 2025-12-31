// Package database provides launchpad database functions
package database

import (
	"fmt"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/utils"
	"github.com/pkg/errors"
)

// GetMissing returns a list of the rest of the apps not in the config

// ClearGroups clears out items related to groups
func (lp *LaunchPad) ClearGroups() error {
	utils.Indent(log.Info, 2)("clear out groups")
	var items []Item
	if err := lp.DB.Where("type in (?)", []int{RootType, FolderRootType, PageType}).Delete(&items).Error; err != nil {
		return fmt.Errorf("delete items associated with groups failed: %w", err)
	}
	// return lp.DB.Exec("DELETE FROM groups;").Error
	return nil
}

// EnableTriggers enables item update triggers
func (lp *LaunchPad) EnableTriggers() error {
	utils.Indent(log.Info, 2)("enabling SQL update triggers")
	if err := lp.DB.Exec("UPDATE dbinfo SET value=0 WHERE key='ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "could not update `ignore_items_update_triggers` to 0")
	}
	return nil
}

// DisableTriggers disables item update triggers
func (lp *LaunchPad) DisableTriggers() error {
	utils.Indent(log.Info, 2)("disabling SQL update triggers")
	if err := lp.DB.Exec("UPDATE dbinfo SET value=1 WHERE key='ignore_items_update_triggers';").Error; err != nil {
		return errors.Wrap(err, "could not update `ignore_items_update_triggers` to 1")
	}
	return nil
}

// TriggersDisabled returns true if triggers are disabled
func (lp *LaunchPad) TriggersDisabled() bool {
	var dbinfo DBInfo
	if err := lp.DB.Where("key in (?)", []string{"ignore_items_update_triggers"}).Find(&dbinfo).Error; err != nil {
		log.WithError(err).Error("dbinfo query failed")
	}
	if dbinfo.Value == "1" {
		return true
	}
	return false
}
