package secrets

import (
	"encoding/json"
	"testing"
)

func TestResultToINI(t *testing.T) {
	t.Run("golden pgbouncer output", func(t *testing.T) {
		var result Result
		err := json.Unmarshal([]byte(`{
  "databases": { "template1": "host=localhost dbname=template1 auth_user=someuser" },
  "pgbouncer": {
    "pool_mode": "session", "listen_port": "6432", "listen_addr": "localhost",
    "auth_type": "md5", "auth_file": "users.txt", "logfile": "pgbouncer.log",
    "pidfile": "pgbouncer.pid", "admin_users": "someuser", "stats_users": "stat_collector"
  }
}`), &result)
		if err != nil {
			t.Fatalf("unable to unmarshal fixture: %v", err)
		}

		expected := `[databases]
template1 = host=localhost dbname=template1 auth_user=someuser

[pgbouncer]
admin_users = someuser
auth_file = users.txt
auth_type = md5
listen_addr = localhost
listen_port = 6432
logfile = pgbouncer.log
pidfile = pgbouncer.pid
pool_mode = session
stats_users = stat_collector
`

		if got := result.ToINI(); got != expected {
			t.Errorf("unexpected INI output\n--- got ---\n%q\n--- want ---\n%q", got, expected)
		}
	})

	t.Run("top level scalars go to default", func(t *testing.T) {
		result := Result{"FOO": "bar", "BAZ": "qux"}
		expected := "[default]\nBAZ = qux\nFOO = bar\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("deterministic ordering", func(t *testing.T) {
		result := Result{
			"zeta":  map[string]any{"b": "2", "a": "1"},
			"alpha": map[string]any{"y": "2", "x": "1"},
		}
		first := result.ToINI()
		for i := 0; i < 20; i++ {
			if got := result.ToINI(); got != first {
				t.Fatalf("output not deterministic:\n%q\n%q", first, got)
			}
		}
	})

	t.Run("keys and sections sanitised", func(t *testing.T) {
		result := Result{
			"db.user name": map[string]any{"user name": "someuser"},
			"plain":        "value",
		}
		expected := "[db_user_name]\nuser_name = someuser\n\n[default]\nplain = value\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("numeric and bool values raw", func(t *testing.T) {
		var result Result
		if err := json.Unmarshal([]byte(`{"port": 6432, "enabled": true, "ratio": 1.5}`), &result); err != nil {
			t.Fatalf("unable to unmarshal fixture: %v", err)
		}
		expected := "[default]\nenabled = true\nport = 6432\nratio = 1.5\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("nil renders null", func(t *testing.T) {
		result := Result{"missing": nil}
		expected := "[default]\nmissing = null\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("newlines folded and equals kept intact", func(t *testing.T) {
		result := Result{"multi": "line1\nline2\rline3", "conn": "host=x dbname=y auth_user=z"}
		expected := "[default]\nconn = host=x dbname=y auth_user=z\nmulti = line1\\nline2\\rline3\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		if got := (Result{}).ToINI(); got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})

	t.Run("identical sanitised keys still merge", func(t *testing.T) {
		result := Result{"db.prod": map[string]any{"a.b": "1", "a-b": "1"}}
		expected := "[db_prod]\na_b = 1\n"
		if got := result.ToINI(); got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})
}
