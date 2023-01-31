//go:build !portable && !embedded && (!js || forcefat)

package cardrank

import (
	_ "embed"
)

// Embedded lookup table.
var (
	//go:embed twoplustwo00.dat
	twoplustwo00Dat []byte
	//go:embed twoplustwo01.dat
	twoplustwo01Dat []byte
	//go:embed twoplustwo02.dat
	twoplustwo02Dat []byte
	//go:embed twoplustwo03.dat
	twoplustwo03Dat []byte
	//go:embed twoplustwo04.dat
	twoplustwo04Dat []byte
	//go:embed twoplustwo05.dat
	twoplustwo05Dat []byte
	//go:embed twoplustwo06.dat
	twoplustwo06Dat []byte
	//go:embed twoplustwo07.dat
	twoplustwo07Dat []byte
	//go:embed twoplustwo08.dat
	twoplustwo08Dat []byte
	//go:embed twoplustwo09.dat
	twoplustwo09Dat []byte
	//go:embed twoplustwo10.dat
	twoplustwo10Dat []byte
	//go:embed twoplustwo11.dat
	twoplustwo11Dat []byte
	//go:embed twoplustwo12.dat
	twoplustwo12Dat []byte
)
