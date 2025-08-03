package llm

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"log/slog"
	"reflect"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SlogHandler implements slog.Handler interface to bridge slog to zap logger
type SlogHandler struct {
	logger log.Logger
	attrs  []slog.Attr
	groups []string
}

// NewSlogHandler creates a new slog.Handler that outputs to the given zap logger
func NewSlogHandler() *SlogHandler {
	return &SlogHandler{
		logger: log.WithName("Genkit"),
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

// Enabled returns whether the handler handles records at the given level
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Convert slog level to zap level and check if enabled
	zapLevel := slogLevelToZapLevel(level)
	return log.ZapLogger().Core().Enabled(zapLevel)
}

// Handle handles the Record
func (h *SlogHandler) Handle(_ context.Context, record slog.Record) error {
	// Convert slog attributes to zap fields
	fields := make([]zap.Field, 0, record.NumAttrs()+len(h.attrs))

	// Add handler-level attributes
	for _, attr := range h.attrs {
		fields = append(fields, slogAttrToZapField(attr, h.groups))
	}

	// Add record-level attributes
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, slogAttrToZapField(attr, h.groups))
		return true
	})

	// Log the message based on level
	switch record.Level {
	case slog.LevelDebug:
		h.logger.Debug(record.Message, fields...)
	case slog.LevelInfo:
		h.logger.Info(record.Message, fields...)
	case slog.LevelWarn:
		h.logger.Warn(record.Message, fields...)
	case slog.LevelError:
		h.logger.Error(record.Message, fields...)
	default:
		h.logger.Info(record.Message, fields...)
	}

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &SlogHandler{
		logger: h.logger,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups
func (h *SlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &SlogHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// slogLevelToZapLevel converts slog.Level to zapcore.Level
func slogLevelToZapLevel(level slog.Level) zapcore.Level {
	switch {
	case level < slog.LevelInfo:
		return zapcore.DebugLevel
	case level < slog.LevelWarn:
		return zapcore.InfoLevel
	case level < slog.LevelError:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}

// slogAttrToZapField converts slog.Attr to zap.Field
func slogAttrToZapField(attr slog.Attr, groups []string) zap.Field {
	key := attr.Key

	// Add group prefixes to the key
	for _, group := range groups {
		key = group + "." + key
	}

	value := attr.Value
	switch value.Kind() {
	case slog.KindString:
		return zap.String(key, value.String())
	case slog.KindInt64:
		return zap.Int64(key, value.Int64())
	case slog.KindUint64:
		return zap.Uint64(key, value.Uint64())
	case slog.KindFloat64:
		return zap.Float64(key, value.Float64())
	case slog.KindBool:
		return zap.Bool(key, value.Bool())
	case slog.KindDuration:
		return zap.Duration(key, value.Duration())
	case slog.KindTime:
		return zap.Time(key, value.Time())
	case slog.KindGroup:
		// Handle groups by converting to a string representation
		// to avoid JSON serialization issues with complex types
		return zap.String(key, value.String())
	default:
		// Handle unknown types safely, including function types
		return handleUnknownType(key, value)
	}
}

// handleUnknownType safely handles unknown slog.Value types, including functions
func handleUnknownType(key string, value slog.Value) zap.Field {
	// Get the underlying value from slog.Value
	anyValue := value.Any()

	// Use reflection to check if it's a function
	if anyValue != nil {
		rv := reflect.ValueOf(anyValue)
		if rv.Kind() == reflect.Func {
			// If it's a function, try to get the function name
			if fn := runtime.FuncForPC(rv.Pointer()); fn != nil {
				funcName := fn.Name()
				return zap.String(key, "func:"+funcName)
			}
			// Fallback for functions without name info
			return zap.String(key, "func:<anonymous>")
		}
	}

	// For non-function types, use the string representation
	// This avoids "json: unsupported type" errors for other complex types
	return zap.String(key, value.String())
}

// DefaultSlogHandler returns a slog.Handler that uses the default logger
func DefaultSlogHandler() slog.Handler {
	return NewSlogHandler()
}

// SetupGenkitLogger configures the default slog logger to use our zap logger
func SetupGenkitLogger() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	handler := DefaultSlogHandler()
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
