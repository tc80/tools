package util

const (
	// ImportAllMaxVersions is the maximum number of versions we will import.
	// When no versions exist in cdnjs and we are trying to import all of them,
	// limit it to the a few last versions to avoid publishing too many outdated
	// versions.
	ImportAllMaxVersions = 10

	// MaxFileSize is the file size in bytes accepted by cdnjs (10MiB).
	MaxFileSize int64 = 10485760

	// MinNpmMonthlyDownloads is the minimum number of monthly downloads
	// from npm needed for a library to be accepted into cdnjs.
	MinNpmMonthlyDownloads = 800

	// MinGitHubStars is the minimum number of stars for a library to be
	// accepted into cdnjs.
	MinGitHubStars = 200

	// MaxBulkWritePayload is the maximum payload size for a request to write
	// to Workers KV in bulk. Note that this will be the sum of all values only.
	// The max bulk request size is 100MiB (104857600), so this will provide
	// enough leeway for any metadata stored with each key (up to 1024 bytes),
	// long keys, and verbose JSON syntax.
	MaxBulkWritePayload int64 = 1e8

	// MaxMetadataSize is the maximum metadata in bytes that can be stored for a
	// particular KV entry.
	MaxMetadataSize int64 = 1024
)

var (
	// AcceptedExtensions maps all accepted extensions to a boolean determining
	// if the extension should be compressed when inserted into KV.
	AcceptedExtensions = map[string]bool{
		".js":    true,
		".css":   true,
		".png":   true,
		".gif":   true,
		".jpg":   true,
		".jpeg":  true,
		".swf":   true,
		".svg":   true,
		".ttf":   true,
		".eot":   true,
		".woff":  false,
		".woff2": false,
		".otf":   true,
		".map":   true,
		".cur":   true,
		".mp3":   true,
		".ogg":   true,
		".aac":   true,
		".scss":  true,
		".webp":  true,
		".json":  true,
		".wasm":  true,
		".ts":    true,
		".hpb":   true,
		".lang":  true,
		".mjs":   true,
	}
)
