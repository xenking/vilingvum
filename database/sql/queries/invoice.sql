-- name: CreateInvoice :exec
INSERT INTO invoices (uuid, user_id, payload)
VALUES ($1, $2, $3);

-- name: AddInvoiceDetails :exec
UPDATE invoices
SET email      = $2,
    updated_at = NOW()
WHERE uuid = $1;

-- name: CompleteInvoice :exec
UPDATE invoices
SET charge_id  = $2,
    updated_at = NOW()
WHERE uuid = $1;