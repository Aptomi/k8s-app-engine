package db

// list of participating values for specific index
type Index func(name string, value Storable) ([]interface{}, error)
