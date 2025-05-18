package packager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type BuildResult struct {
	Platform  Platform
	OutputDir string
	Err       error
}

// BuildAll builds binaries concurrently and runs the UI build once.
// It places the UI and binaries into per-platform directories under `outdir`.
func BuildAll(pluginDir, version, outdir string, platforms []Platform) []BuildResult {
	// Step 1: Prepare all output dirs
	outputDirs := map[string]string{}
	for _, plat := range platforms {
		dir := filepath.Join(pluginDir, outdir, plat.Key())
		if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
			fmt.Printf("❌ Failed to create output dir for %s: %v\n", plat.Key(), err)
			continue
		}
		outputDirs[plat.Key()] = dir
	}

	// Step 2: Copy plugin.yaml meta into root of package
	pluginMeta := filepath.Join(pluginDir, "plugin.yaml")
	for _, plat := range platforms {
		dest := filepath.Join(outputDirs[plat.Key()], "plugin.yaml")
		if err := CopyFile(pluginMeta, dest); err != nil {
			fmt.Printf("❌ Failed to copy plugin.yaml to %s: %v\n", plat.Key(), err)
		}
	}

	// Step 3: Build UI once (concurrently)
	uiErrChan := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := buildUIAndCopy(pluginDir, platforms, outdir)
		uiErrChan <- err
	}()

	// Step 4: Build binaries concurrently
	binResults := make([]BuildResult, len(platforms))
	for i, plat := range platforms {
		wg.Add(1)
		go func(i int, plat Platform) {
			defer wg.Done()
			dir := outputDirs[plat.Key()]
			err := buildBinary(pluginDir, dir, plat)
			binResults[i] = BuildResult{Platform: plat, OutputDir: dir, Err: err}
		}(i, plat)
	}

	wg.Wait()

	if err := <-uiErrChan; err != nil {
		fmt.Println("❌ UI build failed:", err)
		for i := range binResults {
			if binResults[i].Err == nil {
				binResults[i].Err = fmt.Errorf("UI build failed: %v", err)
			}
		}
	}

	return binResults
}

func buildBinary(pluginDir, output string, plat Platform) error {
	binName := "plugin"
	if plat.OS == "windows" {
		binName += ".exe"
	}
	outPath := filepath.Join(output, "bin", binName)

	if _, err := os.Stat(outPath); err == nil {
		fmt.Printf("⚠️  Skipping %s (already built)\n", plat.Key())
		return nil
	}

	fmt.Printf("Building binary for %s...\n", plat.Key())

	cmd := exec.Command("go", "build", "-o", outPath, "./pkg")
	cmd.Dir = pluginDir
	cmd.Env = append(os.Environ(), "GOOS="+plat.OS, "GOARCH="+plat.Arch)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("binary build failed for %s: %w\n%s", plat.Key(), err, string(out))
	}
	fmt.Printf("✅ Built binary for %s\n", plat.Key())
	return nil
}

func buildUIAndCopy(pluginDir string, platforms []Platform, outdir string) error {
	fmt.Printf("Building ui...\n")

	uiPath := filepath.Join(pluginDir, "ui")

	// Run `pnpm run build`
	cmd := exec.Command("pnpm", "run", "build")
	cmd.Dir = uiPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("UI build error: %s\n%s", err, out)
	}

	// Copy dist/assets/* into each platform dir
	srcAssets := filepath.Join(uiPath, "dist", "assets")

	for _, plat := range platforms {
		destAssets := filepath.Join(pluginDir, outdir, plat.Key(), "assets")
		if err := os.MkdirAll(destAssets, 0755); err != nil {
			return fmt.Errorf("failed to create assets dir: %w", err)
		}

		err := filepath.Walk(srcAssets, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(srcAssets, path)
			dest := filepath.Join(destAssets, rel)
			return CopyFile(path, dest)
		})
		if err != nil {
			return fmt.Errorf("failed to copy UI to %s: %w", plat.Key(), err)
		}
	}
	fmt.Println("✅ Built and distributed UI assets")
	return nil
}
