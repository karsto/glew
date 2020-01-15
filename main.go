package main

import "fmt"

func main() {
	fmt.Println("test")
	ctx := map[string]interface{}{}

	fmt.Println(ctx)
}

// Must pass pre cased for casing functions to detect words
// returns various casing formats of the string passed
func getAltCasing(modelName string) {
	// lower upper single plural
}

// returns args := []interface{}{ m.Attributes .. }
func getUpdateAttributes(model interface{}) {
}

// returns read, list, insert, update, delete boiler
func getSql(model interface{}) {
}
