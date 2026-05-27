package domain

const (
	PaymentDatabaseName = "payment_db"
)

type PaymentSchemaDefinition struct {
	DatabaseName string                   `json:"database_name"`
	Tables       []PaymentTableDefinition `json:"tables"`
}

type PaymentTableDefinition struct {
	Name        string   `json:"name"`
	Columns     []string `json:"columns"`
	UniqueKeys  []string `json:"unique_keys"`
	Indexes     []string `json:"indexes"`
	ForeignKeys []string `json:"foreign_keys"`
}

type PaymentSchemaVerification struct {
	DatabaseName       string   `json:"database_name"`
	Ready              bool     `json:"ready"`
	MissingTables      []string `json:"missing_tables,omitempty"`
	MissingIndexes     []string `json:"missing_indexes,omitempty"`
	MissingForeignKeys []string `json:"missing_foreign_keys,omitempty"`
}

func PaymentSchema() PaymentSchemaDefinition {
	tables := make([]PaymentTableDefinition, 0, len(paymentSchemaTables))
	for _, table := range paymentSchemaTables {
		tables = append(tables, PaymentTableDefinition{
			Name:        table.Name,
			Columns:     append([]string(nil), table.Columns...),
			UniqueKeys:  append([]string(nil), table.UniqueKeys...),
			Indexes:     append([]string(nil), table.Indexes...),
			ForeignKeys: append([]string(nil), table.ForeignKeys...),
		})
	}
	return PaymentSchemaDefinition{
		DatabaseName: PaymentDatabaseName,
		Tables:       tables,
	}
}

func (v PaymentSchemaVerification) WithReadyFlag() PaymentSchemaVerification {
	v.Ready = len(v.MissingTables) == 0 && len(v.MissingIndexes) == 0 && len(v.MissingForeignKeys) == 0
	return v
}

var paymentSchemaTables = []PaymentTableDefinition{
	{
		Name: "payments",
		Columns: []string{
			"id",
			"payment_id",
			"order_id",
			"user_id",
			"provider",
			"provider_payment_id",
			"provider_intent_id",
			"status",
			"currency",
			"amount",
			"captured_amount",
			"refunded_amount",
			"idempotency_key",
			"retry_of_payment_id",
			"root_payment_id",
			"attempt_no",
			"retry_request_key",
			"failure_code",
			"failure_message",
			"created_at",
			"updated_at",
		},
		UniqueKeys:  []string{"uk_payments_payment_id", "uk_payments_idempotency", "uk_payments_retry_request", "uk_payments_root_attempt"},
		Indexes:     []string{"idx_payments_order", "idx_payments_provider_payment", "idx_payments_provider_intent", "idx_payments_status_created", "idx_payments_order_status", "idx_payments_root_status"},
		ForeignKeys: []string{"fk_payments_retry_of"},
	},
	{
		Name: "payment_attempts",
		Columns: []string{
			"id",
			"attempt_id",
			"payment_id",
			"provider_attempt_id",
			"status",
			"failure_code",
			"failure_message",
			"raw_provider_response",
			"created_at",
		},
		UniqueKeys:  []string{"uk_payment_attempts_attempt_id"},
		Indexes:     []string{"idx_payment_attempts_payment"},
		ForeignKeys: []string{"fk_payment_attempts_payment"},
	},
	{
		Name: "refunds",
		Columns: []string{
			"id",
			"refund_id",
			"payment_id",
			"provider_refund_id",
			"status",
			"currency",
			"amount",
			"reason",
			"requested_by",
			"reviewed_by",
			"review_reason",
			"reviewed_at",
			"idempotency_key",
			"created_at",
			"updated_at",
		},
		UniqueKeys:  []string{"uk_refunds_refund_id", "uk_refunds_idempotency"},
		Indexes:     []string{"idx_refunds_payment_status", "idx_refunds_provider_refund"},
		ForeignKeys: []string{"fk_refunds_payment"},
	},
	{
		Name: "payment_webhook_events",
		Columns: []string{
			"id",
			"webhook_event_id",
			"provider",
			"provider_event_id",
			"event_type",
			"processed",
			"payload",
			"received_at",
			"processed_at",
		},
		UniqueKeys: []string{"uk_webhook_provider_event", "uk_webhook_event_id"},
		Indexes:    []string{"idx_webhook_processed_received"},
	},
	{
		Name: "payment_reconciliations",
		Columns: []string{
			"id",
			"reconciliation_id",
			"provider",
			"settlement_id",
			"status",
			"payment_id",
			"provider_payment_id",
			"details",
			"created_at",
		},
		UniqueKeys: []string{"uk_reconciliation_id"},
		Indexes:    []string{"idx_reconciliation_provider_status"},
	},
}
