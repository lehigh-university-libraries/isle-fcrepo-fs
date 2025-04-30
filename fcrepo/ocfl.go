package fcrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Inventory struct {
	Head     string              `json:"head"`
	Versions map[string]Version  `json:"versions"`
	Manifest map[string][]string `json:"manifest"`
}

type Version struct {
	State map[string][]string `json:"state"`
}

func getOCFLDir(objectID string) string {
	hash := sha256.Sum256([]byte(objectID))
	digest := hex.EncodeToString(hash[:])

	tupleSize := 3
	numberOfTuples := 3
	base := "/fcrepo"
	var builder strings.Builder
	builder.WriteString(base)

	for i := 0; i < numberOfTuples*tupleSize; i += tupleSize {
		builder.WriteString("/")
		builder.WriteString(digest[i : i+tupleSize])
	}

	builder.WriteString("/")
	builder.WriteString(digest)
	return builder.String()
}

func RealPath(uri string) string {
	ocflDir := getOCFLDir(uri)
	inventoryPath := filepath.Join(ocflDir, "extensions/0005-mutable-head/head/inventory.json")

	inventoryBytes, err := os.ReadFile(inventoryPath)
	if err != nil {
		return ""
	}

	var inventory Inventory

	if err := json.Unmarshal(inventoryBytes, &inventory); err != nil {
		return ""
	}

	components := strings.Split(uri, "/")
	filename := components[len(components)-1]

	state := inventory.Versions[inventory.Head].State
	for digest, files := range state {
		for _, file := range files {
			if file == filename && len(inventory.Manifest[digest]) > 0 {
				return filepath.Join(ocflDir, inventory.Manifest[digest][0])
			}
		}
	}

	return ""
}
