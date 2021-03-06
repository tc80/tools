package kv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	// workaround for now since cloudflare's API does not currently
	// return a cloudflare.Response object for api.ReadWorkersKV
	keyNotFound    = "key not found"
	authError      = "Authentication error"
	serviceFailure = "service failure"
)

var (
	srisNamespaceID               = util.GetEnv("WORKERS_KV_SRIS_NAMESPACE_ID")
	filesNamespaceID              = util.GetEnv("WORKERS_KV_FILES_NAMESPACE_ID")
	versionsNamespaceID           = util.GetEnv("WORKERS_KV_VERSIONS_NAMESPACE_ID")
	packagesNamespaceID           = util.GetEnv("WORKERS_KV_PACKAGES_NAMESPACE_ID")
	aggregatedMetadataNamespaceID = util.GetEnv("WORKERS_KV_AGGREGATED_METADATA_NAMESPACE_ID")
	accountID                     = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiToken                      = util.GetEnv("WORKERS_KV_API_TOKEN")
	api                           = getAPI()
)

// KeyNotFoundError represents a KV key not found.
type KeyNotFoundError struct {
	key string
	err string
}

// Error is used to satisfy the error interface.
func (k KeyNotFoundError) Error() string {
	return fmt.Sprintf("%s (%s): %s", keyNotFound, k.key, k.err)
}

// AuthError represents an authentication error.
type AuthError struct {
	err string
}

// Error is used to satisfy the error interface.
func (a AuthError) Error() string {
	return fmt.Sprintf("%s: %s", authError, a.err)
}

// Represents a KV write request, consisting of
// a string key, a []byte value, and file metadata.
// The name field is used to identify this write request
// with a human-readable friendly name.
type writeRequest struct {
	key   string
	name  string
	value []byte
	meta  *FileMetadata
}

// FileMetadata represents metadata for a
// particular KV.
type FileMetadata struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	SRI          string `json:"sri,omitempty"`
}

// Gets a new *cloudflare.API.
func getAPI() *cloudflare.API {
	a, err := cloudflare.NewWithAPIToken(apiToken, cloudflare.UsingAccount(accountID))
	util.Check(err)
	return a
}

// Ensure a response is successful and the error is nil.
func checkSuccess(r cloudflare.Response, err error) error {
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf("kv fail: %v", r)
	}
	return nil
}

// read reads an entry from Workers KV.
func read(key, namespaceID string) ([]byte, error) {
	var bytes []byte
	var err error
	for i := 0; i < util.MaxKVAttempts; i++ {
		bytes, err = api.ReadWorkersKV(context.Background(), namespaceID, key)
		if err != nil {
			errString := err.Error()

			// check for service failure and retry
			if strings.Contains(errString, serviceFailure) {
				continue
			}

			// check for key not found
			if strings.Contains(errString, keyNotFound) {
				return nil, KeyNotFoundError{key, errString}
			}

			// check for authentication error
			if strings.Contains(errString, authError) {
				return nil, AuthError{errString}
			}
		}

		break
	}

	return bytes, err
}

// Returns all KVs that start with a prefix.
func listByPrefix(prefix, namespaceID string) ([]cloudflare.StorageKey, error) {
	var cursor *string
	var results []cloudflare.StorageKey
	for {
		o := cloudflare.ListWorkersKVsOptions{
			Prefix: &prefix,
			Cursor: cursor,
		}

		resp, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, o)
		if err != nil {
			return nil, err
		}

		results = append(results, resp.Result...)

		if resp.Cursor == "" {
			return results, nil
		}

		cursor = &resp.Cursor
	}
}

// Lists by prefix and then returns only the names of the results.
func listByPrefixNamesOnly(prefix, namespaceID string) ([]string, error) {
	results, err := listByPrefix(prefix, namespaceID)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, r := range results {
		names = append(names, r.Name)
	}

	return names, nil
}

