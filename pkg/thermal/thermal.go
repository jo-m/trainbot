// Package thermal reads temperatures from the Linux sensors subsystem.
package thermal

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const basePath = "/sys/class/thermal/thermal_zone0"

// MeasureDegC reads a thermal measurement from sysfs.
func MeasureDegC() (float64, error) {
	path := filepath.Join(basePath, "temp")
	contents, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)
	if err != nil {
		return 0, err
	}
	return float64(i) / 1000, nil
}
