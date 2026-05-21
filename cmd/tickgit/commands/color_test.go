package commands

import "testing"

func TestTodosCommandRegistersColorFlag(t *testing.T) {
	flag := todosCmd.Flags().Lookup("color")
	if flag == nil {
		t.Fatal("expected todos command to register --color")
	}
	if flag.DefValue != "auto" {
		t.Fatalf("expected --color default to be auto, got %q", flag.DefValue)
	}
}
