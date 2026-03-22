package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hobbymarks/fdn/utils"
	"gorm.io/gorm"
)

func MigrateLegacyFDNDatabases(fdnDir string) error {
	fdnPath := filepath.Join(fdnDir, FDNDBFileName)
	if utils.PathExist(fdnPath) {
		return nil
	}

	cfgPath := filepath.Join(fdnDir, "cfg.db")
	rdPath := filepath.Join(fdnDir, "rd.db")
	hasCfg := utils.PathExist(cfgPath)
	hasRd := utils.PathExist(rdPath)
	if !hasCfg && !hasRd {
		return nil
	}

	if hasCfg && !hasRd {
		return os.Rename(cfgPath, fdnPath)
	}
	if hasRd && !hasCfg {
		return os.Rename(rdPath, fdnPath)
	}

	_db := utils.OpenDB(fdnPath)
	defer utils.DBClose(_db)

	if err := _db.AutoMigrate(&TermWord{}, &ToSepWord{}, &Separator{}, &Record{}); err != nil {
		_ = os.Remove(fdnPath)
		return err
	}

	sqlDB, err := _db.DB()
	if err != nil {
		_ = os.Remove(fdnPath)
		return err
	}

	cfgEsc, err := sqliteQuotePath(cfgPath)
	if err != nil {
		_ = os.Remove(fdnPath)
		return err
	}
	rdEsc, err := sqliteQuotePath(rdPath)
	if err != nil {
		_ = os.Remove(fdnPath)
		return err
	}

	if _, err := sqlDB.Exec("ATTACH DATABASE '" + cfgEsc + "' AS legacy_cfg"); err != nil {
		_ = os.Remove(fdnPath)
		return fmt.Errorf("attach legacy cfg: %w", err)
	}
	if _, err := sqlDB.Exec("ATTACH DATABASE '" + rdEsc + "' AS legacy_rd"); err != nil {
		_, _ = sqlDB.Exec("DETACH DATABASE legacy_cfg")
		_ = os.Remove(fdnPath)
		return fmt.Errorf("attach legacy rd: %w", err)
	}

	for _, tbl := range []string{"term_words", "to_sep_words", "separators"} {
		if err := copyAttachedTable(_db, "legacy_cfg", tbl); err != nil {
			_, _ = sqlDB.Exec("DETACH DATABASE legacy_rd")
			_, _ = sqlDB.Exec("DETACH DATABASE legacy_cfg")
			_ = os.Remove(fdnPath)
			return err
		}
	}
	if err := copyAttachedTable(_db, "legacy_rd", "records"); err != nil {
		_, _ = sqlDB.Exec("DETACH DATABASE legacy_rd")
		_, _ = sqlDB.Exec("DETACH DATABASE legacy_cfg")
		_ = os.Remove(fdnPath)
		return err
	}

	_, _ = sqlDB.Exec("DETACH DATABASE legacy_rd")
	_, _ = sqlDB.Exec("DETACH DATABASE legacy_cfg")

	if err := os.Remove(cfgPath); err != nil {
		return err
	}
	if err := os.Remove(rdPath); err != nil {
		return err
	}
	return nil
}

func copyAttachedTable(db *gorm.DB, attachName, table string) error {
	var n int64
	q := `SELECT count(*) FROM ` + attachName + `.sqlite_master WHERE type = 'table' AND name = ?`
	if err := db.Raw(q, table).Scan(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	return db.Exec(`INSERT INTO ` + table + ` SELECT * FROM ` + attachName + `.` + table).Error
}
