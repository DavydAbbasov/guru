package migrationtool

import (
	"fmt"
	"strings"
)

type Action string

const (
	ActionUp      Action = "up"
	ActionDown    Action = "down"
	ActionDownAll Action = "down_all"
	ActionStatus  Action = "status"
)

type DBMigrator struct {
	cmds map[Action]func() error
}

func NewDBMigrator(p MigratorProvider) *DBMigrator {
	return &DBMigrator{
		cmds: map[Action]func() error{
			ActionUp:      p.Up,
			ActionDown:    p.Down,
			ActionDownAll: p.DownAll,
			ActionStatus:  p.Status,
		},
	}
}

func (m *DBMigrator) Execute(action Action) error {
	normalized := Action(strings.ToLower(string(action)))

	if fn, ok := m.cmds[normalized]; ok {
		return fn()
	}

	return fmt.Errorf("unknown action: %s", action)
}
