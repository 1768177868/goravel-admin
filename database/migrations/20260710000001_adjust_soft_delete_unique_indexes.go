package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260710000001AdjustSoftDeleteUniqueIndexes struct{}

func (r *M20260710000001AdjustSoftDeleteUniqueIndexes) Signature() string {
	return "20260710000001_adjust_soft_delete_unique_indexes"
}

func (r *M20260710000001AdjustSoftDeleteUniqueIndexes) Up() error {
	if err := r.adjustAdminUsernameIndex(); err != nil {
		return err
	}
	return r.adjustPaymentMethodCodeIndex()
}

func (r *M20260710000001AdjustSoftDeleteUniqueIndexes) adjustAdminUsernameIndex() error {
	if !facades.Schema().HasTable("admins") {
		return nil
	}

	compositeIndex := "admins_username_deleted_at_unique"
	if facades.Schema().HasIndex("admins", compositeIndex) {
		if err := facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.DropUnique("username", "deleted_at")
		}); err != nil {
			return err
		}
	}

	singleIndex := "admins_username_unique"
	if !facades.Schema().HasIndex("admins", singleIndex) {
		return facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.Unique("username")
		})
	}
	return nil
}

func (r *M20260710000001AdjustSoftDeleteUniqueIndexes) adjustPaymentMethodCodeIndex() error {
	if !facades.Schema().HasTable("payment_methods") {
		return nil
	}

	compositeIndex := "payment_methods_code_deleted_at_unique"
	if facades.Schema().HasIndex("payment_methods", compositeIndex) {
		if err := facades.Schema().Table("payment_methods", func(table schema.Blueprint) {
			table.DropUnique("code", "deleted_at")
		}); err != nil {
			return err
		}
	}

	singleIndex := "payment_methods_code_unique"
	if !facades.Schema().HasIndex("payment_methods", singleIndex) {
		return facades.Schema().Table("payment_methods", func(table schema.Blueprint) {
			table.Unique("code")
		})
	}
	return nil
}

func (r *M20260710000001AdjustSoftDeleteUniqueIndexes) Down() error {
	return nil
}
