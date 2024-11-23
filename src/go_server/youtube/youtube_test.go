package youtube

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	dir, _ := os.Getwd()
	splitDir := strings.Split(dir, "/")
	if splitDir[len(splitDir)-1] == "youtube" {
		os.Chdir("..")
	}
}

// Warning, these are integration tests
func TestGetVideoIds(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		t.Fatalf("could not load .env")
	}
	ctx := context.Background()
	vIds, err := GetVideoIds(ctx, "UCoOG58fKWhGusmEgAnfhOaw")
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := []string{"REi089fakFI", "vu5ODMuAR5c"}

	if !reflect.DeepEqual(vIds, expected) {
		t.Fatalf(`got: %s, expected %s`, vIds, expected)
	}
}

func TestGetVideo(t *testing.T) {
	vid, err := GetVideo("REi089fakFI")
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "Quick Ableton Beat"

	if vid.Title != expected {
		t.Fatalf(`got: %s, expected %s`, vid.Title, expected)
	}
}

func TestGetVideoTranscripts(t *testing.T) {
	videoIds := []string{"REi089fakFI"}
	tMap := GetVideoTranscripts(videoIds)
	value := strings.TrimSpace(tMap[videoIds[0]])
	expected := "First Characters"

	if value != expected {
		t.Fatalf(`got: %s, expected %s`, value, expected)
	}
}
