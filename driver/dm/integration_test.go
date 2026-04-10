//go:build dm

package dm

import (
	"os"
	"strconv"
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

	prefix := strconv.FormatInt(time.Now().UnixNano(), 10)

	t.Run("CreateReadUpdateDelete", func(t *testing.T) {
		rec := testRecord{Name: "alice_" + prefix, Age: 18}
		if err := db.Table(table).Create(&rec).Error; err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if rec.ID == 0 {
			t.Fatalf("create failed: empty id")
		}

		var got testRecord
		if err := db.Table(table).First(&got, rec.ID).Error; err != nil {
			t.Fatalf("read failed: %v", err)
		}
		if got.Name != rec.Name || got.Age != 18 {
			t.Fatalf("read mismatch: %+v", got)
		}

		if err := db.Table(table).Where("id = ?", rec.ID).Update("age", 19).Error; err != nil {
			t.Fatalf("update failed: %v", err)
		}
		if err := db.Table(table).First(&got, rec.ID).Error; err != nil {
			t.Fatalf("read after update failed: %v", err)
		}
		if got.Age != 19 {
			t.Fatalf("update mismatch: age=%d", got.Age)
		}

		if err := db.Table(table).Delete(&testRecord{}, rec.ID).Error; err != nil {
			t.Fatalf("delete failed: %v", err)
		}
		var cnt int64
		if err := db.Table(table).Where("id = ?", rec.ID).Count(&cnt).Error; err != nil {
			t.Fatalf("count after delete failed: %v", err)
		}
		if cnt != 0 {
			t.Fatalf("delete mismatch: found rows=%d", cnt)
		}
	})

	t.Run("BatchCreateWhereOrderLimitOffsetPluckCount", func(t *testing.T) {
		rows := []testRecord{
			{Name: "batch_" + prefix + "_1", Age: 21},
			{Name: "batch_" + prefix + "_2", Age: 22},
			{Name: "batch_" + prefix + "_3", Age: 23},
		}
		if err := db.Table(table).Create(&rows).Error; err != nil {
			t.Fatalf("batch create failed: %v", err)
		}

		var list []testRecord
		if err := db.Table(table).
			Where("name LIKE ?", "batch_"+prefix+"_%").
			Order("age desc").
			Limit(2).
			Offset(0).
			Find(&list).Error; err != nil {
			t.Fatalf("query list failed: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(list))
		}
		if list[0].Age < list[1].Age {
			t.Fatalf("order mismatch: %+v", list)
		}

		var names []string
		if err := db.Table(table).
			Where("name LIKE ?", "batch_"+prefix+"_%").
			Order("age asc").
			Pluck("name", &names).Error; err != nil {
			t.Fatalf("pluck failed: %v", err)
		}
		if len(names) != 3 {
			t.Fatalf("pluck count mismatch: %d", len(names))
		}

		var cnt int64
		if err := db.Table(table).Where("name LIKE ?", "batch_"+prefix+"_%").Count(&cnt).Error; err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if cnt != 3 {
			t.Fatalf("count mismatch: %d", cnt)
		}
	})

	t.Run("RawQueryAndExec", func(t *testing.T) {
		name := "raw_" + prefix
		if err := db.Exec(
			`INSERT INTO "AGENT_DM_INTEGRATION_RECORDS"("NAME","AGE","CREATED_AT","UPDATED_AT") VALUES (?,?,?,?)`,
			name, 28, time.Now(), time.Now(),
		).Error; err != nil {
			t.Fatalf("raw exec insert failed: %v", err)
		}

		var out struct {
			Cnt int64 `gorm:"column:cnt"`
		}
		if err := db.Raw(`SELECT COUNT(*) AS cnt FROM "AGENT_DM_INTEGRATION_RECORDS" WHERE "NAME" = ?`, name).Scan(&out).Error; err != nil {
			t.Fatalf("raw query failed: %v", err)
		}
		if out.Cnt != 1 {
			t.Fatalf("raw count mismatch: %d", out.Cnt)
		}
	})

	t.Run("TransactionRollbackAndCommit", func(t *testing.T) {
		var cnt int64

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Table(table).Create(&testRecord{Name: "rollback_" + prefix, Age: 99}).Error; err != nil {
				return err
			}
			return gorm.ErrInvalidTransaction
		})
		if err == nil {
			t.Fatalf("expected rollback error, got nil")
		}
		if err := db.Table(table).Where("name = ?", "rollback_"+prefix).Count(&cnt).Error; err != nil {
			t.Fatalf("count rollback row failed: %v", err)
		}
		if cnt != 0 {
			t.Fatalf("rollback failed: found rows=%d", cnt)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			return tx.Table(table).Create(&testRecord{Name: "commit_" + prefix, Age: 20}).Error
		})
		if err != nil {
			t.Fatalf("commit transaction failed: %v", err)
		}
		if err := db.Table(table).Where("name = ?", "commit_"+prefix).Count(&cnt).Error; err != nil {
			t.Fatalf("count commit row failed: %v", err)
		}
		if cnt != 1 {
			t.Fatalf("commit failed: expected 1 row, got %d", cnt)
		}
	})

	t.Run("ManualTransactionBeginCommitRollback", func(t *testing.T) {
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx failed: %v", tx.Error)
		}
		name := "manual_commit_" + prefix
		if err := tx.Table(table).Create(&testRecord{Name: name, Age: 40}).Error; err != nil {
			t.Fatalf("tx create failed: %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("tx commit failed: %v", err)
		}

		var cnt int64
		if err := db.Table(table).Where("name = ?", name).Count(&cnt).Error; err != nil {
			t.Fatalf("count after commit failed: %v", err)
		}
		if cnt != 1 {
			t.Fatalf("manual commit mismatch: %d", cnt)
		}

		tx = db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx2 failed: %v", tx.Error)
		}
		rollbackName := "manual_rollback_" + prefix
		if err := tx.Table(table).Create(&testRecord{Name: rollbackName, Age: 41}).Error; err != nil {
			t.Fatalf("tx2 create failed: %v", err)
		}
		if err := tx.Rollback().Error; err != nil {
			t.Fatalf("tx2 rollback failed: %v", err)
		}
		if err := db.Table(table).Where("name = ?", rollbackName).Count(&cnt).Error; err != nil {
			t.Fatalf("count after rollback failed: %v", err)
		}
		if cnt != 0 {
			t.Fatalf("manual rollback mismatch: %d", cnt)
		}
	})

	t.Run("UniqueConstraintConflict", func(t *testing.T) {
		name := "dup_" + prefix
		if err := db.Table(table).Create(&testRecord{Name: name, Age: 30}).Error; err != nil {
			t.Fatalf("create first unique row failed: %v", err)
		}
		err = db.Table(table).Create(&testRecord{Name: name, Age: 31}).Error
		if err == nil {
			t.Fatalf("expected unique conflict error, got nil")
		}
	})
}
