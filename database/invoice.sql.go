// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: invoice.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

const addInvoiceDetails = `-- name: AddInvoiceDetails :exec
UPDATE invoices
SET email      = $2,
    updated_at = NOW()
WHERE uuid = $1
`

type AddInvoiceDetailsParams struct {
	Uuid  uuid.UUID `db:"uuid" json:"uuid"`
	Email *string   `db:"email" json:"email"`
}

func (q *Queries) AddInvoiceDetails(ctx context.Context, arg *AddInvoiceDetailsParams) error {
	_, err := q.db.Exec(ctx, addInvoiceDetails, arg.Uuid, arg.Email)
	return err
}

const completeInvoice = `-- name: CompleteInvoice :exec
UPDATE invoices
SET charge_id  = $2,
    updated_at = NOW()
WHERE uuid = $1
`

type CompleteInvoiceParams struct {
	Uuid     uuid.UUID `db:"uuid" json:"uuid"`
	ChargeID *string   `db:"charge_id" json:"charge_id"`
}

func (q *Queries) CompleteInvoice(ctx context.Context, arg *CompleteInvoiceParams) error {
	_, err := q.db.Exec(ctx, completeInvoice, arg.Uuid, arg.ChargeID)
	return err
}

const createInvoice = `-- name: CreateInvoice :exec
INSERT INTO invoices (uuid, user_id, payload)
VALUES ($1, $2, $3)
`

type CreateInvoiceParams struct {
	Uuid    uuid.UUID    `db:"uuid" json:"uuid"`
	UserID  int64        `db:"user_id" json:"user_id"`
	Payload pgtype.JSONB `db:"payload" json:"payload"`
}

func (q *Queries) CreateInvoice(ctx context.Context, arg *CreateInvoiceParams) error {
	_, err := q.db.Exec(ctx, createInvoice, arg.Uuid, arg.UserID, arg.Payload)
	return err
}
