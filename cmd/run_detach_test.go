package cmd

import (
	"reflect"
	"testing"
)

func TestRemoveDetachFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "long flag",
			args: []string{"run", "--app-id", "demo", "--detach", "--", "echo", "--detach"},
			want: []string{"run", "--app-id", "demo", "--", "echo", "--detach"},
		},
		{
			name: "short flag",
			args: []string{"run", "-d", "--app-id", "demo"},
			want: []string{"run", "--app-id", "demo"},
		},
		{
			name: "explicit true",
			args: []string{"run", "--detach=true", "--app-id=demo"},
			want: []string{"run", "--app-id=demo"},
		},
		{
			name: "explicit false",
			args: []string{"run", "--detach=false", "--app-id=demo"},
			want: []string{"run", "--app-id=demo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeDetachFlag(tt.args)
			if err != nil {
				t.Fatalf("removeDetachFlag() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("removeDetachFlag() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRemoveDetachFlagRejectsInvalidValue(t *testing.T) {
	if _, err := removeDetachFlag([]string{"run", "--detach=invalid"}); err == nil {
		t.Fatal("expected invalid --detach value to fail")
	}
}
