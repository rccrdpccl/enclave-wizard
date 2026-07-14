package runner

import (
	"sort"
	"strings"
)

func ScenarioKey(playbook string, tags []string) string {
	key := strings.ReplaceAll(playbook, "/", "-")
	if len(tags) == 0 {
		return key
	}
	sorted := make([]string, len(tags))
	copy(sorted, tags)
	sort.Strings(sorted)
	return key + "--" + strings.Join(sorted, ",")
}
