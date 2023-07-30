package asesprite

import (
	"fmt"
)

func AnimationNotFoundError(name string) error {
	return fmt.Errorf("no animation found with name '%s'", name)
}
