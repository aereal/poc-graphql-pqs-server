package domain

type Element string

const (
	ElementUnknown Element = "UNKNOWN"
	ElementPyro    Element = "PYRO"
	ElementHydro   Element = "HYDRO"
	ElementCryo    Element = "CRYO"
	ElementElectro Element = "ELECTRO"
	ElementAnemo   Element = "ANEMO"
	ElementGeo     Element = "GEO"
	ElementDendro  Element = "DENDRO"
)

type WeaponKind string

const (
	WeaponKindUnknown  WeaponKind = "UNKNOWN"
	WeaponKindSword    WeaponKind = "SWORD"
	WeaponKindClaymore WeaponKind = "CLAYMORE"
	WeaponKindBow      WeaponKind = "BOW"
	WeaponKindCatalyst WeaponKind = "CATALYST"
	WeaponKindPolearm  WeaponKind = "POLEARM"
)

type Region string

const (
	RegionUnknown   Region = "UNKNOWN"
	RegionMondstadt Region = "MONDSTADT"
	RegionLiyue     Region = "LIYUE"
	RegionInazuma   Region = "INAZUMA"
	RegionSumeru    Region = "SUMERU"
	RegionFontaine  Region = "FONTAINE"
	RegionNatlan    Region = "NATLAN"
	RegionSnezhnaya Region = "SNEZHNAYA"
)

func (r Region) Name() string {
	switch r {
	case RegionMondstadt:
		return "モンド"
	case RegionLiyue:
		return "璃月"
	case RegionInazuma:
		return "稲妻"
	case RegionSumeru:
		return "スメール"
	case RegionFontaine:
		return "フォンテーヌ"
	case RegionNatlan:
		return "ナタ"
	case RegionSnezhnaya:
		return "スネージナヤ"
	default:
		return ""
	}
}

type Character struct {
	ID                 int        `db:"id"`
	Name               string     `db:"name"`
	Rarelity           int        `db:"rarelity"`
	Element            Element    `db:"element"`
	Health             int        `db:"health"`
	Attack             int        `db:"attack"`
	Defence            int        `db:"defence"`
	UniqueAbility      string     `db:"unique_ability"`
	UniqueAbilityScore float64    `db:"unique_ability_score"`
	ElementEnergy      int        `db:"element_energy"`
	Region             Region     `db:"region"`
	WeaponKind         WeaponKind `db:"weapon_kind"`
}
