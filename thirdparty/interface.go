package thirdparty

// some of the interfaces need to be in here so house doesn't have a cyclical import going on
type Light interface {
	On() error
	Off() error
	Brightness(int32) error
	Flash() error
}
