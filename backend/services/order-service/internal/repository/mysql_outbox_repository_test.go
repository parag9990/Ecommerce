package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLOutboxRepositoryInsertsPendingEvent(t *testing.T) {
	db, mock := newMockDB(t)
	repository, _ := NewMySQLOutboxRepository(db)
	order, _ := validOrderAndClaim()
	event := validCreatedOutboxEvent(order)
	mock.ExpectExec("INSERT INTO order_outbox_events").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.InsertOutboxEvent(context.Background(), event); err != nil {
		t.Fatalf("InsertOutboxEvent() error = %v", err)
	}
	assertMockExpectations(t, mock)
}