// Encodes a byte array to a base64 string.
func encodeToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// Encodes key-value pairs to base64 and writes them to KV in multiple bulk requests.
// Returns the list of human-readable names of successful writes.
func encodeAndWriteKVBulk(ctx context.Context, kvs []*writeRequest, namespaceID string, panicOversized bool) ([]string, error) {
	var bulkWrites []cloudflare.WorkersKVBulkWriteRequest
	var bulkWrite []*cloudflare.WorkersKVPair
	var successfulWrites []string
	var totalSize, totalKeys int64

	for _, kv := range kvs {
		if unencodedSize := int64(len(kv.value)); unencodedSize > util.MaxFileSize {
			util.Debugf(ctx, "ignoring oversized file: %s (%d)\n", kv.key, unencodedSize)
			sentry.NotifyError(fmt.Errorf("ignoring oversized file: %s (%d)", kv.key, unencodedSize))
			if panicOversized {
				panic(fmt.Sprintf("oversized file: %s (%d)", kv.key, unencodedSize))
			}
			continue
		}
		// Note that after encoding in base64 the size may get larger, but after decoding
		// it will be reduced, so it is okay if the size is larger than util.MaxFileSize after encoding base64.
		// However, we still need to check for the KV bulk request limit of 100MiB.
		encodedValue := encodeToBase64(kv.value)
		size := int64(len(encodedValue))
		writePair := &cloudflare.WorkersKVPair{
			Key:    kv.key,
			Value:  encodedValue,
			Base64: true,
		}
		if kv.meta != nil {
			// Marshal metadata into JSON bytes.
			bytes, err := json.Marshal(kv.meta)
			if err != nil {
				return nil, err
			}
			metasize := int64(len(bytes))
			if metasize > util.MaxMetadataSize {
				util.Debugf(ctx, "ignoring oversized metadata: %s (%d)\n", kv.key, metasize)
				sentry.NotifyError(fmt.Errorf("oversized metadata: %s (%d) - %s", kv.key, metasize, bytes))
				if panicOversized {
					panic(fmt.Sprintf("oversized metadata: %s (%d)", kv.key, metasize))
				}
				continue
			}
			util.Debugf(ctx, "writing metadata: %s\n", bytes)
			writePair.Metadata = kv.meta
			size += metasize
		}
		if totalSize+size > util.MaxBulkWritePayload || totalKeys == util.MaxBulkKeys {
			// Create a new bulk since we are over a limit.
			bulkWrites = append(bulkWrites, bulkWrite)
			bulkWrite = []*cloudflare.WorkersKVPair{}
			totalSize = 0
			totalKeys = 0
		}
		bulkWrite = append(bulkWrite, writePair)
		successfulWrites = append(successfulWrites, kv.name)
		totalSize += size
		totalKeys++
	}
	bulkWrites = append(bulkWrites, bulkWrite)

	for i, b := range bulkWrites {
		util.Debugf(ctx, "writing bulk %d/%d (keys=%d)...\n", i+1, len(bulkWrites), len(b))
		for j := 0; j < util.MaxKVAttempts; j++ {
			r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, b)

			// check for service failure and retry
			if err != nil && strings.Contains(err.Error(), serviceFailure) {
				if j == util.MaxKVAttempts-1 {
					return nil, err // no more attempts
				}
				continue // retry
			}

			if err = checkSuccess(r, err); err != nil {
				return nil, err
			}

			break
		}
	}

	return successfulWrites, nil
}

// InsertNewVersionToKV inserts a new version to KV and returns the uploaded version files as JSON.
// The `fullPathToVersion` string will be useful if the version is downloaded to
// a temporary directory, not necessarily always in `$BOT_BASE_PATH/cdnjs/ajax/libs/`.
//
// Note that this function will also compress the files, generating brotli/gzip entries
// to KV where necessary.
//
// Note this function will NOT update package metadata.
//
// For example:
// InsertNewVersionToKV("1000hz-bootstrap-validator", "0.10.0", "/tmp/1000hz-bootstrap-validator/0.10.0")
func InsertNewVersionToKV(ctx context.Context, pkg, version, fullPathToVersion string, metaOnly, srisOnly, filesOnly, noPush, panicOversized bool) ([]string, []byte, []string, []string, int, int, error) {
	fromVersionPaths, err := util.ListFilesInVersion(ctx, fullPathToVersion)
	if err != nil {
		return nil, nil, nil, nil, 0, 0, err
	}
	sort.Strings(fromVersionPaths)

	var versionBytes []byte
	if !filesOnly && !srisOnly && !noPush {
		// write version metadata to KV
		versionBytes, err = updateKVVersion(ctx, pkg, version, fromVersionPaths)
		if err != nil {
			return nil, nil, nil, nil, 0, 0, err
		}
		if metaOnly {
			return fromVersionPaths, versionBytes, nil, nil, 0, 0, nil
		}
	}

	// write files to KV
	srisPushedToKV, filesPushedToKV, theoreticalSRIKeys, theoreticalFileKeys, err := updateKVFiles(ctx, pkg, version, fullPathToVersion, fromVersionPaths, srisOnly, filesOnly, noPush, panicOversized)
	return fromVersionPaths, versionBytes, srisPushedToKV, filesPushedToKV, theoreticalSRIKeys, theoreticalFileKeys, err
}
