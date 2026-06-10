package pagination

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds page/limit query values.
type Params struct {
	Page  int
	Limit int
}

// Options configures Normalize behavior.
type Options struct {
	DefaultLimit int
	MaxLimit     int
}

// Normalize clamps page/limit to sane defaults.
func Normalize(p *Params, opts Options) {
	if opts.DefaultLimit <= 0 {
		opts.DefaultLimit = DefaultLimit
	}
	if opts.MaxLimit <= 0 {
		opts.MaxLimit = MaxLimit
	}
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 || p.Limit > opts.MaxLimit {
		p.Limit = opts.DefaultLimit
	}
}

// Offset returns SQL/list offset for the current page.
func Offset(p Params) int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}
