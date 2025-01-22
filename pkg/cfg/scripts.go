package cfg

type Script struct {
	Main         *Func
	Filepath     string
	FuncsMap     map[string]*Func
	IncludeFiles []string
}

func NewScript(main *Func, filepath string) *Script {
	return &Script{
		Main:     main,
		Filepath: filepath,
		FuncsMap: make(map[string]*Func),
	}
}
func (s *Script) AddFunc(funct *Func) {
	name := funct.GetScopedName()
	s.FuncsMap[name] = funct

}
