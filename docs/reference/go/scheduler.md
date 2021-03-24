# Scheduler

Package scheduler provides a job scheduler for periodically functions.

## New

Will create a new scheduler by the given provider.

The second param will be the option parameter. It depends on the provider which options are available. Please see
the [provider](#provider) section.

The given times, are in `time.UTC` by default.

```go
s, err := scheduler.New(scheduler.GoCron, nil)
if err != nil {
panic(err)
}
err = s.Every(1).Minute().Name("Test").Tag("Import").Do(func (){
fmt.Println("croncall:")
})
s.Start() 
```

### Start

Will start the scheduler executor in a new routine.

```go
s.Start() 
```

### Stop

Will stop the scheduler task runner.

```go
s.Stop() 
```

### Status

Will return the value `Scheduler is running!` or `Scheduler is not running!`.

```go
status := s.Status() 
```

### Jobs

Will return all registered jobs. See [Job details](#details) for the available methods.

```go
jobs := s.Jobs()
//...
```

#### Details

| Name                 | Description      |
|----------------------|------------------|
| Name()           | Will return the name of the job. Will be empty if none was set.|
| Counter()           | Will return an `int` with the number of runs.|
| Tags()           | Returns a `[]string` with all the given tags. Will be empty if none was set.|
| LastRun()           | Returns a `time.Time` for the last run.|
| NextRun()           | Returns a `time.Time` for the next run.|

```go
jobs := s.Jobs()

for _,job := rang jobs{
name := job.Name()
counter := job.Counter()
lastRun := job.LastRun()
//...
}
```

### Every

Will register a new Job with the given interval.  Interval can be an `int`, `time.Duration` or a `string` that parses with time.ParseDuration(). Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

See [Job methods](#methods) for the available methods.

#### Methods

On Every there are the following methods available.

| Name                 | Description      |
|----------------------|------------------|
| Second()               | Second will be set as unit.                 |
| Minute()               | Minute will be set as unit.                 |
| Day()               |   Day will be set as unit.               |
| Monday()               |  Will set the unit to week.                |
| Tuesday()               | Will set the unit to week.                  |
| Wednesday()               | Will set the unit to week.                  |
| Thursday()               | Will set the unit to week.                  |
| Friday()               | Will set the unit to week.                  |
| Saturday()               | Will set the unit to week.                  |
| Sunday()               | Will set the unit to week.                  |
| Week()               |  Week will be set as unit. TODO: needed? its actually a syn for Monday.               |
| Month(...int)               | Month will be set as unit. If no day is given, 1 will be set as default.               |
| At              | Will define the time when to run. It is only available on `Day`, `Week` and `Month` Formant (HH:MM) or (HH:MM:SS)   |
| | | 
| Name(string)           | Sets the job name. | 
| Tag(...string)           | Adds tag(s) to the job. | 
|Singleton()           | Singleton will not spawn a new job if the old one is not finished yet. The job will be re-scheduled for the next run. |
| | | 
| Do(interface{},...interface{})           | Do defines the function which should be called. Parameter can be added. |

```go

// Every minute
err := s.Every(1).Minute().Name("Test").Tag("Import").Do(func (){})
// Every hour
err = s.Every(1).Hour().Name("Test").Tag("Import").Do(func (){})
// Every day (default 00:00:00)
err = s.Every(1).Day().Name("Test").Tag("Import").Do(func (){})
// Every second day at 10:30:05
err = s.Every(2).Day().At("10:30:05").Name("Test").Tag("Import").Do(func (){})
// Every week (default monday)
err = s.Every(1).Week().At("10:30:05").Name("Test").Tag("Import").Do(func (){})
// Every week on wednesday
err = s.Every(1).Wednesday().Name("Test").Tag("Import").Do(func (){})
// Every month (default 1.)
err = s.Every(1).Month().Name("Test").Tag("Import").Do(func (){})
// Every month at 15. at 10:30
err = s.Every(1).Month(15).At("10:30").Name("Test").Tag("Import").Do(func (){})
```
