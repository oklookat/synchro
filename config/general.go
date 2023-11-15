package config

type General struct {
	Debug bool
}

func (c General) Default() any {
	return General{
		Debug: true,
	}
}

func (c General) Key() Key {
	return KeyGeneral
}
