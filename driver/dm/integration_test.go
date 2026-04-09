//go:build dm

package dm

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
)

type testRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:64;not null;uniqueIndex:idx_name_unique"`
	Age       int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func TestDMCrudAndTransaction(t *testing.T) {
	dsn := os.Getenv("DM_TEST_DSN")
	if dsn == "" {
		t.Skip("skip dm integration test: DM_TEST_DSN is empty")
	}

	db, err := gorm.Open(New(Config{
		DSN:      dsn,
		GormMode: 0,
	}))
	if err != nil {
		t.Fatalf("open dm failed: %v", err)
	}

	table := "agent_dm_integration_records"
	if err := db.Table(table).AutoMigrate(&testRecord{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Table(table).Migrator().DropTable(&testRecord{})
	})

	// create
	rec := testRecord{Name: "alice", Age: 18}
	if err := db.Table(table).Create(&rec).Error; err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if rec.ID == 0 {
		t.Fatalf("create failed: empty id")
	}

	// read
	var got testRecord
	if err := db.Table(table).First(&got, rec.ID).Error; err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if got.Name != "alice" || got.Age != 18 {
		t.Fatalf("read mismatch: %+v", got)
	}

	// update
	if err := db.Table(table).Where("id = ?", rec.ID).Update("age", 19).Error; err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if err := db.Table(table).First(&got, rec.ID).Error; err != nil {
		t.Fatalf("read after update failed: %v", err)
	}
	if got.Age != 19 {
		t.Fatalf("update mismatch: age=%d", got.Age)
	}

	// transaction rollback
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(table).Create(&testRecord{Name: "rollback", Age: 99}).Error; err != nil {
			return err
		}
		return gorm.ErrInvalidTransaction // force rollback
	})
	if err == nil {
		t.Fatalf("expected rollback error, got nil")
	}
	var cnt int64
	if err := db.Table(table).Where("name = ?", "rollback").Count(&cnt).Error; err != nil {
		t.Fatalf("count rollback row failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("rollback failed: found rows=%d", cnt)
	}

	// transaction commit
	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.Table(table).Create(&testRecord{Name: "commit", Age: 20}).Error
	})
	if err != nil {
		t.Fatalf("commit transaction failed: %v", err)
	}
	if err := db.Table(table).Where("name = ?", "commit").Count(&cnt).Error; err != nil {
		t.Fatalf("count commit row failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("commit failed: expected 1 row, got %d", cnt)
	}

	// unique constraint conflict
	if err := db.Table(table).Create(&testRecord{Name: "dup_name", Age: 30}).Error; err != nil {
		t.Fatalf("create first unique row failed: %v", err)
	}
	err = db.Table(table).Create(&testRecord{Name: "dup_name", Age: 31}).Error
	if err == nil {
		t.Fatalf("expected unique conflict error, got nil")
	}

	// delete
	if err := db.Table(table).Delete(&testRecord{}, rec.ID).Error; err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := db.Table(table).Where("id = ?", rec.ID).Count(&cnt).Error; err != nil {
		t.Fatalf("count after delete failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("delete mismatch: found rows=%d", cnt)
	}
}
