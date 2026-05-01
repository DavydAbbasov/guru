package migrationtool

type MigratorProvider interface {
	Up() error
	Down() error
	DownAll() error
	Status() error
}
