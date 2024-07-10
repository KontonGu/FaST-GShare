package main

import (
	"pkg/fast-configurator/fastconfigurator"
)

func main() {
	// bm := bitmap.NewBitmap(20)
	// fmt.Println(bm.Get(2))
	fastconfigurator.Run("new")
}
