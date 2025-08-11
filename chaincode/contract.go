package chaincode

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Pretty logging for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	// Set global log level
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// GetClientIdentity returns the client identity from the transaction context
func (t *SimpleChaincode) GetClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	clientIdentity, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get client identity")
		return "", err
	}
	return clientIdentity, nil
}
const index = "color~name"

// SimpleChaincode implements the fabric-contract-api-go programming model
type SimpleChaincode struct {
	contractapi.Contract
}

type Asset struct {
	DocType        string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID             string `json:"ID"`      //the field tags are needed to keep case from bouncing around
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int    `json:"appraisedValue"`
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// PaginatedQueryResult structure used for returning paginated query results and metadata
type PaginatedQueryResult struct {
	Records             []*Asset `json:"records"`
	FetchedRecordsCount int32    `json:"fetchedRecordsCount"`
	Bookmark            string   `json:"bookmark"`
}

// CreateAsset initializes a new asset in the ledger
func (t *SimpleChaincode) CreateAsset(ctx contractapi.TransactionContextInterface, assetID, color string, size int, owner string, appraisedValue int) error {
	log.Info().
		Str("function", "CreateAsset").
		Str("assetID", assetID).
		Str("color", color).
		Int("size", size).
		Str("owner", owner).
		Int("appraisedValue", appraisedValue).
		Msg("Creating new asset")

	exists, err := t.AssetExists(ctx, assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to check if asset exists")
		return fmt.Errorf("failed to get asset: %v", err)
	}
	if exists {
		log.Warn().Str("assetID", assetID).Msg("Asset already exists")
		return fmt.Errorf("asset already exists: %s", assetID)
	}

	asset := &Asset{
		DocType:        "asset",
		ID:             assetID,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to marshal asset to JSON")
		return err
	}

	err = ctx.GetStub().PutState(assetID, assetBytes)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to put asset in ledger")
		return err
	}

	log.Debug().Str("assetID", assetID).Msg("Asset successfully stored in ledger")

	//  Create an index to enable color-based range queries, e.g. return all blue assets.
	//  An 'index' is a normal key-value entry in the ledger.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	colorNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.Color, asset.ID})
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Str("color", color).Msg("Failed to create composite key for color index")
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	err = ctx.GetStub().PutState(colorNameIndexKey, value)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Str("color", color).Msg("Failed to store color index")
		return err
	}

	log.Info().Str("assetID", assetID).Str("color", color).Msg("Asset created successfully with color index")
	return nil
}

// ReadAsset retrieves an asset from the ledger
func (t *SimpleChaincode) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	log.Info().Str("function", "ReadAsset").Str("assetID", assetID).Msg("Reading asset from ledger")

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to get asset from ledger")
		return nil, fmt.Errorf("failed to get asset %s: %v", assetID, err)
	}
	if assetBytes == nil {
		log.Warn().Str("assetID", assetID).Msg("Asset does not exist")
		return nil, fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetBytes, &asset)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to unmarshal asset from JSON")
		return nil, err
	}

	log.Info().Str("assetID", assetID).Str("owner", asset.Owner).Str("color", asset.Color).Msg("Asset read successfully")
	return &asset, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *SimpleChaincode) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	log.Info().Str("function", "DeleteAsset").Str("assetID", assetID).Msg("Deleting asset from ledger")

	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to read asset before deletion")
		return err
	}

	err = ctx.GetStub().DelState(assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to delete asset from ledger")
		return fmt.Errorf("failed to delete asset %s: %v", assetID, err)
	}

	colorNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.Color, asset.ID})
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Str("color", asset.Color).Msg("Failed to create composite key for color index deletion")
		return err
	}

	// Delete index entry
	err = ctx.GetStub().DelState(colorNameIndexKey)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Str("color", asset.Color).Msg("Failed to delete color index")
		return err
	}

	log.Info().Str("assetID", assetID).Str("color", asset.Color).Msg("Asset and color index deleted successfully")
	return nil
}

