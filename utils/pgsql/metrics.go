package pgsql

import (
	"strings"
	"time"

	"guru/utils/metrics"

	"gorm.io/gorm"
)

const (
	dbMetricsStartKey   = "guru:pgsql:metrics:start"
	dbMetricsCallbackNS = "guru:metrics:"

	phaseBefore = "before"
	phaseAfter  = "after"

	opCreate = "create"
	opQuery  = "query"
	opUpdate = "update"
	opDelete = "delete"
	opRow    = "row"
	opRaw    = "raw"
)

func cbName(phase, op string) string { return dbMetricsCallbackNS + phase + "_" + op }

func RegisterMetrics(db *gorm.DB, m *metrics.Metrics) error {
	before := func(d *gorm.DB) {
		d.InstanceSet(dbMetricsStartKey, time.Now())
	}
	after := func(d *gorm.DB) {
		v, ok := d.InstanceGet(dbMetricsStartKey)
		if !ok {
			return
		}
		start, ok := v.(time.Time)
		if !ok {
			return
		}
		op := operationFromSQL(d.Statement.SQL.String())
		table := d.Statement.Table
		if table == "" {
			table = "unknown"
		}
		m.DBQueryDuration.WithLabelValues(op, table).
			Observe(time.Since(start).Seconds())
	}

	cb := db.Callback()
	if err := cb.Create().Before("*").Register(cbName(phaseBefore, opCreate), before); err != nil {
		return err
	}
	if err := cb.Create().After("*").Register(cbName(phaseAfter, opCreate), after); err != nil {
		return err
	}
	if err := cb.Query().Before("*").Register(cbName(phaseBefore, opQuery), before); err != nil {
		return err
	}
	if err := cb.Query().After("*").Register(cbName(phaseAfter, opQuery), after); err != nil {
		return err
	}
	if err := cb.Update().Before("*").Register(cbName(phaseBefore, opUpdate), before); err != nil {
		return err
	}
	if err := cb.Update().After("*").Register(cbName(phaseAfter, opUpdate), after); err != nil {
		return err
	}
	if err := cb.Delete().Before("*").Register(cbName(phaseBefore, opDelete), before); err != nil {
		return err
	}
	if err := cb.Delete().After("*").Register(cbName(phaseAfter, opDelete), after); err != nil {
		return err
	}
	if err := cb.Row().Before("*").Register(cbName(phaseBefore, opRow), before); err != nil {
		return err
	}
	if err := cb.Row().After("*").Register(cbName(phaseAfter, opRow), after); err != nil {
		return err
	}
	if err := cb.Raw().Before("*").Register(cbName(phaseBefore, opRaw), before); err != nil {
		return err
	}
	if err := cb.Raw().After("*").Register(cbName(phaseAfter, opRaw), after); err != nil {
		return err
	}
	return nil
}

func operationFromSQL(sql string) string {
	fields := strings.Fields(sql)
	if len(fields) == 0 {
		return "unknown"
	}
	return strings.ToLower(fields[0])
}
