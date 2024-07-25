package sqlite

func (d *SqliteDB) Version() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.version
}

// SetVersion
func (d *SqliteDB) SetVersion(value string) {
	d.mu.Lock()
	d.version = value
	d.mu.Unlock()
}
