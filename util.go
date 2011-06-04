// Copyright (c) Paulo Suzart. All rights reserved.
// The use and distribution terms for this software are covered by the
// Eclipse Public License 1.0 (http://opensource.org/licenses/eclipse-1.0.php)
// which can be found in the file epl-v10.html at the root of this distribution.
// By using this software in any fashion, you are agreeing to be bound by
// the terms of this license.
// You must not remove this notice, or any other, from this software.

package main

import (
	"os"
	"strings"
	"time"
)

//Return the min or y if x is -1
func Min(x, y int64) int64 {
	if x == -1 {
		return y
	}

	if x < y {
		return x
	}
	return y

}

//Return the max
func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y

}
//Masters implements this interface.
//If a timeout occour, the gb will call the
//shutdown function.
type Supervised interface {
	Shutdown()
}
//Used by template to generate gb output        
type StringWritter struct {
	s string
}

//Writes the template as string
func (self *StringWritter) Write(p []byte) (n int, err os.Error) {
	self.s += string(p)
	return len(self.s), nil
}

//Parse any flag represented by
//a key-value with a separator.
func parseKV(param *string, separator, errmsg string) (k, v string, err os.Error) {
	if *param == "" {
		return
	}
	data := strings.Split(*param, separator, 2)

	if len(data) != 2 {
		err = os.NewError(errmsg)
	}
	k = data[0]
	v = data[1]
	return
}

func counting(f func()) int64{
	start := time.Nanoseconds()
	f()
	return time.Nanoseconds() - start
}
