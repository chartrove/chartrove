// fasten-backup is an explicit local operator tool. It never loads config.yaml,
// FASTEN_* variables, an application profile, a provider or a web listener.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg/database"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "Backup failed:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out, diagnostic io.Writer) error {
	flags := flag.NewFlagSet("fasten-backup", flag.ContinueOnError)
	flags.SetOutput(diagnostic)
	source := flags.String("source", "", "Absolute path to an existing encrypted Chartrove SQLite database")
	output := flags.String("output", "", "Absolute NEW backup directory inside an existing private parent")
	keyFile := flags.String("key-file", "", "Absolute private regular file containing exact key bytes (not a command-line key)")
	timeout := flags.Duration("timeout", 5*time.Minute, "Maximum duration, greater than zero and at most 5m")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *source == "" || *output == "" || *keyFile == "" || flags.NArg() != 0 || *timeout <= 0 || *timeout > 5*time.Minute {
		return errors.New("explicit source, output, key-file and a timeout in (0, 5m] are required")
	}
	key, err := readPrivateKey(*keyFile)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	receipt, err := database.BackupEncryptedSQLite(ctx, *source, *output, key)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(receipt)
}

func readPrivateKey(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("key file path must be absolute")
	}
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode().Perm()&0077 != 0 {
		return "", errors.New("key file must be a private regular file, not a symlink")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", errors.New("could not open private key file")
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) || after.Mode().Perm()&0077 != 0 {
		return "", errors.New("key file changed while opening")
	}
	key, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(key) == 0 || len(key) > 4096 {
		return "", errors.New("key file must contain 1 to 4096 exact key bytes")
	}
	// Do not TrimSpace: whitespace is part of an existing encryption key.
	return string(key), nil
}
