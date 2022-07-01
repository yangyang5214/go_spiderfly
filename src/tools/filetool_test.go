package tools

import (
	"fmt"
	"testing"
)

func TestFileBasename(t *testing.T) {
	fmt.Println(FileBasename("/xxx/1.txt"))
	fmt.Println(FileBasename("/xxx/1"))
}
