package gimei

import (
	cryptorand "crypto/rand"
	"embed"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed data/addresses.yml data/names.yml data/postalcodes.yml
	assets embed.FS

	names       name
	addresses   address
	postalCodes postalCode
	onceName    sync.Once
	onceAddress sync.Once
	oncePostal  sync.Once

	lastNameIndex        [4]map[string]Item
	maleFirstNameIndex   [4]map[string]Item
	femaleFirstNameIndex [4]map[string]Item
	cityIndex            [3]map[string]Item
	townIndex            [3]map[string]Item
)

// Item take four figure for japanese. Kanji/Hiragana/Katakana/Romaji.
// It's not only the difference of sounds, it just letters.
type Item []string

func cloneItem(item Item) Item {
	return append(Item(nil), item...)
}

// String implement Stringer
func (i Item) String() string {
	return i.Kanji()
}

// Kanji return string of Item as kanji.
func (i Item) Kanji() string {
	return i[0]
}

// Hiragana return string of Item as hiragana.
func (i Item) Hiragana() string {
	return i[1]
}

// Katakana return string of Item as katakana.
func (i Item) Katakana() string {
	return i[2]
}

// Romaji return string of Item as romaji.
func (i Item) Romaji() string {
	if len(i) <= 3 {
		return ""
	}
	return cases.Title(language.Und, cases.NoLower).String(i[3])
}

// Sex store Male or Female.
type Sex int

// String implement Stringer.
func (s Sex) String() string {
	switch s {
	case Male:
		return "男"
	case Female:
		return "女"
	}
	return "？"
}

// list of sex
const (
	Male   Sex = iota + 1 // 男
	Female                // 女
)

// name store data sturecture just same as names.yml.
type name struct {
	FirstName struct {
		Male   []Item `yaml:"male"`
		Female []Item `yaml:"female"`
		Animal []Item `yaml:"animal"`
	} `yaml:"first_name"`
	LastName    []Item `yaml:"last_name"`
	LastNameDog []Item `yaml:"last_name_dog"`
	LastNameCat []Item `yaml:"last_name_cat"`
}

// Name store name and sex for a person.
type Name struct {
	First Item
	Last  Item
	Sex   Sex
}

// Gimei generates names, addresses, and postal codes using an instance-local
// random number generator. A Gimei is safe for concurrent use, but must not be
// copied after first use.
type Gimei struct {
	mu sync.Mutex
	r  *rand.Rand
}

// New returns a Gimei that uses rnd as its random number generator.
// rnd must not be nil and must not be used outside the returned Gimei.
func New(rnd *rand.Rand) *Gimei {
	if rnd == nil {
		panic("gimei: nil random number generator")
	}
	return &Gimei{r: rnd}
}

// NewWithSeed returns a Gimei backed by PCG, using the same seed convention as
// gofakeit: a non-zero seed initializes PCG(seed, seed), while zero requests a
// cryptographically random seed.
func NewWithSeed(seed uint64) *Gimei {
	if seed == 0 {
		var seeds [16]byte
		if _, err := cryptorand.Read(seeds[:]); err != nil {
			panic(fmt.Errorf("gimei: generate random seed: %w", err))
		}
		return New(rand.New(rand.NewPCG(
			binary.LittleEndian.Uint64(seeds[:8]),
			binary.LittleEndian.Uint64(seeds[8:]),
		)))
	}
	return New(rand.New(rand.NewPCG(seed, seed)))
}

func loadNames() {
	if b, err := assets.ReadFile("data/names.yml"); err == nil {
		if err = yaml.Unmarshal(b, &names); err == nil {
			buildNameIndex()
			return
		}
	}
	panic("failed to load names data")
}