// TransferAsset transfers an asset by setting a new owner name on the asset
func (t *SimpleChaincode) TransferAsset(ctx contractapi.TransactionContextInterface, assetID, newOwner string) error {
	log.Info().
		Str("function", "TransferAsset").
		Str("assetID", assetID).
		Str("newOwner", newOwner).
		Msg("Transferring asset ownership")

	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to read asset for transfer")
		return err
	}

	oldOwner := asset.Owner
	asset.Owner = newOwner
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to marshal asset for transfer")
		return err
	}

	err = ctx.GetStub().PutState(assetID, assetBytes)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to update asset in ledger during transfer")
		return err
	}

	log.Info().
		Str("assetID", assetID).
		Str("oldOwner", oldOwner).
		Str("newOwner", newOwner).
		Msg("Asset transferred successfully")
	return nil
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Asset, error) {
	log.Debug().Msg("Constructing query response from iterator")

	var assets []*Asset
	assetCount := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get next result from iterator")
			return nil, err
		}
		var asset Asset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			log.Error().Err(err).Str("key", queryResult.Key).Msg("Failed to unmarshal asset from query result")
			return nil, err
		}
		assets = append(assets, &asset)
		assetCount++
	}

	log.Debug().Int("assetCount", assetCount).Msg("Query response construction completed")
	return assets, nil
}

// GetAssetsByRange performs a range query based on the start and end keys provided.
// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
func (t *SimpleChaincode) GetAssetsByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Asset, error) {
	log.Info().
		Str("function", "GetAssetsByRange").
		Str("startKey", startKey).
		Str("endKey", endKey).
		Msg("Performing range query on assets")

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		log.Error().Err(err).Str("startKey", startKey).Str("endKey", endKey).Msg("Failed to get state by range")
		return nil, err
	}
	defer resultsIterator.Close()

	assets, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		log.Error().Err(err).Str("startKey", startKey).Str("endKey", endKey).Msg("Failed to construct query response")
		return nil, err
	}

	log.Info().Int("count", len(assets)).Str("startKey", startKey).Str("endKey", endKey).Msg("Range query completed successfully")
	return assets, nil
}

// TransferAssetByColor will transfer assets of a given color to a certain new owner.
// Uses GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// Example: GetStateByPartialCompositeKey/RangeQuery
func (t *SimpleChaincode) TransferAssetByColor(ctx contractapi.TransactionContextInterface, color, newOwner string) error {
	log.Info().
		Str("function", "TransferAssetByColor").
		Str("color", color).
		Str("newOwner", newOwner).
		Msg("Transferring all assets of specified color")

	// Execute a key range query on all keys starting with 'color'
	coloredAssetResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{color})
	if err != nil {
		log.Error().Err(err).Str("color", color).Msg("Failed to get state by partial composite key")
		return err
	}
	defer coloredAssetResultsIterator.Close()

	transferCount := 0
	for coloredAssetResultsIterator.HasNext() {
		responseRange, err := coloredAssetResultsIterator.Next()
		if err != nil {
			log.Error().Err(err).Str("color", color).Msg("Failed to get next result from iterator")
			return err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			log.Error().Err(err).Str("color", color).Str("key", responseRange.Key).Msg("Failed to split composite key")
			return err
		}

		if len(compositeKeyParts) > 1 {
			returnedAssetID := compositeKeyParts[1]
			log.Debug().Str("assetID", returnedAssetID).Str("color", color).Msg("Processing asset for color transfer")

			asset, err := t.ReadAsset(ctx, returnedAssetID)
			if err != nil {
				log.Error().Err(err).Str("assetID", returnedAssetID).Str("color", color).Msg("Failed to read asset during color transfer")
				return err
			}
			asset.Owner = newOwner
			assetBytes, err := json.Marshal(asset)
			if err != nil {
				log.Error().Err(err).Str("assetID", returnedAssetID).Str("color", color).Msg("Failed to marshal asset during color transfer")
				return err
			}
			err = ctx.GetStub().PutState(returnedAssetID, assetBytes)
			if err != nil {
				log.Error().Err(err).Str("assetID", returnedAssetID).Str("color", color).Msg("Failed to update asset during color transfer")
				return fmt.Errorf("transfer failed for asset %s: %v", returnedAssetID, err)
			}
			transferCount++
		}
	}

	log.Info().Str("color", color).Str("newOwner", newOwner).Int("transferCount", transferCount).Msg("Color-based asset transfer completed successfully")
	return nil
}

