package chatserver

import (
	"fmt"
	"strconv"
	"strings"
)

func IncrementLamport(lamport int) int {
	return lamport + 1
}

func SyncLamport(curLamport int, newLamport int) int {
	if curLamport < newLamport {
		curLamport = newLamport
	}
	return curLamport
}

func SplitLamport(str string) (int, string) {
	newstr := strings.Split(str, " | ")
	num, err := strconv.Atoi(newstr[1])
	if err != nil {
		fmt.Println(err)
	}
	return num, newstr[0]
}
