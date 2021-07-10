package generate

type jobstubStructArgs struct {
	Name string
}

type clientArgs struct {
	RootPackageName string
	JobNames        []string
}

type serverArgs struct {
	RootPackageName string
	JobNames        []string
}
