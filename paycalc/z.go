package paycalc

func init() {
	if err := registerTables(); err != nil {
		panic(err)
	}
}
