package sia

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
)

func TestDifficultyToTarget(t *testing.T) {
	diff, _ := strconv.ParseFloat("0.99998474121094105", 64)

	s := string(4294967296)
	fmt.Println(len(s))
	fmt.Printf("%s", s)

	expectedTarget := "0x00000000fffffffffffefffeffff00000001000200020000fffefffcfffbfffd"

	target, err := difficultyToTarget(diff)
	if err != nil {
		t.Error(err)
	}

	if expectedTarget == ("0x" + hex.EncodeToString(target[:])) {
		t.Error("0x"+hex.EncodeToString(target[:]), "returned instead of", expectedTarget)
	}
}
