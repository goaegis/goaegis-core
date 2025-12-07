package remote

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dovakiin0/goaegis-core/aegis/addons"
	"github.com/dovakiin0/goaegis-core/aegis/config"
)

// S3Addon simulates loading config from a remote source (like S3) with hot reload.
// In production, this would use actual AWS SDK to fetch from S3.
// This demonstrates how to create addons for remote config sources.
type S3Addon struct {
	bucket       string
	key          string
	pollInterval time.Duration
	watchCh      chan struct{}
	stopCh       chan struct{}
	lastModTime  time.Time
	// For demo purposes, we use a local file to simulate S3
	localFilePath string
}

// New creates a new S3 config loader addon
func New(bucket, key string, pollInterval time.Duration, localSimulation string) *S3Addon {
	return &S3Addon{
		bucket:        bucket,
		key:           key,
		pollInterval:  pollInterval,
		localFilePath: localSimulation, // Simulates S3 with local file
	}
}

func (s *S3Addon) Name() string {
	return "s3-config-loader"
}

func (s *S3Addon) Init(core interface{}) error {
	s.watchCh = make(chan struct{}, 1)
	s.stopCh = make(chan struct{})

	log.Printf("[s3-loader] Initialized (bucket: %s, key: %s)", s.bucket, s.key)
	return nil
}

// OnBeforeConfigLoad provides this addon as the config source
func (s *S3Addon) OnBeforeConfigLoad(path string) (addons.ConfigSource, error) {
	log.Printf("[s3-loader] Providing S3 config source (simulated with: %s)", s.localFilePath)

	// Get initial mod time
	info, err := os.Stat(s.localFilePath)
	if err == nil {
		s.lastModTime = info.ModTime()
	}

	// Start polling for changes
	go s.pollForChanges()

	// Return ourselves as the ConfigSource
	return s, nil
}

// LoadFiles implements addons.ConfigSource - fetches config from S3
// Supports both single file and multi-file (nested directory) configurations
func (s *S3Addon) LoadFiles() (map[string][]byte, error) {
	log.Printf("[s3-loader] Loading config from s3://%s/%s", s.bucket, s.key)

	// In production with single file:
	// result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
	//     Bucket: &s.bucket,
	//     Key:    &s.key,
	// })
	// data, _ := io.ReadAll(result.Body)
	// return map[string][]byte{s.key: data}, nil

	// In production with folder/prefix (for nested structure):
	// Use s3Client.ListObjectsV2 to get all YAML files under a prefix,
	// then fetch each one with GetObject and add to the map.

	// For demo, read from local file (single file simulation)
	data, err := os.ReadFile(s.localFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config (simulating S3): %w", err)
	}

	log.Printf("[s3-loader] Loaded %d bytes from S3 (simulated)", len(data))

	// Return as single-file map
	return map[string][]byte{
		s.key: data,
	}, nil
}

// Watch implements addons.ConfigSource - returns channel for hot reload
func (s *S3Addon) Watch() <-chan struct{} {
	return s.watchCh
}

func (s *S3Addon) OnConfigValidate(cfg *config.Config) (*config.Config, error) {
	return cfg, nil
}

func (s *S3Addon) OnConfigLoad(cfg *config.Config) error {
	log.Println("[s3-loader] Config loaded successfully")
	return nil
}

func (s *S3Addon) OnAuthorize(ctx *addons.Context) (addons.Decision, error) {
	return addons.Abstain, nil
}

func (s *S3Addon) pollForChanges() {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	log.Printf("[s3-loader] Polling S3 for changes every %v", s.pollInterval)

	for {
		select {
		case <-ticker.C:
			// In production, check S3 object ETag or LastModified
			// For demo, check local file modification time
			info, err := os.Stat(s.localFilePath)
			if err != nil {
				continue
			}

			if info.ModTime().After(s.lastModTime) {
				s.lastModTime = info.ModTime()
				log.Println("[s3-loader] S3 config changed, triggering reload")
				select {
				case s.watchCh <- struct{}{}:
				default:
				}
			}
		case <-s.stopCh:
			log.Println("[s3-loader] Stopped polling")
			return
		}
	}
}

func (s *S3Addon) Shutdown() error {
	log.Println("[s3-loader] Shutting down")
	close(s.stopCh)
	return nil
}
