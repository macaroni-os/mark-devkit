/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package helpers

import (
	"fmt"
)

func SanitizeMapVersionsField(atomName string,
	vars *map[string]interface{}) error {

	// I need to convert []interface{} to []string
	vlist := []string{}

	for k, v := range *vars {
		if k == "versions" {

			ilist, ok := v.([]interface{})
			if !ok {
				return fmt.Errorf(
					"Invalid type on versions var for package %s",
					atomName)
			}

			for _, vv := range ilist {
				str, ok := vv.(string)
				if !ok {
					return fmt.Errorf(
						"Invalid value %v on versions var for package %s",
						vv, atomName)
				}
				vlist = append(vlist, str)
			}
		}
	}

	if len(vlist) > 0 {
		values := *vars
		values["versions"] = vlist
	}

	return nil
}

func GetSlotFromValues(vars *map[string]interface{}) string {
	values := *vars
	slot, ok := values["slot"].(string)
	if !ok {
		islot, valid := values["slot"].(int)
		if valid {
			slot = fmt.Sprintf("%d", islot)
		}
	}
	return slot
}
