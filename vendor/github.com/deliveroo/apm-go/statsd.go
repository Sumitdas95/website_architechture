//go:generate mockery --inpackage --name StatsDClientInterface --dir .

package apm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"go.uber.org/zap"
)

// StatsDClientInterface simply shadows the datadog SDK's statsd.ClientInterface so that we can generate a mock easily
type StatsDClientInterface interface {
	statsd.ClientInterface
}

// StatType is a type of StatsD Metric.
type StatType string

const (
	// CountStatType is a StatsD count metric.
	CountStatType StatType = "count"

	// DistributionStatType is a StatsD distribution metric.
	DistributionStatType StatType = "distribution"

	// GaugeStatType is a StatsD gauge metric.
	GaugeStatType StatType = "gauge"

	// HistogramStatType is a StatsD histogram metric.
	HistogramStatType StatType = "histogram"

	// IncrStatType is a StatsD count metric of 1.
	IncrStatType StatType = "incr"

	// TimingStatType is a StatsD timing metric.
	TimingStatType StatType = "timing"

	// EventStatType is a StatsD event metric.
	EventStatType StatType = "event"
)

// StatsDMetric represents a StatsD metric.
type StatsDMetric struct {
	Type StatType

	DurationValue time.Duration // DurationValue is a timing value.
	FloatValue    float64       // FloatValue is a gauge or histogram value.
	IntValue      int64         // IntValue is the count or incr value.
	Name          string        // Name is the metric name or event title as it would appear in DataDog.
	Rate          float64       // Rate is the rate value.
	Tags          []string      // Tags are the tags in the form "tag:value".
}

// EventOptions are optional parameters to give an event more detail.
type EventOptions struct {
	// Timestamp is a timestamp for the event.  If not provided, the dogstatsd
	// server will set this to the current time.
	Timestamp time.Time
	// Hostname for the event.
	Hostname string
	// AggregationKey groups this event with others of the same key.
	AggregationKey string
	// SourceTypeName for integrations provide the optional source.
	// List of integrations here: https://docs.datadoghq.com/integrations/faq/list-of-api-source-attribute-value/
	SourceTypeName string
	// IsLowPriority
	IsLowPriority bool
	// AlertType options: [info|error|warning|success]  default is 'info'
	AlertType string
	// TagPairs in the form of alternating key values e.g. []string{"deployment", "green", "variant" : "experiment25"}
	TagPairs []string
}

// String returns the StatsD metric format, roughly like
// "namespace.timing:6.000000|ms|@0.1|#tag1:val1,tag2:val2". It's potentially
// slow and meant for debugging and isn't representative of the full StatsD
// metric format specification.
func (m StatsDMetric) String() string {
	rate := ""
	if m.Rate != 1 {
		rate = fmt.Sprintf("|@%f", m.Rate)
	}
	var v string
	switch m.Type {
	case TimingStatType:
		v = strconv.FormatInt(m.DurationValue.Milliseconds(), 10) + "|ms"
	case GaugeStatType, HistogramStatType:
		v = fmt.Sprintf("%f", m.FloatValue)
	default:
		v = strconv.FormatInt(m.IntValue, 10)
	}
	tags := strings.Join(m.Tags, ",")
	return fmt.Sprintf("%s.%s:%s%s|#%s", m.Name, m.Type, v, rate, tags)
}

// Metrics provides StatsD metric methods like counters and gauges.
//
// The signatures are purposefully different from their apparent equivalents
// in statsd.ClientInterface, because the features implemented are not the same,
// with the Metrics concrete implementation liable to emit logs or publish to a
// channel along with the statsd emission, which original ClientInterface
// implementations will not do, meaning it is better to leave Metrics not be a
// subset of ClientInterface, to avoid the risk of callers mixing those.
type Metrics interface {
	// Count tracks how many times something happened per second.
	Count(name string, value int64, rate float64, tagPairs ...string)

	// Distribution tracks the statistical distribution of a set of values across your infrastructure.
	Distribution(name string, value float64, rate float64, tagPairs ...string)

	// Event sends a DataDog event.
	Event(title, text string, options EventOptions)

	// Gauge measures the value of a metric at a particular time.
	Gauge(name string, value float64, rate float64, tagPairs ...string)

	// Histogram tracks the statistical distribution of a set of values on each host.
	Histogram(name string, value float64, rate float64, tagPairs ...string)

	// Incr is a Count of 1.
	Incr(name string, rate float64, tagPairs ...string)

	// Timing sends timing information in milliseconds. It is flushed by
	// statsd with percentiles, mean and other info. See
	// https://github.com/etsy/statsd/blob/master/docs/metric_types.md#timing
	Timing(name string, value time.Duration, rate float64, tagPairs ...string)
}

// Count tracks how many times something happened per second.
func (t *tracer) Count(name string, value int64, rate float64, tagPairs ...string) {
	t.doCount(name, value, rate, tagPairs...)
}