func buildNameIndex() {
	for i := range 4 {
		lastNameIndex[i] = make(map[string]Item, len(names.LastName))
		for _, item := range names.LastName {
			if i < len(item) {
				key := item[i]
				if i == 3 {
					key = strings.ToLower(key)
				}
				if _, ok := lastNameIndex[i][key]; !ok {
					lastNameIndex[i][key] = item
				}
			}
		}
		maleFirstNameIndex[i] = make(map[string]Item, len(names.FirstName.Male))
		for _, item := range names.FirstName.Male {
			if i < len(item) {
				key := item[i]
				if i == 3 {
					key = strings.ToLower(key)
				}
				if _, ok := maleFirstNameIndex[i][key]; !ok {
					maleFirstNameIndex[i][key] = item
				}
			}
		}
		femaleFirstNameIndex[i] = make(map[string]Item, len(names.FirstName.Female))
		for _, item := range names.FirstName.Female {
			if i < len(item) {
				key := item[i]
				if i == 3 {
					key = strings.ToLower(key)
				}
				if _, ok := femaleFirstNameIndex[i][key]; !ok {
					femaleFirstNameIndex[i][key] = item
				}
			}
		}
	}
}

// String implement Stringer.
func (n *Name) String() string {
	return n.Kanji()
}

// Kanji return string of Name as kanji.
func (n *Name) Kanji() string {
	return n.Last.Kanji() + " " + n.First.Kanji()
}

// Hiragana return string of Name as hiragana.
func (n *Name) Hiragana() string {
	return n.Last.Hiragana() + " " + n.First.Hiragana()
}

// Katakana return string of Name as katakana.
func (n *Name) Katakana() string {
	return n.Last.Katakana() + " " + n.First.Katakana()
}

// Romaji return string of Name as romaji.
func (n *Name) Romaji() string {
	return n.First.Romaji() + " " + n.Last.Romaji()
}

// IsMale return true if he is male.
func (n *Name) IsMale() bool {
	return n.Sex == Male
}

// IsFemale return true if she is female.
func (n *Name) IsFemale() bool {
	return n.Sex == Female
}

// NewName returns a new person.
func (g *Gimei) NewName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.newName()
}

func (g *Gimei) newName() *Name {
	onceName.Do(loadNames)
	if g.r.IntN(2) == 0 {
		return &Name{
			First: cloneItem(names.FirstName.Male[g.r.IntN(len(names.FirstName.Male))]),
			Last:  cloneItem(names.LastName[g.r.IntN(len(names.LastName))]),
			Sex:   Male,
		}
	}
	return &Name{
		First: cloneItem(names.FirstName.Female[g.r.IntN(len(names.FirstName.Female))]),
		Last:  cloneItem(names.LastName[g.r.IntN(len(names.LastName))]),
		Sex:   Female,
	}
}

// NewDogName returns a new animal name whose last name begins "inu".
func (g *Gimei) NewDogName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Animal[g.r.IntN(len(names.FirstName.Animal))]),
		Last:  cloneItem(names.LastNameDog[g.r.IntN(len(names.LastNameDog))]),
	}
}

// NewCatName returns a new animal name whose last name begins "neko".
func (g *Gimei) NewCatName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Animal[g.r.IntN(len(names.FirstName.Animal))]),
		Last:  cloneItem(names.LastNameCat[g.r.IntN(len(names.LastNameCat))]),
	}
}

// NewMaleName returns a new male person.
func (g *Gimei) NewMaleName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Male[g.r.IntN(len(names.FirstName.Male))]),
		Last:  cloneItem(names.LastName[g.r.IntN(len(names.LastName))]),
		Sex:   Male,
	}
}

// NewFemaleName returns a new female person.
func (g *Gimei) NewFemaleName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Female[g.r.IntN(len(names.FirstName.Female))]),
		Last:  cloneItem(names.LastName[g.r.IntN(len(names.LastName))]),
		Sex:   Female,
	}
}

// NewMaleDogName returns a new male person whose last name begins "inu".
func (g *Gimei) NewMaleDogName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Male[g.r.IntN(len(names.FirstName.Male))]),
		Last:  cloneItem(names.LastNameDog[g.r.IntN(len(names.LastNameDog))]),
		Sex:   Male,
	}
}

// NewFemaleDogName returns a new female person whose last name begins "inu".
func (g *Gimei) NewFemaleDogName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Female[g.r.IntN(len(names.FirstName.Female))]),
		Last:  cloneItem(names.LastNameDog[g.r.IntN(len(names.LastNameDog))]),
		Sex:   Female,
	}
}

