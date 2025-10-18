package exceptions

import "fmt"

// NectarliteException is the base exception for all nectarlite errors.
type NectarliteException struct {
	Message string
}

func (e *NectarliteException) Error() string {
	return e.Message
}

// TransactionError represents an error during transaction operations.
type TransactionError struct {
	*NectarliteException
}

// NewTransactionError creates a new TransactionError.
func NewTransactionError(msg string) *TransactionError {
	return &TransactionError{
		NectarliteException: &NectarliteException{Message: msg},
	}
}

// MissingKeyError represents an error when a key is missing.
type MissingKeyError struct {
	*NectarliteException
}

// NewMissingKeyError creates a new MissingKeyError.
func NewMissingKeyError(account, role string) *MissingKeyError {
	return &MissingKeyError{
		NectarliteException: &NectarliteException{
			Message: fmt.Sprintf("No %s key for account '%s'", role, account),
		},
	}
}

// InvalidKeyFormatError represents an error when a key has an invalid format.
type InvalidKeyFormatError struct {
	*NectarliteException
}

// NewInvalidKeyFormatError creates a new InvalidKeyFormatError.
func NewInvalidKeyFormatError(msg string) *InvalidKeyFormatError {
	return &InvalidKeyFormatError{
		NectarliteException: &NectarliteException{Message: msg},
	}
}

// NodeError represents an error from a Hive node.
type NodeError struct {
	*NectarliteException
}

// NewNodeError creates a new NodeError.
func NewNodeError(msg string) *NodeError {
	return &NodeError{
		NectarliteException: &NectarliteException{Message: msg},
	}
}
