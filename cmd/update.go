package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}

func newUpdateCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update fdn to the latest release from GitHub",
		Long: `Check for the latest fdn release on GitHub and update the current binary.
Downloads the correct archive for your platform and verifies checksums.
On Windows, the binary is written to fdn.new.exe and must be replaced manually.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Check for updates without installing")
	return cmd
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate(cmd *cobra.Command, dryRun bool) error {
	out := cmd.OutOrStdout()

	current := versionString()
	currentClean := strings.TrimPrefix(current, "v")
	if i := strings.Index(currentClean, "-"); i >= 0 {
		currentClean = currentClean[:i]
	}

	latest, assets, err := fetchLatestRelease(cmd)
	if err != nil {
		return fmt.Errorf("fetch release: %w", err)
	}

	latestVer := strings.TrimPrefix(latest, "v")
	if latestVer == currentClean || latestVer == "" {
		fmt.Fprintf(out, "Already up-to-date (%s)\n", current)
		return nil
	}

	fmt.Fprintf(out, "Current: %s  Latest: %s\n", current, latest)

	if dryRun {
		fmt.Fprintf(out, "Would update to %s (dry-run)\n", latest)
		return nil
	}

	assetName := releaseAssetName(latestVer)

	var assetURL string
	for _, a := range assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (%s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	checksums, err := fetchChecksums(cmd, assets)
	if err != nil {
		return fmt.Errorf("fetch checksums: %w", err)
	}

	data, err := downloadAsset(cmd, assetURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	expectedHash, ok := checksums[assetName]
	if ok {
		h := sha256.Sum256(data)
		got := hex.EncodeToString(h[:])
		if !strings.EqualFold(got, expectedHash) {
			return fmt.Errorf("checksum mismatch for %s", assetName)
		}
	} else {
		fmt.Fprintf(out, "Warning: no checksum found for %s (verifying skipped)\n", assetName)
	}

	bin, err := extractBinary(data, assetName)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if err := replaceBinary(cmd, bin); err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated to %s\n", latest)
	return nil
}

func releaseAssetName(ver string) string {
	suffix := ".tar.gz"
	if runtime.GOOS == "windows" {
		suffix = ".zip"
	}
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}
	return fmt.Sprintf("fdn_%s_%s_%s%s", ver, cases.Title(language.English).String(runtime.GOOS), arch, suffix)
}

func fetchLatestRelease(cmd *cobra.Command) (string, []ghAsset, error) {
	req, err := http.NewRequestWithContext(cmd.Context(), "GET",
		"https://api.github.com/repos/hobbymarks/fdn/releases/latest", nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", nil, fmt.Errorf("parse release: %w", err)
	}
	return rel.TagName, rel.Assets, nil
}

func fetchChecksums(cmd *cobra.Command, assets []ghAsset) (map[string]string, error) {
	var url string
	for _, a := range assets {
		if strings.HasSuffix(a.Name, "checksums.txt") {
			url = a.BrowserDownloadURL
			break
		}
	}
	if url == "" {
		return map[string]string{}, nil
	}

	req, err := http.NewRequestWithContext(cmd.Context(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	entries := map[string]string{}
	for line := range strings.SplitSeq(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			entries[parts[1]] = strings.ToLower(parts[0])
		}
	}
	return entries, nil
}

func downloadAsset(cmd *cobra.Command, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(cmd.Context(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	name := url[strings.LastIndex(url, "/")+1:]
	fmt.Fprintf(cmd.OutOrStdout(), "Downloading %s...", name)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				fmt.Fprintln(cmd.OutOrStdout())
				return
			case <-ticker.C:
				fmt.Fprint(cmd.OutOrStdout(), ".")
			}
		}
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 100<<20))
	close(done)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func extractBinary(data []byte, assetName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractFromZip(data)
	}
	return extractFromTarGz(data)
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && (hdr.Name == "fdn" || hdr.Name == "fdn.exe") {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
	}
	return nil, fmt.Errorf("binary not found in archive")
}

func extractFromZip(data []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("zip: %w", err)
	}
	for _, f := range reader.File {
		if f.Name == "fdn.exe" || f.Name == "fdn" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, rc); err != nil {
				rc.Close()
				return nil, err
			}
			rc.Close()
			return buf.Bytes(), nil
		}
	}
	return nil, fmt.Errorf("binary not found in archive")
}

func replaceBinary(cmd *cobra.Command, data []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find current executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	if runtime.GOOS == "windows" {
		newPath := exe + ".new"
		if err := os.WriteFile(newPath, data, 0755); err != nil {
			return fmt.Errorf("write new binary: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s. Replace %s manually and restart.\n", newPath, exe)
		return nil
	}

	tmp, err := os.CreateTemp(filepath.Dir(exe), ".fdn-update-*")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Chmod(0755); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("chmod: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}
