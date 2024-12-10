package core

import (
	"bufio"
	"os"
)

func ReadAllLines(processor func(string)) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			panic(err)
		}

		processor(text)
	}
}

func SpawnAllLines(ch chan string) {
	ReadAllLines(func(line string) {
		ch <- line
	})
	close(ch)
}
