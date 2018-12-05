package gobusterdns

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	ShowIPs        bool
	ShowCNAME      bool
	WildcardForced bool
}

// NewOptionsDNS returns a new initialized OptionsDNS
func NewOptionsDNS() *OptionsDNS {
	return &OptionsDNS{}
}
