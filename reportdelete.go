package main

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
)

var reportDirRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}/\d{6}-[A-Z2-7]{8}$`)

type reportDeleteServer struct {
	token    string
	root     string
	s3Bucket string
	s3Client *minio.Client
}

func (d *reportDeleteServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || subtle.ConstantTimeCompare([]byte(token), []byte(d.token)) != 1 {
		http.Error(w, "Unauthorised.", http.StatusUnauthorized)
		return
	}
	reportDir := strings.TrimPrefix(r.URL.Path, "/")
	if !reportDirRegexp.MatchString(reportDir) {
		http.Error(w, "invalid report path", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	log := zerolog.Ctx(ctx).With().
		Str("component", "report_delete").
		Str("report_dir", reportDir).
		Logger()

	deleted := 0
	for obj := range d.s3Client.ListObjects(ctx, d.s3Bucket, minio.ListObjectsOptions{Prefix: reportDir + "/", Recursive: true}) {
		if obj.Err != nil {
			log.Err(obj.Err).Msg("Failed to list report objects")
			http.Error(w, "error listing report", http.StatusInternalServerError)
			return
		}
		if err := d.s3Client.RemoveObject(ctx, d.s3Bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
			log.Err(err).Str("key", obj.Key).Msg("Failed to delete report object")
			http.Error(w, "error deleting report", http.StatusInternalServerError)
			return
		}
		deleted++
	}

	localPath := filepath.Join(d.root, filepath.FromSlash(reportDir))
	if _, err := os.Stat(localPath); err == nil {
		if err := os.RemoveAll(localPath); err != nil {
			log.Err(err).Msg("Failed to delete local report")
			http.Error(w, "error deleting report", http.StatusInternalServerError)
			return
		}
		deleted++
	}

	if deleted == 0 {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	log.Info().Int("deleted", deleted).Msg("Deleted report")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"deleted": deleted})
}
