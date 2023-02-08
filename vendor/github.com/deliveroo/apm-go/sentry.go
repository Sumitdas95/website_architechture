package apm

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// alert sends an error to Sentry and logs the error at ErrorLevel.
func (t *tracer) alert(skipCallers int, err error, user *sentry.User, fields ...zap.Field) {
	t.captureException(3+skipCallers, nil, err, user, fields...)
}

func (t *tracer) reportSentrySpan(skipCallers int, s *Span) {
	s.RLock()
	defer s.RUnlock()
	if s.alertable && s.err != nil {
		t.captureException(skipCallers+6, s, s.err, nil)
	}
	for _, child := range s.children {
		t.reportSentrySpan(skipCallers+1, child)
	}
}

func (t *tracer) captureException(skipCallers int, s *Span, err error, user *sentry.User, fields ...zap.Field) {
	if err == nil {
		return
	}
	if t.isDevelopment {
		fields = append(fields, zap.Error(err))
	}
	if s != nil {
		fields = append(fields,
			zap.Namespace("span"),
			zap.String("name", s.name),
			zap.String("resource", s.resource),
			zap.Uint64("trace-id", s.traceID),
		)
		if s.parentID != 0 {
			fields = append(fields, zap.Uint64("parent-id", s.parentID))
		}

		var rawFields []zap.Field

		s.meta.Range(func(key, value interface{}) bool {
			rawFields = append(rawFields, zap.String(key.(string), value.(string)))
			return true
		})

		if len(rawFields) > 0 {
			fields = append(fields, zap.Namespace("meta"))
			fields = append(fields, rawFields...)
		}
	}

	if t.sentry != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			for k, v := range flattenFields(fields) {
				scope.SetTag(k, v)
			}
			if user != nil {
				scope.SetUser(*user)
			}
			t.sentry.CaptureException(err, &sentry.EventHint{OriginalException: err}, scope)
		})
	}
	if user != nil {
		fields = append(fields,
			zap.Namespace("sentry.user"),
			zap.String("email", obfuscate(user.Email)),
			zap.String("id", user.ID),
			zap.String("ip-address", obfuscate(user.IPAddress)),
			zap.String("username", user.Username),
		)
	}
	if t.isDevelopment {
		t.getLogger(skipCallers).Error("sentry:alert", fields...)
	} else {
		t.getLogger(skipCallers).Error(err.Error(), fields...)
	}
}

// FIXME That function MAY enter an infinite loop if the fields contain a cycle.
func extractFields(prefix string, fields map[string]interface{}, result map[string]string) {
	for k, v := range fields {
		if subfields, ok := v.(map[string]interface{}); ok {
			extractFields(k, subfields, result)
		} else {
			key := k
			if prefix != "" {
				key = prefix + "." + key
			}
			result[key] = fmt.Sprintf("%v", v)
		}
	}
}

func flattenFields(fields []zap.Field) map[string]string {
	encoder := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(encoder)
	}
	flattened := make(map[string]string)
	extractFields("", encoder.Fields, flattened)
	return flattened
}

func obfuscate(s string) string {
	omit := 0
	length := utf8.RuneCountInString(s)
	if length == 0 {
		return s
	}
	if length > 16 {
		omit = length - 16
	}
	cut := length / 8
	if cut > 4 {
		cut = 4
	}
	i := 0
	right := length - cut
	return strings.Map(func(r rune) rune {
		out := '*'
		if i < cut || i >= right {
			out = r
		} else if omit > 0 {
			omit--
			out = rune(-1)
		}
		i++
		return out
	}, s)
}
