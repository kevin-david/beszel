// fork-config edits local monitoring configuration while the hub is stopped.
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	dataDir := flag.String("data-dir", "", "Existing PocketBase data directory")
	systemName := flag.String("system", "virusdesk", "Monitored system name")
	percent := flag.Float64("memory-percent", 90, "Memory warning percentage")
	minutes := flag.Int("minutes", 2, "Minutes above threshold")
	offline := flag.Bool("offline", false, "Confirm that the hub using this directory is stopped")
	flag.Parse()
	if !*offline || !filepath.IsAbs(*dataDir) || *percent <= 0 || *percent > 100 || *minutes < 1 || *minutes > 255 {
		fmt.Fprintln(os.Stderr, "Require --offline, an absolute --data-dir, a valid percentage, and 1-255 minutes.")
		os.Exit(2)
	}
	if _, err := os.Stat(filepath.Join(*dataDir, "data.db")); err != nil {
		fmt.Fprintln(os.Stderr, "Existing data.db is required.")
		os.Exit(2)
	}
	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: *dataDir})
	if err := configure(app, *systemName, *percent, *minutes); err != nil {
		fmt.Fprintln(os.Stderr, "Configuration failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Memory warning saved: %s above %.1f%% for %d minutes. Start the hub to load it.\n", *systemName, *percent, *minutes)
}

func configure(app *pocketbase.PocketBase, systemName string, percent float64, minutes int) error {
	if err := app.Bootstrap(); err != nil {
		return err
	}
	defer app.ClearBootstrap()
	systems, err := app.FindRecordsByFilter("systems", "name={:name}", "", 2, 0, dbx.Params{"name": systemName})
	if err != nil {
		return err
	}
	if len(systems) != 1 {
		return errors.New("system name must identify exactly one system")
	}
	system := systems[0]
	users := system.GetStringSlice("users")
	if len(users) != 1 {
		return errors.New("configuration requires a system with exactly one owner")
	}
	alert, err := app.FindFirstRecordByFilter("alerts", "system={:system} && user={:user} && name='Memory'", dbx.Params{"system": system.Id, "user": users[0]})
	if errors.Is(err, sql.ErrNoRows) {
		collection, err := app.FindCollectionByNameOrId("alerts")
		if err != nil {
			return err
		}
		alert = core.NewRecord(collection)
	} else if err != nil {
		return err
	}
	alert.Set("system", system.Id)
	alert.Set("user", users[0])
	alert.Set("name", "Memory")
	alert.Set("value", percent)
	alert.Set("min", minutes)
	if err := app.Save(alert); err != nil {
		return err
	}
	// Verify via the database, not the object that was just written.
	stored, err := app.FindRecordById("alerts", alert.Id)
	if err != nil {
		return err
	}
	if stored.GetFloat("value") != percent || stored.GetInt("min") != minutes {
		return errors.New("stored memory warning differs from the requested settings")
	}
	return nil
}
