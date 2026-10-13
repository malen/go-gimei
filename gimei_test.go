package gimei_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/mattn/go-gimei"
)

func TestDeterministicRandom(t *testing.T) {
	prev := collectNewResults(gimei.NewWithSeed(42))
	curr := collectNewResults(gimei.NewWithSeed(42))

	for i := range curr {
		if prev[i].String() != curr[i].String() {
			t.Errorf("curr[%d] == %q, want %q", i, curr[i], prev[i])
		}
	}
}

func TestNewWithSeedUsesGofakeitPCGConvention(t *testing.T) {
	const seed = uint64(42)
	want := collectNewResults(gimei.New(rand.New(rand.NewPCG(seed, seed))))
	got := collectNewResults(gimei.NewWithSeed(seed))

	for i := range got {
		if got[i].String() != want[i].String() {
			t.Errorf("got[%d] == %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNewRejectsNilRandom(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("New(nil) did not panic")
		}
	}()
	gimei.New(nil)
}

func TestGeneratedItemsDoNotAliasSharedData(t *testing.T) {
	const seed = uint64(42)
	name := gimei.NewWithSeed(seed).NewName()
	want := name.Kanji()
	name.First[0] = "modified"

	if got := gimei.NewWithSeed(seed).NewName().Kanji(); got != want {
		t.Fatalf("generated item changed shared data: got %q, want %q", got, want)
	}
}

// returns slice of fmt.Stringer which return value of gimei 'New' functions
func collectNewResults(g *gimei.Gimei) []fmt.Stringer {
	var s []fmt.Stringer

	s = append(s, g.NewName())
	s = append(s, g.NewMaleName())
	s = append(s, g.NewFemaleName())
	s = append(s, g.NewAddress())
	s = append(s, g.NewPrefecture())
	s = append(s, g.NewCity())
	s = append(s, g.NewTown())

	return s
}

func TestFindName(t *testing.T) {
	var target string
	name := gimei.NewWithSeed(1).NewName()

	target = name.Kanji()
	if gimei.FindNameByKanji(target) == nil {
		t.Errorf("FindNameByKanji not found: %s", target)
	}
	target = name.Hiragana()
	if gimei.FindNameByHiragana(target) == nil {
		t.Errorf("FindNameByHiragana not found: %s", target)
	}
	target = name.Katakana()
	if gimei.FindNameByKatakana(target) == nil {
		t.Errorf("FindNameByKatakana not found: %s", target)
	}
	target = name.Romaji()
	if gimei.FindNameByRomaji(target) == nil {
		t.Errorf("FindNameByRomaji not found: %s", target)
	}
}

func TestFindNameInvalidParams(t *testing.T) {
	target := "oneword"

	if gimei.FindNameByRomaji(target) != nil {
		t.Errorf("FindNameByRomaji should return nil: %s", target)
	}
}

func TestFindAddress(t *testing.T) {
	var target string
	addr := gimei.NewWithSeed(1).NewAddress()

	target = addr.Kanji()
	if gimei.FindAddressByKanji(target) == nil {
		t.Errorf("FindAddressByKanji not found: %s", target)
	}
	target = addr.Hiragana()
	if gimei.FindAddressByHiragana(target) == nil {
		t.Errorf("FindAddressByHiragana not found: %s", target)
	}
	target = addr.Katakana()
	if gimei.FindAddressByKatakana(target) == nil {
		t.Errorf("FindAddressByKatakana not found: %q", target)
	}
}

// Prefecture/City/Town.Romaji() should return empty string
func TestEmptyRomaji(t *testing.T) {
	g := gimei.NewWithSeed(1)
	if g.NewPrefecture().Romaji() != "" {
		t.Errorf("Prefecture.Romaji() should return empty string")
	}
	if g.NewCity().Romaji() != "" {
		t.Errorf("City.Romaji() should return empty string")
	}
	if g.NewTown().Romaji() != "" {
		t.Errorf("Town.Romaji() should return empty string")
	}
}

func TestCheckRaceCondition(t *testing.T) {
	g := gimei.NewWithSeed(1)
	for i := range 3 {
		t.Run(fmt.Sprintf("TestCheckRaceCondition: %v", i), func(t *testing.T) {
			t.Parallel()
			g.NewName()
			g.NewDogName()
			g.NewCatName()
			g.NewMaleName()
			g.NewFemaleName()
			g.NewMaleDogName()
			g.NewFemaleDogName()
			g.NewMaleCatName()
			g.NewFemaleCatName()
			g.NewAddress()
			g.NewPrefecture()
			g.NewTown()
			g.NewCity()
			g.NewPostalCode()
			gimei.FindNameByKanji("小林 顕士")
			gimei.FindAddressByKanji("岡山県大島郡大和村稲木町")
			gimei.CountData()
		})
	}
}

func TestPostalCode(t *testing.T) {
	postal := gimei.NewWithSeed(42).NewPostalCode()
	if postal == nil {
		t.Fatal("NewPostalCode should not return nil")
	}

	if postal.Kanji() == "" {
		t.Fatal("PostalCode.Kanji() should not return empty string")
	}

	if postal.String() == "" {
		t.Fatal("PostalCode.String() should not return empty string")
	}
}
