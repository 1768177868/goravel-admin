package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20251227063517AddFulltextIndexToOperationLogsRequest struct{}

// Signature The unique signature for the migration.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Signature() string {
	return "20251227063517_add_fulltext_index_to_operation_logs_request"
}

// Up Run the migrations.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Up() error {
	const tableName = "operation_logs"
	const indexName = "ft_request"
	const columnName = "request"

	if !facades.Schema().HasTable(tableName) {
		return nil
	}

	hasIndex, err := hasIndex(tableName, indexName)
	if err != nil {
		return err
	}
	if hasIndex {
		return nil
	}

	return createCompatibleTextIndex(tableName, columnName, indexName)
}

// Down Reverse the migrations.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Down() error {
	const tableName = "operation_logs"
	const indexName = "ft_request"

	if !facades.Schema().HasTable(tableName) {
		return nil
	}

	hasIndex, err := hasIndex(tableName, indexName)
	if err != nil {
		return err
	}
	if !hasIndex {
		return nil
	}

	return dropIndexCompatible(tableName, indexName)
}