// NewMaleCatName returns a new male person whose last name begins "neko".
func (g *Gimei) NewMaleCatName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Male[g.r.IntN(len(names.FirstName.Male))]),
		Last:  cloneItem(names.LastNameCat[g.r.IntN(len(names.LastNameCat))]),
		Sex:   Male,
	}
}

// NewFemaleCatName returns a new female person whose last name begins "neko".
func (g *Gimei) NewFemaleCatName() *Name {
	g.mu.Lock()
	defer g.mu.Unlock()

	onceName.Do(loadNames)
	return &Name{
		First: cloneItem(names.FirstName.Female[g.r.IntN(len(names.FirstName.Female))]),
		Last:  cloneItem(names.LastNameCat[g.r.IntN(len(names.LastNameCat))]),
		Sex:   Female,
	}
}

func findNameByIndex(n string, i int) *Name {
	onceName.Do(loadNames)
	token := strings.SplitN(n, " ", 2)
	if len(token) != 2 {
		return nil
	}
	lastKey, firstKey := token[0], token[1]
	if i == 3 { // by romaji
		lastKey, firstKey = strings.ToLower(token[1]), strings.ToLower(token[0])
	}
	last, ok := lastNameIndex[i][lastKey]
	if !ok {
		return nil
	}
	if first, ok := maleFirstNameIndex[i][firstKey]; ok {
		return &Name{First: cloneItem(first), Last: cloneItem(last), Sex: Male}
	}
	if first, ok := femaleFirstNameIndex[i][firstKey]; ok {
		return &Name{First: cloneItem(first), Last: cloneItem(last), Sex: Female}
	}
	return nil
}

// FindNameByKanji find Name by kanji.
func FindNameByKanji(kanji string) *Name {
	return findNameByIndex(kanji, 0)
}

// FindNameByHiragana find Name by hiragana.
func FindNameByHiragana(hiragana string) *Name {
	return findNameByIndex(hiragana, 1)
}

// FindNameByKatakana find Name by katakana.
func FindNameByKatakana(katakana string) *Name {
	return findNameByIndex(katakana, 2)
}

// FindNameByRomaji find Name by romaji.
func FindNameByRomaji(romaji string) *Name {
	return findNameByIndex(romaji, 3)
}

// address store data sturecture just same as addresses.yml.
type address struct {
	Addresses struct {
		Prefecture []Item `yaml:"prefecture"`
		City       []Item `yaml:"city"`
		Town       []Item `yaml:"town"`
	} `yaml:"addresses"`
}

// postalCode store data structure for postal codes
type postalCode struct {
	PostalCodes []Item `yaml:"postal_codes"`
}

// Address store address that is pointed by prefecture/city/town.
type Address struct {
	Prefecture Item
	City       Item
	Town       Item
}

func loadAddresses() {
	if b, err := assets.ReadFile("data/addresses.yml"); err == nil {
		if err = yaml.Unmarshal(b, &addresses); err == nil {
			buildAddressIndex()
			return
		}
	}
	panic("failed to load addresses data")
}

func buildAddressIndex() {
	for i := range 3 {
		cityIndex[i] = make(map[string]Item, len(addresses.Addresses.City))
		for _, item := range addresses.Addresses.City {
			if _, ok := cityIndex[i][item[i]]; !ok {
				cityIndex[i][item[i]] = item
			}
		}
		townIndex[i] = make(map[string]Item, len(addresses.Addresses.Town))
		for _, item := range addresses.Addresses.Town {
			if _, ok := townIndex[i][item[i]]; !ok {
				townIndex[i][item[i]] = item
			}
		}
	}
}

func loadPostalCodes() {
	if b, err := assets.ReadFile("data/postalcodes.yml"); err == nil {
		if err = yaml.Unmarshal(b, &postalCodes); err == nil {
			return
		}
	}
	panic("failed to load postal codes data")
}

// String implement Stringer.
func (a *Address) String() string {
	return a.Kanji()
}

// Kanji return string of Address as kanji.
func (a *Address) Kanji() string {
	return a.Prefecture.Kanji() + a.City.Kanji() + a.Town.Kanji()
}

