package cronc

import (
	"flag"
	"fmt"
	"inventory-movement-processing/pkg/logger"
	sctx "inventory-movement-processing/pkg/service_context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	defaultTimeZone = "Asia/Ho_Chi_Minh"
)

// JobFunc is the function signature for a cron job handler.
type JobFunc func()

// JobDefinition holds a cron expression and its handler.
type JobDefinition struct {
	Name     string
	Schedule string // 6-field cron: "second minute hour day month weekday"
	Handler  JobFunc
}

// CronComponent is the public interface exposed to the rest of the application.
// Retrieve it from ServiceContext via MustGet(id).
type CronComponent interface {
	sctx.Component

	// AddJob registers a job. Safe to call before or after Activate.
	AddJob(job JobDefinition) error

	// GetCron returns the underlying *cron.Cron for advanced use (e.g. RemoveJob).
	GetCron() *cron.Cron
}

// ── internal implementation ──────────────────────────────────────────────────

type Config struct {
	timeZone string
}

type cronComponent struct {
	*Config
	id       string
	logger   logger.Logger
	cron     *cron.Cron
	jobs     []JobDefinition
	entryIDs []cron.EntryID
	mu       sync.Mutex
}

// NewCron creates a new CronComponent with the given service-context ID.
func NewCron(id string) CronComponent {
	return &cronComponent{
		Config: new(Config),
		id:     id,
	}
}

// ── CronComponent interface ──────────────────────────────────────────────────

func (c *cronComponent) AddJob(job JobDefinition) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.jobs = append(c.jobs, job)

	// If the scheduler is already running, register immediately.
	if c.cron != nil {
		return c.scheduleJob(job)
	}
	return nil
}

func (c *cronComponent) GetCron() *cron.Cron {
	return c.cron
}

// ── sctx.Component interface ─────────────────────────────────────────────────

func (c *cronComponent) ID() string { return c.id }

func (c *cronComponent) InitFlags() {
	flag.StringVar(
		&c.timeZone,
		"cron-timezone",
		defaultTimeZone,
		"timezone for cron scheduler (e.g. UTC, Asia/Ho_Chi_Minh)",
	)
}

func (c *cronComponent) Activate(serviceContext sctx.ServiceContext) error {
	c.logger = serviceContext.Logger(c.id)
	c.logger.Infof("init cron scheduler timezone=%s", c.timeZone)

	loc, err := time.LoadLocation(c.timeZone)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", c.timeZone, err)
	}

	c.cron = cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(), // 6-field: second minute hour day month weekday
		cron.WithLogger(cron.DefaultLogger),
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, job := range c.jobs {
		if err := c.scheduleJob(job); err != nil {
			return err
		}
	}

	c.cron.Start()
	c.logger.Infof("cron scheduler started with %d job(s)", len(c.jobs))
	return nil
}

func (c *cronComponent) Stop() error {
	if c.cron != nil {
		c.logger.Info("stopping cron scheduler...")
		<-c.cron.Stop().Done() // wait for any running jobs to finish
		c.logger.Info("cron scheduler stopped")
	}
	return nil
}

// ── private helpers ──────────────────────────────────────────────────────────

func (c *cronComponent) scheduleJob(job JobDefinition) error {
	entryID, err := c.cron.AddFunc(job.Schedule, func() {
		c.logger.Infof("running job: %s", job.Name)
		job.Handler()
	})
	if err != nil {
		return fmt.Errorf("schedule job %q: %w", job.Name, err)
	}
	c.entryIDs = append(c.entryIDs, entryID)
	c.logger.Infof(
		"registered job: name=%s schedule=%q entry_id=%d",
		job.Name, job.Schedule, entryID,
	)
	return nil
}
