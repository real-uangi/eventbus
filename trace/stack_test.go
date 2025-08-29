/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/28 12:47
 */

// Package trace
package trace

import (
	"errors"
	"log"
	"testing"
)

func TestStack(t *testing.T) {
	p()
	e()
}

func p() {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(Stack(3))
		}
	}()
	panic(errors.New("this is a panic"))
}

func e() {
	log.Println(Stack(1))
}