// Hiragana return string of Address as hiragana.
func (a *Address) Hiragana() string {
	return a.Prefecture.Hiragana() + a.City.Hiragana() + a.Town.Hiragana()
}

// Katakana return string of Address as katakana.
func (a *Address) Katakana() string {
	return a.Prefecture.Katakana() + a.City.Katakana() + a.Town.Katakana()
}

// NewAddress returns a new address.
func (g *Gimei) NewAddress() *Address {
	g.mu.Lock()
	defer g.mu.Unlock()

	return &Address{
		Prefecture: g.newPrefecture(),
		City:       g.newCity(),
		Town:       g.newTown(),
	}
}

// NewPrefecture returns a new prefecture.
func (g *Gimei) NewPrefecture() Item {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.newPrefecture()
}

func (g *Gimei) newPrefecture() Item {
	onceAddress.Do(loadAddresses)
	return cloneItem(addresses.Addresses.Prefecture[g.r.IntN(len(addresses.Addresses.Prefecture))])
}

// NewTown returns a new town.
func (g *Gimei) NewTown() Item {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.newTown()
}

func (g *Gimei) newTown() Item {
	onceAddress.Do(loadAddresses)
	return cloneItem(addresses.Addresses.Town[g.r.IntN(len(addresses.Addresses.Town))])
}

// NewCity returns a new city.
func (g *Gimei) NewCity() Item {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.newCity()
}

func (g *Gimei) newCity() Item {
	onceAddress.Do(loadAddresses)
	return cloneItem(addresses.Addresses.City[g.r.IntN(len(addresses.Addresses.City))])
}

func findAddressByIndex(a string, i int) *Address {
	onceAddress.Do(loadAddresses)
	for _, prefecture := range addresses.Addresses.Prefecture {
		pref := prefecture[i]
		if !strings.HasPrefix(a, pref) {
			continue
		}
		rest := a[len(pref):]
		for k := 1; k < len(rest); k++ {
			city, okC := cityIndex[i][rest[:k]]
			if !okC {
				continue
			}
			town, okT := townIndex[i][rest[k:]]
			if !okT {
				continue
			}
			return &Address{
				Prefecture: cloneItem(prefecture),
				City:       cloneItem(city),
				Town:       cloneItem(town),
			}
		}
	}
	return nil
}

// FindAddressByKanji find Address by kanji.
func FindAddressByKanji(kanji string) *Address {
	return findAddressByIndex(kanji, 0)
}

// FindAddressByHiragana find Address by hiragana.
func FindAddressByHiragana(hiragana string) *Address {
	return findAddressByIndex(hiragana, 1)
}

// FindAddressByKatakana find Address by katakana.
func FindAddressByKatakana(katakana string) *Address {
	return findAddressByIndex(katakana, 2)
}

// PostalCode store postal code
type PostalCode struct {
	Code Item
}

// String implement Stringer.
func (p *PostalCode) String() string {
	return p.Kanji()
}

// Kanji return string of PostalCode as kanji.
func (p *PostalCode) Kanji() string {
	return p.Code.Kanji()
}

// NewPostalCode returns a new postal code.
func (g *Gimei) NewPostalCode() *PostalCode {
	g.mu.Lock()
	defer g.mu.Unlock()

	oncePostal.Do(loadPostalCodes)
	return &PostalCode{
		Code: cloneItem(postalCodes.PostalCodes[g.r.IntN(len(postalCodes.PostalCodes))]),
	}
}

func CountData() string {
	onceName.Do(loadNames)
	onceAddress.Do(loadAddresses)
	oncePostal.Do(loadPostalCodes)

	var addr = &addresses.Addresses
	return fmt.Sprintf(`FirstName:
  Male:       %5d 
  Female:     %5d
LastName:     %5d
Addresses:
  Prefecture: %5d
  City:       %5d
  Town:       %5d
PostalCodes: %5d`,
		len(names.FirstName.Male), len(names.FirstName.Female), len(names.LastName),
		len(addr.Prefecture), len(addr.City), len(addr.Town),
		len(postalCodes.PostalCodes))
}
