package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"syscall"
	"time"

	"code.crute.us/mcrute/golib/secrets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rclone/rclone/backend/b2"
	"github.com/rclone/rclone/cmd"
	rcloneFs "github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	fslog "github.com/rclone/rclone/fs/log"
	"github.com/rclone/rclone/fs/sync"
	"github.com/rclone/rclone/lib/atexit"
	"github.com/rclone/rclone/lib/terminal"

	_ "github.com/rclone/rclone/backend/all" // import all backends
	_ "github.com/rclone/rclone/lib/plugin"  // import plugins
)

// These variables get set at build time from the Makefile and exist
// to make it easier to open-source this code without passing out a
// bunch of internal information about the backup system.
var (
	defaultVaultMaterial string
	defaultB2Bucket      string
	defaultInstanceName  string
	defaultPushGateway   string
)

type b2Secret struct {
	AccountID string `mapstructure:"id"`
	Key       string `mapstructure:"key"`
}

type b2ConfigMapper struct {
	config map[string]string
}

func newB2ConfigMapper(secret *b2Secret) *b2ConfigMapper {
	return &b2ConfigMapper{
		config: map[string]string{
			// These values are the hard-coded defaults in the Rclone B2 backend
			// but are mandatory to make the backend work
			"chunk_size":             (96 * rcloneFs.Mebi).String(),
			"upload_cutoff":          (200 * rcloneFs.Mebi).String(),
			"upload_concurrency":     "4",
			"copy_cutoff":            (4 * rcloneFs.Gibi).String(),
			"download_auth_duration": (rcloneFs.Duration(7 * 24 * time.Hour)).String(),

			// B2 secrets
			"account": secret.AccountID,
			"key":     secret.Key,

			// Actually delete the files, don't just hide them
			"hard_delete": "true",
		},
	}
}

func (b *b2ConfigMapper) Get(key string) (value string, ok bool) {
	value, ok = b.config[key]
	return
}

// Set is no-op since it's not used and just here to satisfy the interface
func (b *b2ConfigMapper) Set(key, value string) {}

func getB2Config(ctx context.Context, secretPath string, disableAutodiscover bool) (cfg *b2ConfigMapper, err error) {
	var sc secrets.ClientManager

	if disableAutodiscover {
		if sc, err = secrets.NewVaultClient(nil); err != nil {
			return nil, err
		}
	} else {
		if sc, err = secrets.NewAutodiscoverVaultClient(ctx); err != nil {
			return nil, err
		}
	}

	if err := sc.Authenticate(ctx); err != nil {
		return nil, err
	}

	var b2Creds b2Secret
	if _, err := sc.Secret(ctx, secretPath, &b2Creds); err != nil {
		return nil, err
	}

	return newB2ConfigMapper(&b2Creds), nil
}

func loadEnvFile(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	defer fd.Close()

	var settings map[string]string
	if err := json.NewDecoder(fd).Decode(&settings); err != nil {
		return err
	}

	for k, v := range settings {
		syscall.Setenv(k, v)
	}

	return nil
}

func main() {
	vaultMaterial := flag.String("vault-material", defaultVaultMaterial, "Path to Vault material containing B2 credential")
	b2Bucket := flag.String("b2-bucket", defaultB2Bucket, "B2 bucket name for sync destination")
	instanceName := flag.String("instance", defaultInstanceName, "Instance label for pushgateway")
	pushGateway := flag.String("pushgateway", defaultPushGateway, "URL for pushgateway")
	jobName := flag.String("job", "rcloneReporter", "Job name label for pushgateway")
	noVaultAutodiscover := flag.Bool("no-discover-vault", false, "Disable autodiscovery of Vault host")
	envFile := flag.String("env-file", "/etc/default/restic-backup.json", "JSON file with environment variables to inject at process start; skipped if non-existent")
	flag.Parse()

	if *envFile != "" {
		if err := loadEnvFile(*envFile); err != nil {
			log.Fatalf("Error loading environment file %s: %s", *envFile, err)
		}
	}

	backupDir := os.Getenv("RCLONE_TO_B2_FROM")
	if len(flag.Args()) != 1 {
		if backupDir == "" {
			fmt.Printf("usage: %s <source>\n", os.Args[0])
			os.Exit(1)
		}
	} else {
		backupDir = flag.Args()[0]
	}

	ctx := context.Background()

	// Setup Rclone
	err := rcloneFs.GlobalOptionsInit()
	if err != nil {
		log.Fatalf("Error Rclone.GlobalOptionsInit: %s", err)
	}

	// Configure as if using command line arguments
	ci := rcloneFs.GetConfig(ctx)
	ci.UseMmap = true                         // --use-mmap
	ci.UseListR = true                        // --fast-list
	ci.DeleteMode = rcloneFs.DeleteModeDuring // --delete-during
	ci.LogLevel = rcloneFs.LogLevelInfo

	// Initialize the rest of Rclone
	fslog.InitLogging()
	accounting.Start(ctx)
	terminal.HideConsole()

	b2Config, err := getB2Config(ctx, *vaultMaterial, *noVaultAutodiscover)
	if err != nil {
		log.Fatalf("Error fetching B2 config: %s", err)
	}

	fdst, err := b2.NewFs(ctx, ":b2{GwbCk}", *b2Bucket, b2Config)
	if err != nil {
		log.Fatalf("Error creating B2 FS: %s", err)
	}

	fsrc, _ := cmd.NewFsFile(backupDir)

	if err := sync.Sync(ctx, fdst, fsrc, false); err != nil {
		log.Fatalf("Error during sync: %s", err)
	}

	remoteStats, err := accounting.GlobalStats().RemoteStats(false)
	if err != nil {
		log.Fatalf("Error gathering metrics: %s", err)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(&RcloneCollector{
		Instance: *instanceName,
		Metrics:  remoteStats,
	})

	push.New(*pushGateway, *jobName).Gatherer(reg).Push()

	atexit.Run()
}
