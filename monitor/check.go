package monitor

import (
	"context"

	"encore.app/site"
	"encore.dev/cron"
	"encore.dev/storage/sqldb"
	"golang.org/x/sync/errgroup"
)

// Check checks a single site.
//
//encore:api public method=POST path=/check/:siteID
func Check(ctx context.Context, siteID int) error {
	site, err := site.Get(ctx, siteID)
	if err != nil {
		return err
	}
	return check(ctx, site)
}

// CheckAll checks all sites.
//
//encore:api public method=POST path=/checkall
func CheckAll(ctx context.Context) error {
	// Get all the tracked sites.
	resp, err := site.List(ctx)
	if err != nil {
		return err
	}

	// Check up to 8 sites concurrently.
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for _, site := range resp.Sites {
		site := site // capture for closure
		g.Go(func() error {
			return check(ctx, site)
		})
	}
	return g.Wait()
}

func check(ctx context.Context, site *site.Site) error {
	result, err := Ping(ctx, site.URL)
	if err != nil {
		return err
	}

	// Publish a Pub/Sub message if the site transitions
	// from up->down or from down->up.
	if err := publishOnTransition(ctx, site, result.Up); err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO checks (site_id, up, checked_at)
		VALUES ($1, $2, NOW())
	`, site.ID, result.Up)
	return err
}

// Check all tracked sites every 1 hour.
var _ = cron.NewJob("check-all", cron.JobConfig{
	Title:    "Check all sites",
	Endpoint: CheckAll,
	Every:    1 * cron.Hour,
})

// Define a database named 'monitor', using the database migrations
// in the "./migrations" folder. Encore automatically provisions,
// migrates, and connects to the database.
var db = sqldb.NewDatabase("monitor", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