// Distribution tracks the statistical distribution of a set of values across your infrastructure.
func (t *tracer) Distribution(name string, value float64, rate float64, tagPairs ...string) {
	t.doDistribution(name, value, rate, tagPairs...)
}

// Event sends a DataDog event.
func (t *tracer) Event(title, text string, options EventOptions) {
	t.doEvent(title, text, options)
}

// Gauge measures the value of a metric at a particular time.
func (t *tracer) Gauge(name string, value float64, rate float64, tagPairs ...string) {
	t.doGauge(name, value, rate, tagPairs...)
}

// Histogram tracks the statistical distribution of a set of values on each host.
func (t *tracer) Histogram(name string, value float64, rate float64, tagPairs ...string) {
	t.doHistogram(name, value, rate, tagPairs...)
}

// Incr is a Count of 1.
func (t *tracer) Incr(name string, rate float64, tagPairs ...string) {
	t.doIncr(name, rate, tagPairs...)
}

// Timing sends timing information in milliseconds. It is flushed by
// statsd with percentiles, mean and other info. See
// https://github.com/etsy/statsd/blob/master/docs/metric_types.md#timing
func (t *tracer) Timing(name string, value time.Duration, rate float64, tagPairs ...string) {
	t.doTiming(name, value, rate, tagPairs...)
}

func (t *tracer) doCount(name string, value int64, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Count(name, value, tags, rate)
	t.logStatsd("count", name, tags, zap.Int64("value", value), zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:     CountStatType,
			IntValue: value,
			Name:     t.statsdNamePrefix + name,
			Rate:     rate,
			Tags:     append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) doDistribution(name string, value float64, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Distribution(name, value, tags, rate)
	t.logStatsd("distribution", name, tags, zap.Float64("value", value), zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:       DistributionStatType,
			FloatValue: value,
			Name:       t.statsdNamePrefix + name,
			Rate:       rate,
			Tags:       append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) doEvent(title, text string, options EventOptions) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(options.TagPairs)
	priority := statsd.Normal
	if options.IsLowPriority {
		priority = statsd.Low
	}
	t.logStatsd("event", title, tags)
	_ = t.statsd.Event(&statsd.Event{
		Title:          t.statsdNamePrefix + title,
		Text:           text,
		Timestamp:      options.Timestamp,
		Hostname:       options.Hostname,
		AggregationKey: options.AggregationKey,
		Priority:       priority,
		SourceTypeName: options.SourceTypeName,
		AlertType:      statsd.EventAlertType(options.AlertType),
		Tags:           tags,
	})
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type: EventStatType,
			Name: title,
			Tags: tags,
		}
	}
}

func (t *tracer) doGauge(name string, value float64, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Gauge(name, value, tags, rate)
	t.logStatsd("gauge", name, tags, zap.Float64("value", value), zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:       GaugeStatType,
			FloatValue: value,
			Name:       t.statsdNamePrefix + name,
			Rate:       rate,
			Tags:       append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) doHistogram(name string, value float64, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Histogram(name, value, tags, rate)
	t.logStatsd("histogram", name, tags, zap.Float64("value", value), zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:       HistogramStatType,
			FloatValue: value,
			Name:       t.statsdNamePrefix + name,
			Rate:       rate,
			Tags:       append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) doIncr(name string, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Incr(name, tags, rate)
	t.logStatsd("incr", name, tags, zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:     IncrStatType,
			IntValue: 1,
			Name:     t.statsdNamePrefix + name,
			Rate:     rate,
			Tags:     append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) doTiming(name string, value time.Duration, rate float64, tagPairs ...string) {
	if t == nil {
		return
	}
	tags := t.tagsFromPairs(tagPairs)
	// The datadog-go.Statsd implementation is safe on nil.
	_ = t.statsd.Timing(name, value, tags, rate)
	t.logStatsd("timing", name, tags, zap.Duration("value", value), zap.Float64("rate", rate))
	if t.statsdChannel != nil {
		t.statsdChannel <- StatsDMetric{
			Type:          TimingStatType,
			DurationValue: value,
			Name:          t.statsdNamePrefix + name,
			Rate:          rate,
			Tags:          append(tags, t.statsdTags...),
		}
	}
}

func (t *tracer) logStatsd(method string, name string, tags []string, fields ...zap.Field) {
	if !t.logStatsD {
		return
	}
	fields = append([]zap.Field{zap.String("name", name)}, fields...)
	fields = append(fields, zap.Strings("tags", tags))
	t.getLogger(3).Info("statsd:"+method, fields...)
}

func (t *tracer) tagsFromPairs(pairs []string) []string {
	if len(pairs) == 0 {
		return nil
	}
	tags := make([]string, 0, len(pairs)/2)
	i := 0
	for i < len(pairs) {
		if i+1 < len(pairs) {
			tags = append(tags, pairs[i]+":"+pairs[i+1])
		} else {
			t.getLogger(3).Warn("ignored key without a value", zap.String("ignored", pairs[i]))
		}
		i += 2
	}
	return tags
}
