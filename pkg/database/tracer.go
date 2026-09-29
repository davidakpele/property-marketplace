package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type queryTracer struct {
	log *logrus.Logger
}

type traceKey struct{}

type traceData struct {
	sql    string
	args   []interface{}
	start  time.Time
}

func newQueryTracer(log *logrus.Logger) *queryTracer {
	return &queryTracer{log: log}
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	args := make([]interface{}, len(data.Args))
	copy(args, data.Args)
	return context.WithValue(ctx, traceKey{}, traceData{
		sql:   data.SQL,
		args:  args,
		start: time.Now(),
	})
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	td, ok := ctx.Value(traceKey{}).(traceData)
	if !ok {
		return
	}
	dur := time.Since(td.start)
	entry := t.log.WithFields(logrus.Fields{
		"sql":      td.sql,
		"args":     td.args,
		"duration": dur.String(),
		"rows":     data.CommandTag.RowsAffected(),
	})
	if data.Err != nil {
		entry.WithField("error", data.Err.Error()).Error("query failed")
	} else {
		entry.Debug("query executed")
	}
}
