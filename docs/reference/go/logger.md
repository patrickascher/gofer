# Logger

The package logger provides an interface for logging. It wraps awesome existing go loggers with that interface. In that
case, it is easy to change the log provider without breaking anything in your application.

Additionally log level, fields, time duration or caller information can be added.

### Register

To register a logger by your requirements, simple use the `logger.Register` function. The registered logger must
implement the `logger.Provider` interface.

This should be done in the `init()` function or any other early stage in your application.

For a list of all existing provider see [providers section](logger.md#providers).

```go 
// Example of the memory cache provider
logrus := logrus.New()
// the wrapped logrus instance can be configured as needed.
logrus.Instance.ReportCaller = true

// register the logger under a defined name
err = logrus.Register("importLogger",logrus)
//...
```

### Usage

Use `Get()` to get a registered logger instance.

```go 
// get the logger
log,err := logger.Get("importLogger")

// set some default logger settings.
log.SetCallerFields(true)
log.SetLogLevel(logger.TRACE)

// creates a new info log entry.
log.Info("something")

// log with some details
log.WithFields(logger.Fields{"foo":"bar"}).Info("something")

// log with timer
ltimer := log.WithTimer()
//.. some logic
ltimer.Debug("some time") //Field "duration" is added with the required time.
```

### SetCallerFields

If you call `SetCallerFields` on a logger instance, it will add the file name and line number as fields. Default it is
set to `false`. The caller will be set global for this logger instance.

```go 
log.SetCallerFields(true)
```

### SetLogLevel

Defines the level where it should start logging. Available log levels are `logger.TRACE`,`logger.DEBUG`,`logger.INFO`
,`logger.WARNING`,`logger.ERROR` and `logger.PANIC`. Default it is set to `logger.DEBUG`. The level will be set global
for this logger instance.

```go 
log.SetLogLevel(logger.INFO)
```

### Log Level

The following log levels are available.

```go 
log.Trace("msg")
log.Debug("msg")
log.Info("msg")
log.Warning("msg")
log.Error("msg")
log.Panic("msg")
```

### WithFields

Sometimes a log message is not enough and some additional information is required.

!!! info

    `WithFields` will create a new instance.

```go 
log.WithFields(logger.Fields{"foo":"bar"}).Info("msg")
```

### WithTimer

Sometimes it is useful for debugging to see the required time of a function.
`WithTimer` can be combined `WithFields` or vis-a-vis.

!!! info

    `WithTimer` will create a new instance.

```go 
ltimer := log.WithTimer()
//...
ltimer.Debug("msg") // a Field "duration" with the required time is added.

// combinde with fields
ltimer = log.WithFields(logger.Fields("foo":"bar")).WithTimer()
//...
ltimer.Debug("msg")// a Field "duration" and "foo" is added.
```

### New

Sometimes its useful to create a logger with slightly a different configuration. The new instance will inherit the
setting of the parent logger by default.

```go 
log,err := logger.Get("importLogger")
log.SetCallerFields(true)
log.SetLogLevel(logger.WARNING)

// create a new instance and add different settings.
log2 := log.New()
log2.SetCallFields(false)
log2.SetLogLevel(logger.TRACE)
```

## Providers

All pre-defined providers:

### Logrus

A wrapper for [logrus](https://github.com/sirupsen/logrus).

Package:

`github.com/patrickascher/gofer/logger/logrus`

Options:

The original logrus struct can be accessed by the `Instance` field. Please check the github page for the documentation.

Usage:

```go 
import "github.com/patrickascher/gofer/logger"
import "github.com/patrickascher/gofer/logger/logrus"

logrusProvider := logrus.New()
logrusProvider.Instance.ReportCaller = true

// register the logger
err := logger.Register("logger-name", logrusProvider)

// somewhere in the application
log,err := logger.Get("logger-name")

```

## Create your own provider

To create your own provider, you have to implement the `logger.Provider` interface.

```go 
type Provider interface {
	Log(Entry)
}
```

The registered value must be of the type `logger.Provider`.

```go 

type MyLogger struct{
}

func (ml *MyLogger) Log(Entry){
    // ... do something
}

// register the logger
err := logger.Register("my-logger", &MyLogger{})

// somewhere in the application
log,err := logger.Get("my-logger")
```

### Entry

The `logger.Entry` holds all information about the log message.

```go 
type Entry struct {
	Level     level
	Timestamp time.Time
	Message   string
	Fields    Fields
}
```
