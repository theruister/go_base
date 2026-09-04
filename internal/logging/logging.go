package logging

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	LOGGER_CONFIG_MODULE        = "config"
	LOGGER_GRPC_MODULE          = "grpc"
	LOGGER_GRPC_INTERNAL_MODULE = "grpc-internal"
	LOGGER_API_MODULE           = "api"
	LOGGER_METRIC_MODULE        = "metric"
	LOGGER_DATASTORE_MODULE     = "datastore"
	LOGGER_SYSTEM_MODULE        = "system"
	LOGGER_EVENT_MODULE         = "event"

	FileLogBaseName = `f:\tmp\`
)

// alias the logrus specific field definitions
type Fields = log.Fields
type Logger = log.Logger
type Level = log.Level

const (
	// Mapping to logrus
	Panic = log.PanicLevel
	Fatal = log.FatalLevel
	Error = log.ErrorLevel
	Warn  = log.WarnLevel
	Info  = log.InfoLevel
	Debug = log.DebugLevel
	Trace = log.TraceLevel
)

// package wide variables
var defaultLogLevel = Warn
var moduleLogMap map[string]*Logger
var logLock = &sync.RWMutex{}

func GetValidModules() []string {
	return []string{LOGGER_CONFIG_MODULE, LOGGER_GRPC_MODULE, LOGGER_GRPC_INTERNAL_MODULE, LOGGER_API_MODULE, LOGGER_METRIC_MODULE,
		LOGGER_DATASTORE_MODULE, LOGGER_SYSTEM_MODULE, LOGGER_EVENT_MODULE}
}

type ModuleLogLvl struct {
	ModuleName string `json:"moduleName"`
	LogLevel   string `json:"logLevel"`
}

// init method
func init() {
	// Initialize the formatter to be json
	log.SetFormatter(&log.JSONFormatter{})

	// Create the map
	moduleLogMap = make(map[string]*log.Logger, 0)

	// Define all the acceptable module names
	modules := GetValidModules()

	for _, module := range modules {
		alogger, err := createModuleLogger(module)
		if err != nil {
			// error creating - just assign to system
			moduleLogMap[module] = log.StandardLogger()
		} else {
			moduleLogMap[module] = alogger
		}
	}
}

func createModuleLogger(module string) (*Logger, error) {
	alogger := log.New()
	alogger.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg: "message",
		},
	})
	alogger.SetLevel(defaultLogLevel)
	imhook, _ := NewIMModuleHook(module)
	alogger.AddHook(imhook)

	return alogger, nil
}

// Helper Methods
func GetCurrentLogLevels() []ModuleLogLvl {
	modLevels := make([]ModuleLogLvl, 0)

	modules := GetValidModules()

	for _, module := range modules {
		var modLvl ModuleLogLvl

		modLvl.ModuleName = module
		modLvl.LogLevel = GetModuleLogLevelString(module)

		modLevels = append(modLevels, modLvl)
	}

	return modLevels
}

func IsValidLevel(level string) bool {
	_, err := log.ParseLevel(level)
	if err != nil {
		return false
	}

	return true
}

func IsValidModule(module string) bool {
	if module == "all" {
		return true
	}

	if _, ok := moduleLogMap[module]; ok {
		return true
	}

	return false
}

func ResetAllLogLevels(level Level) error {
	modules := GetValidModules()

	for _, module := range modules {
		err := SetModuleLogLevel(module, level)
		if err != nil {
			return err
		}
	}

	return nil
}

func ResetAllLogLevelsByString(level string) error {
	modules := GetValidModules()

	for _, module := range modules {
		err := SetModuleLogLevelByString(module, level)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetModuleLogger(module string) *Logger {
	pLogger, ok := moduleLogMap[module]
	if ok {
		return pLogger
	}

	// not found, return system default logger
	return log.StandardLogger()
}

func GetModuleLogLevel(module string) Level {
	pLogger, ok := moduleLogMap[module]
	if ok {
		return pLogger.GetLevel()
	}

	return log.GetLevel()
}

func GetModuleLogLevelString(module string) string {
	pLogger, ok := moduleLogMap[module]
	if ok {
		return pLogger.GetLevel().String()
	}

	return log.GetLevel().String()
}

func SetModuleLogLevel(module string, level Level) error {
	if !IsValidModule(module) {
		return fmt.Errorf("Invalid logging module")
	}

	logLock.Lock()
	defer logLock.Unlock()

	if module == "all" {
		// Set each module
		for _, pLogger := range moduleLogMap {
			pLogger.SetLevel(level)
		}

		// also set the default
		log.SetLevel(level)
	} else {
		pLogger, ok := moduleLogMap[module]
		if ok {
			pLogger.SetLevel(level)
		}
	}

	return nil
}

func SetModuleLogLevelByString(module string, level string) error {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	if !IsValidModule(module) {
		return fmt.Errorf("Invalid logging module")
	}

	logLock.Lock()
	defer logLock.Unlock()

	if module == "all" {
		//Set each mocule
		for _, pLogger := range moduleLogMap {
			pLogger.SetLevel(lvl)
		}

		// also set the default
		log.SetLevel(lvl)

	} else {
		pLogger, ok := moduleLogMap[module]
		if ok {
			pLogger.SetLevel(lvl)
		}
	}

	return nil
}

func ModuleLevelMapToString() string {
	logLock.RLock()
	defer logLock.RUnlock()

	var sb strings.Builder

	for module, pLogger := range moduleLogMap {
		fmt.Fprintf(&sb, "%s : %s", module, pLogger.GetLevel().String())
	}

	return sb.String()
}

// Logrus specific Hooks
// Add Hook to include module name with every log message
type IMModuleHook struct {
	Module string
}

func (hook *IMModuleHook) Levels() []Level {
	return log.AllLevels
}

func (hook *IMModuleHook) Fire(pEntry *log.Entry) error {
	// Add in the module name as a field to all logging
	if _, ok := pEntry.Data["module"]; !ok {
		pEntry.Data["module"] = hook.Module
	}

	return nil
}

func NewIMModuleHook(module string) (*IMModuleHook, error) {
	return &IMModuleHook{module}, nil
}
