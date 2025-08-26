/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/26 16:34
 */

// Package eventbus

package eventbus

import "fmt"

func printPanic() {
	if r := recover(); r != nil {
		fmt.Printf("executor panic recovered: %v, restarting...\n", r)
	}
}