// QueryAssetsByOwner queries for assets based on the owners name.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (t *SimpleChaincode) QueryAssetsByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Asset, error) {
	log.Info().Str("function", "QueryAssetsByOwner").Str("owner", owner).Msg("Querying assets by owner")

	queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)
	log.Debug().Str("queryString", queryString).Msg("Generated query string for owner")

	assets, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		log.Error().Err(err).Str("owner", owner).Msg("Failed to query assets by owner")
		return nil, err
	}

	log.Info().Str("owner", owner).Int("count", len(assets)).Msg("Owner query completed successfully")
	return assets, nil
}

// QueryAssets uses a query string to perform a query for assets.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Ad hoc rich query
func (t *SimpleChaincode) QueryAssets(ctx contractapi.TransactionContextInterface, queryString string) ([]*Asset, error) {
	log.Info().Str("function", "QueryAssets").Str("queryString", queryString).Msg("Performing ad hoc query on assets")

	assets, err := getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		log.Error().Err(err).Str("queryString", queryString).Msg("Failed to perform ad hoc query")
		return nil, err
	}

	log.Info().Str("queryString", queryString).Int("count", len(assets)).Msg("Ad hoc query completed successfully")
	return assets, nil
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Asset, error) {
	log.Debug().Str("queryString", queryString).Msg("Executing query string")

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		log.Error().Err(err).Str("queryString", queryString).Msg("Failed to get query result")
		return nil, err
	}
	defer resultsIterator.Close()

	assets, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		log.Error().Err(err).Str("queryString", queryString).Msg("Failed to construct query response from iterator")
		return nil, err
	}

	log.Debug().Str("queryString", queryString).Int("count", len(assets)).Msg("Query string execution completed")
	return assets, nil
}

// GetAssetsByRangeWithPagination performs a range query based on the start and end key,
// page size and a bookmark.
// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// Example: Pagination with Range Query
func (t *SimpleChaincode) GetAssetsByRangeWithPagination(ctx contractapi.TransactionContextInterface, startKey string, endKey string, pageSize int, bookmark string) (*PaginatedQueryResult, error) {
	log.Info().
		Str("function", "GetAssetsByRangeWithPagination").
		Str("startKey", startKey).
		Str("endKey", endKey).
		Int("pageSize", pageSize).
		Str("bookmark", bookmark).
		Msg("Performing paginated range query on assets")

	resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	if err != nil {
		log.Error().Err(err).Str("startKey", startKey).Str("endKey", endKey).Int("pageSize", pageSize).Msg("Failed to get state by range with pagination")
		return nil, err
	}
	defer resultsIterator.Close()

	assets, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		log.Error().Err(err).Str("startKey", startKey).Str("endKey", endKey).Msg("Failed to construct query response for paginated range query")
		return nil, err
	}

	result := &PaginatedQueryResult{
		Records:             assets,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}

	log.Info().
		Str("startKey", startKey).
		Str("endKey", endKey).
		Int("fetchedCount", int(responseMetadata.FetchedRecordsCount)).
		Str("bookmark", responseMetadata.Bookmark).
		Msg("Paginated range query completed successfully")
	return result, nil
}

