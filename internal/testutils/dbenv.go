package testutils

import "os"

func DBEnv() (values map[string]string, ok bool) {
	values = make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if !found {
			return values, false
		}
		values[envVar] = value
	}

	return values, true
}
