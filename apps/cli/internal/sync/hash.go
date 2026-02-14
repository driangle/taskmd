package sync

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

// HashExternalTask returns a deterministic hash of an external task's content.
func HashExternalTask(ext ExternalTask) string {
	h := sha256.New()
	fmt.Fprintf(h, "id:%s\n", ext.ExternalID)
	fmt.Fprintf(h, "title:%s\n", ext.Title)
	fmt.Fprintf(h, "desc:%s\n", ext.Description)
	fmt.Fprintf(h, "status:%s\n", ext.Status)
	fmt.Fprintf(h, "priority:%s\n", ext.Priority)
	fmt.Fprintf(h, "assignee:%s\n", ext.Assignee)
	fmt.Fprintf(h, "labels:%s\n", strings.Join(ext.Labels, ","))
	fmt.Fprintf(h, "url:%s\n", ext.URL)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// HashLocalFile returns the SHA-256 hash of a file's contents.
func HashLocalFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:]), nil
}