// QueryAssetsWithPagination uses a query string, page size and a bookmark to perform a query
// for assets. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// Example: Pagination with Ad hoc Rich Query
func (t *SimpleChaincode) QueryAssetsWithPagination(ctx contractapi.TransactionContextInterface, queryString string, pageSize int, bookmark string) (*PaginatedQueryResult, error) {
	log.Info().
		Str("function", "QueryAssetsWithPagination").
		Str("queryString", queryString).
		Int("pageSize", pageSize).
		Str("bookmark", bookmark).
		Msg("Performing paginated ad hoc query on assets")

	return getQueryResultForQueryStringWithPagination(ctx, queryString, int32(pageSize), bookmark)
}

// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringWithPagination(ctx contractapi.TransactionContextInterface, queryString string, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {
	log.Debug().
		Str("queryString", queryString).
		Int32("pageSize", pageSize).
		Str("bookmark", bookmark).
		Msg("Executing paginated query string")

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		log.Error().Err(err).Str("queryString", queryString).Int32("pageSize", pageSize).Msg("Failed to get query result with pagination")
		return nil, err
	}
	defer resultsIterator.Close()

	assets, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		log.Error().Err(err).Str("queryString", queryString).Msg("Failed to construct query response for paginated query")
		return nil, err
	}

	result := &PaginatedQueryResult{
		Records:             assets,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}

	log.Debug().
		Str("queryString", queryString).
		Int("fetchedCount", int(responseMetadata.FetchedRecordsCount)).
		Str("bookmark", responseMetadata.Bookmark).
		Msg("Paginated query string execution completed")
	return result, nil
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
func (t *SimpleChaincode) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Info().Str("function", "GetAssetHistory").Str("assetID", assetID).Msg("Getting asset history")

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to get history for key")
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	recordCount := 0
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			log.Error().Err(err).Str("assetID", assetID).Msg("Failed to get next history record")
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				log.Error().Err(err).Str("assetID", assetID).Str("txId", response.TxId).Msg("Failed to unmarshal asset from history record")
				return nil, err
			}
		} else {
			asset = Asset{
				ID: assetID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			log.Error().Err(err).Str("assetID", assetID).Str("txId", response.TxId).Msg("Failed to parse timestamp from history record")
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
		recordCount++
	}

	log.Info().Str("assetID", assetID).Int("recordCount", recordCount).Msg("Asset history retrieved successfully")
	return records, nil
}

// AssetExists returns true when asset with given ID exists in the ledger.
func (t *SimpleChaincode) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	log.Debug().Str("function", "AssetExists").Str("assetID", assetID).Msg("Checking if asset exists")

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		log.Error().Err(err).Str("assetID", assetID).Msg("Failed to read asset from world state")
		return false, fmt.Errorf("failed to read asset %s from world state. %v", assetID, err)
	}

	exists := assetBytes != nil
	log.Debug().Str("assetID", assetID).Bool("exists", exists).Msg("Asset existence check completed")
	return exists, nil
}

// InitLedger creates the initial set of assets in the ledger.
func (t *SimpleChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Info().Str("function", "InitLedger").Msg("Initializing ledger with sample assets")

	assets := []Asset{
		{DocType: "asset", ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{DocType: "asset", ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{DocType: "asset", ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{DocType: "asset", ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{DocType: "asset", ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{DocType: "asset", ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	log.Info().Int("assetCount", len(assets)).Msg("Creating initial assets in ledger")

	for i, asset := range assets {
		log.Debug().
			Int("index", i).
			Str("assetID", asset.ID).
			Str("color", asset.Color).
			Str("owner", asset.Owner).
			Msg("Creating initial asset")

		err := t.CreateAsset(ctx, asset.ID, asset.Color, asset.Size, asset.Owner, asset.AppraisedValue)
		if err != nil {
			log.Error().Err(err).Str("assetID", asset.ID).Msg("Failed to create initial asset")
			return err
		}
	}

	log.Info().Int("assetCount", len(assets)).Msg("Ledger initialization completed successfully")
	return nil
}
