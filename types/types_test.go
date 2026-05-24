package types

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
)

func TestParseAmount(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		amt, err := ParseAmount("123.456 HIVE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(amt.Value-123.456) > 1e-9 {
			t.Fatalf("unexpected value: %v", amt.Value)
		}
		if amt.Symbol != "HIVE" {
			t.Fatalf("unexpected symbol: %s", amt.Symbol)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		if _, err := ParseAmount("1000HIVE"); err == nil {
			t.Fatalf("expected error for malformed amount")
		}
	})

	t.Run("invalid numeric value", func(t *testing.T) {
		if _, err := ParseAmount("abc HIVE"); err == nil {
			t.Fatalf("expected error for invalid numeric value")
		}
	})
}

func TestAmountBytes(t *testing.T) {
	t.Run("serializes hive amount", func(t *testing.T) {
		a := NewAmount(1.234, "HIVE")
		b, err := a.Bytes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(b) != 16 {
			t.Fatalf("unexpected length: %d", len(b))
		}

		value := binary.LittleEndian.Uint64(b[:8])
		if value != 1234 {
			t.Fatalf("unexpected satoshi value: %d", value)
		}
		if b[8] != 0x03 {
			t.Fatalf("unexpected precision byte: %d", b[8])
		}
		expectedSymbol := []byte{'S', 'T', 'E', 'E', 'M', 0x00, 0x00}
		for i, v := range expectedSymbol {
			if b[9+i] != v {
				t.Fatalf("unexpected symbol byte at %d: got %d want %d", 9+i, b[9+i], v)
			}
		}
	})

	t.Run("unknown symbol", func(t *testing.T) {
		a := NewAmount(1, "TEST")
		if _, err := a.Bytes(); err == nil {
			t.Fatalf("expected error for unknown symbol")
		}
	})

	t.Run("symbol too long", func(t *testing.T) {
		a := NewAmount(1, "LONGSYMB")
		if _, err := a.Bytes(); err == nil {
			t.Fatalf("expected error for long symbol")
		}
	})
}

func TestAppliedOperationUnmarshal(t *testing.T) {
	jsonData := `{
		"trx_id": "0000000000000000000000000000000000000000",
		"block": 106666224,
		"trx_in_block": 4294967295,
		"op_in_trx": 4294967295,
		"virtual_op": true,
		"op": ["custom_json", {"id": "test"}]
	}`

	var op AppliedOperation
	if err := json.Unmarshal([]byte(jsonData), &op); err != nil {
		t.Fatalf("failed to unmarshal AppliedOperation: %v", err)
	}

	if op.TrxInBlock != 4294967295 {
		t.Errorf("expected TrxInBlock 4294967295, got %d", op.TrxInBlock)
	}
	if op.OpInTrx != 4294967295 {
		t.Errorf("expected OpInTrx 4294967295, got %d", op.OpInTrx)
	}
}

func TestOperationTupleUnmarshal(t *testing.T) {
	t.Run("array format", func(t *testing.T) {
		jsonData := `["transfer", {"from": "alice", "to": "bob", "amount": "1.000 HIVE"}]`
		var ot OperationTuple
		if err := json.Unmarshal([]byte(jsonData), &ot); err != nil {
			t.Fatalf("unexpected error unmarshaling array-based OperationTuple: %v", err)
		}
		if len(ot) != 2 || ot[0] != "transfer" {
			t.Fatalf("unexpected result: %#v", ot)
		}
	})

	t.Run("object format", func(t *testing.T) {
		jsonData := `{"type": "transfer", "value": {"from": "alice", "to": "bob", "amount": "1.000 HIVE"}}`
		var ot OperationTuple
		if err := json.Unmarshal([]byte(jsonData), &ot); err != nil {
			t.Fatalf("unexpected error unmarshaling object-based OperationTuple: %v", err)
		}
		if len(ot) != 2 || ot[0] != "transfer" {
			t.Fatalf("unexpected result: %#v", ot)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		jsonData := `"just a string"`
		var ot OperationTuple
		if err := json.Unmarshal([]byte(jsonData), &ot); err == nil {
			t.Fatalf("expected error for invalid OperationTuple format")
		}
	})
}
