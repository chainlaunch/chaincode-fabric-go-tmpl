package chaincode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAssetStruct tests the Asset struct
func TestAssetStruct(t *testing.T) {
	asset := Asset{
		DocType:        "asset",
		ID:             "asset1",
		Color:          "blue",
		Size:           5,
		Owner:          "John",
		AppraisedValue: 100,
	}

	assert.Equal(t, "asset", asset.DocType)
	assert.Equal(t, "asset1", asset.ID)
	assert.Equal(t, "blue", asset.Color)
	assert.Equal(t, 5, asset.Size)
	assert.Equal(t, "John", asset.Owner)
	assert.Equal(t, 100, asset.AppraisedValue)
}

// TestHistoryQueryResult tests the HistoryQueryResult struct
func TestHistoryQueryResult(t *testing.T) {
	asset := &Asset{
		DocType:        "asset",
		ID:             "asset1",
		Color:          "blue",
		Size:           5,
		Owner:          "John",
		AppraisedValue: 100,
	}

	timestamp := time.Now()
	result := HistoryQueryResult{
		Record:    asset,
		TxId:      "tx123",
		Timestamp: timestamp,
		IsDelete:  false,
	}

	assert.Equal(t, asset, result.Record)
	assert.Equal(t, "tx123", result.TxId)
	assert.Equal(t, timestamp, result.Timestamp)
	assert.False(t, result.IsDelete)
}

// TestPaginatedQueryResult tests the PaginatedQueryResult struct
func TestPaginatedQueryResult(t *testing.T) {
	assets := []*Asset{
		{
			DocType:        "asset",
			ID:             "asset1",
			Color:          "blue",
			Size:           5,
			Owner:          "John",
			AppraisedValue: 100,
		},
		{
			DocType:        "asset",
			ID:             "asset2",
			Color:          "red",
			Size:           10,
			Owner:          "Jane",
			AppraisedValue: 200,
		},
	}

	result := PaginatedQueryResult{
		Records:             assets,
		FetchedRecordsCount: 2,
		Bookmark:            "bookmark123",
	}

	assert.Len(t, result.Records, 2)
	assert.Equal(t, int32(2), result.FetchedRecordsCount)
	assert.Equal(t, "bookmark123", result.Bookmark)
	assert.Equal(t, "asset1", result.Records[0].ID)
	assert.Equal(t, "asset2", result.Records[1].ID)
}

// TestSimpleChaincode tests that the SimpleChaincode struct can be instantiated
func TestSimpleChaincode(t *testing.T) {
	chaincode := &SimpleChaincode{}
	assert.NotNil(t, chaincode)
}

// TestIndexConstant tests that the index constant is defined
func TestIndexConstant(t *testing.T) {
	assert.Equal(t, "color~name", index)
}
