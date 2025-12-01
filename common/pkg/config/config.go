package config

import (
	"fmt"
	"os"
)

func RequireEnv(vars ...string) error {
	missing := []string{}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("variaveis de ambiente obrigatórias ausentes: %v", missing)
	}
	return nil
}
