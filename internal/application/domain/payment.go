package domain

type PaymentData struct {
	Receipt Receipt `json:"receipt"`
}

type Receipt struct {
	Items []ReceiptItem `json:"items"`
}

type ReceiptItem struct {
	Description string            `json:"description"`
	Quantity    string            `json:"quantity"`
	Amount      ReceiptItemAmount `json:"amount"`
	VAT         VAT               `json:"vat_code"`
}

type ReceiptItemAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type VAT int

const (
	VATUnknown VAT = iota
	VATNo          // Без НДС
	VAT0           // НДС по ставке 0%
	VAT10          // НДС по ставке 10%
	VAT20          // НДС чека по ставке 20%
	VAT110         // НДС чека по расчетной ставке 10/110
	VAT120         // НДС чека по расчетной ставке 20/120
)
