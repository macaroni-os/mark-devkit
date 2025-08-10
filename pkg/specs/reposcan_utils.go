/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"fmt"
	"strings"
)

func SanitizeCategory(cat string, slot string) string {
	ans := cat
	if slot != "0" {
		// Ignore sub-slot
		if strings.Contains(slot, "/") {
			slot = slot[0:strings.Index(slot, "/")]
		}

		if slot != "0" && slot != "" {
			ans = fmt.Sprintf("%s-%s", cat, slot)
		}
	}
	return ans
}
