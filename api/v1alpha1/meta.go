package v1alpha1

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type meta struct{ string }

func (m meta) String() string { return m.string }

var (
	// MetaID is a unique ID for the message.
	// Required.
	// https://github.com/cloudevents/spec/blob/master/spec.md#id
	MetaID = meta{"id"}
	// MetaSource is the source of the messages as a Unique Resource Identifier (URI).
	// Required.
	// https://github.com/cloudevents/spec/blob/master/spec.md#source-1
	MetaSource = meta{"source"}
	// MetaTime is the time of the message. As meta-data, this might be different to the event-time (which might be within the message).
	// For example, it might be the last-modified time of a file, but the file itself was created at another time.
	// Optional.
	// https://github.com/cloudevents/spec/blob/master/spec.md#time
	MetaTime = meta{"time"}
)

func ContextWithMeta(ctx context.Context, source, id string, time time.Time) context.Context {
	return context.WithValue(
		context.WithValue(
			context.WithValue(
				ctx,
				MetaSource,
				source,
			),
			MetaID,
			id,
		),
		MetaTime,
		time,
	)
}

func MetaFromContext(ctx context.Context) (source, id string, t time.Time) {
	return ctx.Value(MetaSource).(string), ctx.Value(MetaID).(string), ctx.Value(MetaTime).(time.Time)
}

func MetaInject(ctx context.Context, h interface{}) {
	source, id, t := MetaFromContext(ctx)
	switch v := h.(type) {
	case http.Header:
		v.Add(MetaSource.String(), source)
		v.Add(MetaID.String(), id)
		v.Add(MetaTime.String(), t.Format(time.RFC3339))
	default:
		panic(fmt.Errorf("unexpected type %T", h))
	}
}

func MetaExtract(ctx context.Context, h interface{}) context.Context {
	switch v := h.(type) {
	case http.Header:
		t, _ := time.Parse(time.RFC3339, v.Get(MetaTime.String()))
		return ContextWithMeta(ctx,
			v.Get(MetaSource.String()),
			v.Get(MetaID.String()),
			t,
		)
	default:
		panic(fmt.Errorf("unexpected type %T", h))
	}
}
