package scheduler

import (
	"github.com/go-co-op/gocron"
	"time"
)

// init registers the "gocron" provider, which is an awesome scheduler.
// Please check out the github repo: (github.com/go-co-op/gocron).
func init() {
	err := Register(GoCron, fn)
	if err != nil {
		panic(err)
	}
}

// fn creates a new Provider.
func fn(option interface{}) (Provider, error) {
	// TODO set options for timezone.
	cron := &cron{}
	cron.scheduler = gocron.NewScheduler(time.UTC)

	return cron, nil
}

type cron struct {
	scheduler   *gocron.Scheduler
	currentName string
	names       []string
}

type job struct {
	job  *gocron.Job
	name string
}

//------------------------------
// Provider interface
//------------------------------

// Every satisfy the Provider interface.
func (c *cron) Start() {
	c.scheduler.StartAsync()
}

// Every satisfy the Provider interface.
func (c *cron) Stop() {
	c.scheduler.Stop()
}

// Every satisfy the Provider interface.
func (c *cron) Status() string {
	if c.scheduler.IsRunning() {
		return StatusRunning
	}
	return StatusNotRunning
}

// Every satisfy the Provider interface.
func (c *cron) Jobs() []ProviderJobDetail {

	scheduledJobs := c.scheduler.Jobs()
	jobs := make([]ProviderJobDetail, len(scheduledJobs))

	for i, j := range scheduledJobs {
		jobs[i] = &job{job: j, name: c.names[i]}
	}

	return jobs
}

//------------------------------
// ProviderJob interface
//------------------------------

// Every satisfy the ProviderJob interface.
func (c *cron) Every(interval interface{}) ProviderJob {
	c.scheduler.Every(interval)
	c.currentName = ""
	return c
}

// Second satisfy the ProviderJob interface.
func (c *cron) Second() ProviderJob {
	c.scheduler.Second()
	return c
}

// Minute satisfy the ProviderJob interface.
func (c *cron) Minute() ProviderJob {
	c.scheduler.Minute()
	return c
}

// Day satisfy the ProviderJob interface.
func (c *cron) Day() ProviderJob {
	c.scheduler.Day()
	return c
}

// Monday satisfy the ProviderJob interface.
func (c *cron) Monday() ProviderJob {
	c.scheduler.Monday()
	return c
}

// Tuesday satisfy the ProviderJob interface.
func (c *cron) Tuesday() ProviderJob {
	c.scheduler.Tuesday()
	return c
}

// Wednesday satisfy the ProviderJob interface.
func (c *cron) Wednesday() ProviderJob {
	c.scheduler.Wednesday()
	return c
}

// Thursday satisfy the ProviderJob interface.
func (c *cron) Thursday() ProviderJob {
	c.scheduler.Thursday()
	return c
}

// Friday satisfy the ProviderJob interface.
func (c *cron) Friday() ProviderJob {
	c.scheduler.Friday()
	return c
}

// Saturday satisfy the ProviderJob interface.
func (c *cron) Saturday() ProviderJob {
	c.scheduler.Saturday()
	return c
}

// Sunday satisfy the ProviderJob interface.
func (c *cron) Sunday() ProviderJob {
	c.scheduler.Sunday()
	return c
}

// Week satisfy the ProviderJob interface.
func (c *cron) Week() ProviderJob {
	c.scheduler.Week()
	return c
}

// At satisfy the ProviderJob interface.
func (c *cron) At(at string) ProviderJob {
	c.scheduler.At(at)
	return c
}

// Name satisfy the ProviderJob interface.
func (c *cron) Name(name string) ProviderJob {
	c.currentName = name
	return c
}

// Month satisfy the ProviderJob interface.
func (c *cron) Month(dayOfMonth ...int) ProviderJob {
	day := 1
	if len(dayOfMonth) == 1 {
		day = dayOfMonth[0]
	}
	c.scheduler.Month(day)
	return c
}

// Tag satisfy the ProviderJob interface.
func (c *cron) Tag(t ...string) ProviderJob {
	c.scheduler.Tag(t...)
	return c
}

// Singleton satisfy the ProviderJob interface.
func (c *cron) Singleton() ProviderJob {
	c.scheduler.SingletonMode()
	return c
}

// Do satisfy the ProviderJob interface.
func (c *cron) Do(jobFun interface{}, params ...interface{}) error {
	c.names = append(c.names, c.currentName)
	_, err := c.scheduler.Do(jobFun, params...)
	return err
}

//------------------------------
// ProviderJobDetail interface
//------------------------------

// Name satisfy the ProviderJobDetail interface.
func (j *job) Name() string {
	return j.name
}

// Counter satisfy the ProviderJobDetail interface.
func (j *job) Counter() int {
	return j.job.RunCount()
}

// Tags satisfy the ProviderJobDetail interface.
func (j *job) Tags() []string {
	return j.job.Tags()
}

// LastRun satisfy the ProviderJobDetail interface.
func (j *job) LastRun() time.Time {
	return j.job.LastRun()
}

// NextRun satisfy the ProviderJobDetail interface.
func (j *job) NextRun() time.Time {
	return j.job.NextRun()
}
