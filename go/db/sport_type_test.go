package db

import (
	"reflect"
	"testing"
)

func TestSportTypeName(t *testing.T) {
	want := "MLB"
	st := SportType(1)
	sportTypes[st] = sportType{name: want}
	got := st.Name()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestSportTypeURL(t *testing.T) {
	want := "mlb"
	st := SportType(2)
	sportTypes[st] = sportType{url: want}
	got := st.URL()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestSportTypeDisplayOrder(t *testing.T) {
	want := 8
	st := SportType(3)
	sportTypes[st] = sportType{displayOrder: want}
	got := st.DisplayOrder()
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}

func TestSportTypeFromURL(t *testing.T) {
	sportTypeFromURLTests := []struct {
		loadedSportTypes map[SportType]sportType
		url              string
		want             SportType
	}{
		{},
		{
			loadedSportTypes: map[SportType]sportType{
				SportType(2): {url: "here"},
				SportType(8): {url: "somewhere"},
			},
			url:  "somewhere",
			want: SportType(8),
		},
		{
			loadedSportTypes: map[SportType]sportType{
				SportType(3): {url: "*"},
			},
			url:  "anywhere",
			want: SportType(0),
		},
	}
	for i, test := range sportTypeFromURLTests {
		sportTypes = test.loadedSportTypes
		got := SportTypeFromURL(test.url)
		if test.want != got {
			t.Errorf("Test :%v:\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

func TestSportTypes(t *testing.T) {
	want := []SportType{
		SportType(2),
		SportType(3),
		SportType(1),
	}
	sportTypes = map[SportType]sportType{
		SportType(1): {displayOrder: 3},
		SportType(2): {displayOrder: 1},
		SportType(3): {displayOrder: 2},
	}
	got := SportTypes()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted %v, but got %v", want, got)
	}
}
